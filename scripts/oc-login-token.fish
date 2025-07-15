#!/usr/bin/env fish

# Enhanced oc-login-token script with gum integration and Kerberos authentication
# Can be used as a standalone script or sourced as a function

# AIDEV-NOTE: This script improves the original oc-login-token function with:
# - Kerberos authentication check and kinit support
# - gum integration for better UI
# - ocp-sso-token integration for automatic token retrieval
# - Context selection with existing contexts only (no new context creation)

# Color definitions for gum styling
set -l blue "#268BD2"
set -l green "#859900"
set -l red "#DC322F"

function oc-login-token --description "Enhanced OpenShift login with token using gum and Kerberos"
    # Default values
    set -l default_krb_user "pmuller"
    set -l default_provider "redhat-sso"
    
    # Parse command line arguments
    argparse 'u/user=' 'p/provider=' 'h/help' -- $argv
    or return 1
    
    # Declare variables at function level first
    set -l krb_user
    set -l provider
    
    # Set defaults after parsing, with command line overrides
    if set -q _flag_user
        set krb_user $_flag_user
    else
        set krb_user $default_krb_user
    end
    
    if set -q _flag_provider
        set provider $_flag_provider
    else
        set provider $default_provider
    end
    
    if set -q _flag_help
        echo "Usage: oc-login-token [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  -u, --user USER      Kerberos username (default: pmuller)"
        echo "  -p, --provider PROVIDER  SSO provider name (default: redhat-sso)"
        echo "  -h, --help           Show this help message"
        return 0
    end
    
    # Check if required tools are available
    if not command -q gum
        echo "Error: gum is required but not installed"
        return 1
    end
    
    if not command -q oc
        echo "Error: oc is required but not installed"
        return 1
    end
    
    if not command -q ocp-sso-token
        echo "Error: ocp-sso-token is required but not installed"
        return 1
    end
    
    # Header
    gum style --foreground=$blue --bold --border rounded --align=center --padding "0 2" --border-foreground=$blue "OpenShift Login Token"
    
    # Check Kerberos ticket
    echo ""
    gum style --foreground=$blue --bold "Checking Kerberos authentication..."
    
    if not klist -s 2>/dev/null
        gum style --foreground=$red "No valid Kerberos ticket found"
        if gum confirm "Run kinit for user $krb_user?"
            kinit $krb_user
            or begin
                gum style --foreground=$red --bold "Failed to obtain Kerberos ticket"
                return 1
            end
        else
            gum style --foreground=$red --bold "Kerberos authentication required"
            return 1
        end
    else
        gum style --foreground=$green "✓ Valid Kerberos ticket found"
    end
    
    # Get available contexts
    echo ""
    gum style --foreground=$blue --bold "Getting available contexts..."
    
    set -l contexts (oc config get-contexts -o name)
    if test $status -ne 0
        gum style --foreground=$red --bold "Failed to get contexts"
        return 1
    end
    
    gum style --foreground=$green "✓ Found "(count $contexts)" contexts"
    
    if test (count $contexts) -eq 0
        gum style --foreground=$red --bold "No contexts found in kubeconfig"
        return 1
    end
    
    # Let user select context
    echo ""
    set -l selected_context (printf '%s\n' $contexts | gum choose --header "Select OpenShift context:")
    
    if test -z "$selected_context"
        gum style --foreground=$red --bold "No context selected"
        return 1
    end
    
    gum style --foreground=$blue "Selected context: $selected_context"
    
    # Get cluster information
    set -l cluster (oc config view -o yaml | yq '.contexts[] | select(.name=="'$selected_context'").context.cluster' 2>/dev/null)
    if test -z "$cluster"
        gum style --foreground=$red --bold "Failed to get cluster for context $selected_context"
        return 1
    end
    
    set -l api_url (oc config view -o yaml | yq '.clusters[] | select(.name=="'$cluster'").cluster.server' 2>/dev/null)
    if test -z "$api_url"
        gum style --foreground=$red --bold "Failed to get API URL for cluster $cluster"
        return 1
    end
    
    set -l user (oc config view -o yaml | yq '.contexts[] | select(.name=="'$selected_context'").context.user' 2>/dev/null)
    if test -z "$user"
        gum style --foreground=$red --bold "Failed to get user for context $selected_context"
        return 1
    end
    
    # Show cluster information
    echo ""
    gum style --foreground=$blue --bold "Cluster Information:"
    echo "  Context: $selected_context"
    echo "  Cluster: $cluster"
    echo "  API URL: $api_url"
    echo "  User: $user"
    
    # Get token using ocp-sso-token
    echo ""
    gum style --foreground=$blue --bold "Obtaining OAuth token..."
    
    set -l token (ocp-sso-token --identity-providers $provider $api_url)
    set -l ocp_status $status
    
    if test $ocp_status -ne 0 -o -z "$token"
        gum style --foreground=$red --bold "Failed to obtain token from ocp-sso-token (exit code: $ocp_status)"
        return 1
    end
    
    # Update kubeconfig with new token
    if oc config set-credentials $user --token=$token 2>/dev/null
        gum style --foreground=$green --bold "✓ Token updated successfully for user $user"
    else
        gum style --foreground=$red --bold "Failed to update token in kubeconfig"
        return 1
    end
    
    # Test the token
    echo ""
    gum style --foreground=$blue --bold "Testing authentication..."
    
    if oc --context=$selected_context whoami >/dev/null 2>&1
        set -l whoami_result (oc --context=$selected_context whoami 2>/dev/null)
        gum style --foreground=$green --bold "✓ Authentication successful"
        gum style --foreground=$green "Logged in as: $whoami_result"
    else
        gum style --foreground=$red --bold "Authentication test failed"
        return 1
    end
    
    echo ""
    gum style --foreground=$green --bold --border rounded --align=center --padding "0 2" --border-foreground=$green "Login completed successfully!"
end

# If script is being executed directly (not sourced), run the function
if test (status filename) = (realpath (status filename))
    oc-login-token $argv
end