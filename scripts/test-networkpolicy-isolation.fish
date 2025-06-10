#!/usr/bin/env fish

function show_usage
    echo "Usage: test-networkpolicy-isolation.fish <namespace>"
    echo "Tests that NetworkPolicy denies outbound internet access from pods in the given namespace"
end

function check_required_binaries
    set required_binaries oc gum
    
    for binary in $required_binaries
        if not command -v $binary >/dev/null 2>&1
            echo "Error: Required binary '$binary' is not installed or not in PATH"
            exit 1
        end
    end
end

# Check for required binaries before proceeding
check_required_binaries

if test (count $argv) -ne 1
    gum style --foreground="#ff0000" "Error: Namespace parameter is required"
    show_usage
    exit 1
end

set namespace $argv[1]
set pod_name "networkpolicy-test-pod"
set test_url "https://www.google.com"

gum style --foreground="#00ff00" --bold "Testing NetworkPolicy isolation in namespace: $namespace"

# Check if namespace exists
if not oc get namespace $namespace >/dev/null 2>&1
    gum style --foreground="#ff0000" "Error: Namespace '$namespace' does not exist"
    exit 1
end

# Create temporary YAML file
set temp_yaml "/tmp/networkpolicy-test-pod.yaml"

gum style --foreground="#0000ff" "Creating Pod YAML..."

echo "apiVersion: v1
kind: Pod
metadata:
  name: $pod_name
  namespace: $namespace
spec:
  restartPolicy: Never
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: $pod_name
    image: quay.io/curl/curl:latest
    command: [\"curl\", \"-s\", \"--connect-timeout\", \"10\", \"--max-time\", \"15\", \"$test_url\"]
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
      runAsNonRoot: true
      seccompProfile:
        type: RuntimeDefault" > $temp_yaml

echo "Pod YAML content:"
cat $temp_yaml

gum style --foreground="#0000ff" "Applying Pod manifest..."
set pod_output (oc apply -f $temp_yaml 2>&1)

if test $status -ne 0
    gum style --foreground="#ff0000" "Error: Failed to create test pod"
    echo "Command that failed:"
    echo "oc apply -f $temp_yaml"
    echo "oc output:"
    echo $pod_output
    rm -f $temp_yaml
    exit 1
end

# Wait for pod to complete (it will exit after curl finishes)
gum style --foreground="#0000ff" "Waiting for pod to complete curl attempt..."
oc wait --for=condition=Ready pod/$pod_name --namespace=$namespace --timeout=60s >/dev/null 2>&1

# Wait a bit more for curl to complete
sleep 10

# Check pod exit code to determine if curl succeeded
set pod_exit_code (oc get pod $pod_name --namespace=$namespace -o jsonpath='{.status.containerStatuses[0].state.terminated.exitCode}' 2>/dev/null)
set curl_result (oc logs $pod_name --namespace=$namespace 2>/dev/null)

# If pod is still running, curl might be hanging (network blocked)
set pod_phase (oc get pod $pod_name --namespace=$namespace -o jsonpath='{.status.phase}' 2>/dev/null)
if test "$pod_phase" = "Running"
    set curl_exit_code 1
    set curl_result "Connection timeout/blocked"
else if test -n "$pod_exit_code"
    set curl_exit_code $pod_exit_code
else
    set curl_exit_code 1
end

# Clean up the pod and temp file
oc delete pod $pod_name --namespace=$namespace >/dev/null 2>&1
rm -f $temp_yaml

if test $curl_exit_code -eq 0
    gum style --foreground="#ff0000" --bold "❌ TEST FAILED: NetworkPolicy is NOT blocking outbound traffic"
    gum style --foreground="#ff0000" "The pod was able to reach $test_url"
    exit 1
else
    gum style --foreground="#00ff00" --bold "✅ TEST PASSED: NetworkPolicy is blocking outbound traffic"
    gum style --foreground="#00ff00" "The pod was unable to reach $test_url (as expected)"
    exit 0
end