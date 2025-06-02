#!/usr/bin/env fish

function show_usage
    echo "Usage: search-ci-failures.fish REPOSITORY JOB PATTERN [MIN_PR]"
    echo ""
    echo "Arguments:"
    echo "  REPOSITORY  GitHub repository in org/repo format"
    echo "  JOB         CI job name to search"
    echo "  PATTERN     Regex pattern to search for in build logs"
    echo "  MIN_PR      Optional minimum PR number (skip earlier PRs)"
    echo ""
    echo "Example:"
    echo "  search-ci-failures.fish openshift/origin periodic-ci-openshift-origin-master-e2e-aws 'test failed'"
    echo "  search-ci-failures.fish openshift/origin periodic-ci-openshift-origin-master-e2e-aws 'test failed' 1000"
    exit 1
end

function check_dependencies
    if not command -v gsutil >/dev/null 2>&1
        gum style --foreground "#FF0000" "Error: gsutil is not installed or not in PATH"
        exit 1
    end

    if not command -v gum >/dev/null 2>&1
        echo "Error: gum is not installed or not in PATH"
        echo "Install it from: https://github.com/charmbracelet/gum"
        exit 1
    end
end

function main
    if test (count $argv) -lt 3 -o (count $argv) -gt 4
        show_usage
    end

    check_dependencies

    set repository $argv[1]
    set job $argv[2]
    set pattern $argv[3]
    set min_pr 0

    if test (count $argv) -eq 4
        set min_pr $argv[4]
        if not string match -qr '^\d+$' $min_pr
            gum style --foreground "#FF0000" "Error: MIN_PR must be a number"
            exit 1
        end
    end

    # Convert org/repo to org_repo format
    set flattened_repo (string replace "/" "_" $repository)

    gum style --foreground "#0000FF" --bold "🔍 Searching for CI failures"
    gum style --foreground "#00FFFF" "Repository: $repository"
    gum style --foreground "#00FFFF" "Job: $job"
    gum style --foreground "#00FFFF" "Pattern: $pattern"
    if test $min_pr -gt 0
        gum style --foreground "#00FFFF" "Minimum PR: $min_pr"
    end
    echo

    set bucket_path "gs://test-platform-results/pr-logs/pull/$flattened_repo"

    # Get list of PRs
    gum style --foreground "#FFFF00" "📋 Fetching PR list..."
    set all_prs (gsutil ls $bucket_path/ 2>/dev/null | string match -r '\d+' | sort -n)

    if test (count $all_prs) -eq 0
        gum style --foreground "#FF0000" "No PRs found for repository $repository"
        exit 1
    end

    # Filter PRs based on minimum PR number
    set prs
    for pr in $all_prs
        if test $pr -ge $min_pr
            set -a prs $pr
        end
    end

    if test (count $prs) -eq 0
        gum style --foreground "#FF0000" "No PRs found >= $min_pr for repository $repository"
        exit 1
    end

    gum style --foreground "#00FF00" "Found "(count $prs)" PRs to process"
    echo

    set matching_jobs
    set total_prs (count $prs)
    set processed_prs 0
    set total_jobs_processed 0
    set start_time (date +%s)

    for pr in $prs
        set processed_prs (math $processed_prs + 1)
        gum style --foreground "#0000FF" "🔄 Processing PR #$pr ($processed_prs/$total_prs)"

        set pr_path "$bucket_path/$pr/$job"
        set job_runs (gsutil ls $pr_path/ 2>/dev/null | string match -r '[^/]+/?$' | string replace -r '/?$' '' | grep -v 'latest-build.txt')

        if test (count $job_runs) -eq 0
            gum style --foreground "#808080" "  No job runs found for PR #$pr"
            continue
        end

        set total_jobs (count $job_runs)
        set processed_jobs 0
        set total_jobs_processed (math $total_jobs_processed + $total_jobs)

        for job_run in $job_runs
            set processed_jobs (math $processed_jobs + 1)
            gum style --foreground "#808080" "  Checking job run $job_run ($processed_jobs/$total_jobs)"

            set job_path "$pr_path/$job_run"

            # Check if started.json exists and get timestamp
            set started_content (gsutil cat "$job_path/started.json" 2>/dev/null)
            if test $status -ne 0
                gum style --foreground "#808080" "    Skipping: no started.json"
                continue
            end

            set timestamp (echo $started_content | jq -r '.timestamp // empty' 2>/dev/null)
            if test -z "$timestamp"
                gum style --foreground "#808080" "    Skipping: no timestamp in started.json"
                continue
            end

            # Check if build-log.txt exists (plain or gzipped) and search for pattern
            set found_match false

            if gsutil ls "$job_path/build-log.txt" >/dev/null 2>&1
                # First try as plain text
                if gsutil cat "$job_path/build-log.txt" 2>/dev/null | /usr/bin/grep -E "$pattern" >/dev/null 2>&1
                    set found_match true
                    # If that fails, try as gzipped content
                else if gsutil cat "$job_path/build-log.txt" 2>/dev/null | gunzip 2>/dev/null | /usr/bin/grep -E "$pattern" >/dev/null 2>&1
                    set found_match true
                end
            else if gsutil ls "$job_path/build-log.txt.gz" >/dev/null 2>&1
                if gsutil cat "$job_path/build-log.txt.gz" 2>/dev/null | gunzip 2>/dev/null | /usr/bin/grep -E "$pattern" >/dev/null 2>&1
                    set found_match true
                end
            else
                gum style --foreground "#808080" "    Skipping: no build-log.txt or build-log.txt.gz"
                continue
            end

            if test $found_match = true
                gum style --foreground "#FFA500" "    ✗ Pattern matched!"
                set -a matching_jobs "$flattened_repo|$pr|$job|$job_run|$timestamp"
            else
                gum style --foreground "#00FF00" "    ✓ No match"
            end
        end

        # Calculate time estimates after processing each PR
        set current_time (date +%s)
        set elapsed_time (math $current_time - $start_time)
        set remaining_prs (math $total_prs - $processed_prs)

        if test $processed_prs -gt 0
            set avg_jobs_per_pr (math "round($total_jobs_processed / $processed_prs)")
            set estimated_remaining_jobs (math "round($remaining_prs * $avg_jobs_per_pr)")
            set avg_time_per_pr (math "$elapsed_time / $processed_prs")
            set estimated_remaining_time (math "round($remaining_prs * $avg_time_per_pr)")

            set elapsed_min (math "round($elapsed_time / 60)")
            set elapsed_sec (math "$elapsed_time % 60")
            set remaining_min (math "round($estimated_remaining_time / 60)")
            set remaining_sec (math "$estimated_remaining_time % 60")

            gum style --foreground "#FF00FF" "  📊 Progress:      $processed_prs/$total_prs PRs, ~$estimated_remaining_jobs jobs remaining (avg: $avg_jobs_per_pr runs/PR)"
            gum style --foreground "#FF00FF" "  ⏱️ Total elapsed: "$elapsed_min"m"$elapsed_sec"s, Est. remaining: "$remaining_min"m"$remaining_sec"s"
        end
        echo
    end

    echo
    gum style --foreground "#0000FF" --bold "📊 Results"

    if test (count $matching_jobs) -eq 0
        gum style --foreground "#00FF00" "No jobs found with the specified pattern"
        exit 0
    end

    gum style --foreground "#FFFF00" "Found "(count $matching_jobs)" jobs with failures:"
    echo

    # Sort by timestamp (reverse order - most recent first)
    set sorted_jobs (printf '%s\n' $matching_jobs | sort -t'|' -k5 -nr)

    for job_info in $sorted_jobs
        set repo_flat (echo $job_info | cut -d'|' -f1)
        set pr (echo $job_info | cut -d'|' -f2)
        set job_name (echo $job_info | cut -d'|' -f3)
        set job_id (echo $job_info | cut -d'|' -f4)
        set ts (echo $job_info | cut -d'|' -f5)

        set url "https://prow.ci.openshift.org/view/gs/test-platform-results/pr-logs/pull/$repo_flat/$pr/$job_name/$job_id"
        set date_str (date -d @$ts '+%Y-%m-%d %H:%M:%S UTC' 2>/dev/null || date -r $ts '+%Y-%m-%d %H:%M:%S UTC' 2>/dev/null || echo "timestamp: $ts")

        gum style --foreground "#00FFFF" "🔗 $url"
        gum style --foreground "#808080" "   Executed: $date_str"
        echo
    end
end

main $argv
