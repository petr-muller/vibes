#!/usr/bin/env fish

function show_usage
    echo "Usage: test-networkpolicy-pod-to-pod.fish <namespace>"
    echo "Tests that NetworkPolicy blocks pod-to-pod communication by deploying server and client pods"
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
set server_pod_name "networkpolicy-server-pod"
set client_pod_name "networkpolicy-client-pod"
set service_name "networkpolicy-server-service"
set port 8080

gum style --foreground="#00ff00" --bold "Testing NetworkPolicy pod-to-pod blocking in namespace: $namespace"

# Check if namespace exists
if not oc get namespace $namespace >/dev/null 2>&1
    gum style --foreground="#ff0000" "Error: Namespace '$namespace' does not exist"
    exit 1
end

# Create temporary YAML files
set temp_server_yaml "/tmp/networkpolicy-server-pod.yaml"
set temp_client_yaml "/tmp/networkpolicy-client-pod.yaml"
set temp_service_yaml "/tmp/networkpolicy-server-service.yaml"

gum style --foreground="#0000ff" "Creating Server Pod YAML..."

echo "apiVersion: v1
kind: Pod
metadata:
  name: $server_pod_name
  namespace: $namespace
  labels:
    app: networkpolicy-server
spec:
  restartPolicy: Never
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: http-server
    image: python:3.12-alpine
    command: [\"python3\", \"-m\", \"http.server\", \"$port\"]
    ports:
    - containerPort: $port
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
      runAsNonRoot: true
      seccompProfile:
        type: RuntimeDefault" > $temp_server_yaml

gum style --foreground="#0000ff" "Creating Client Pod YAML..."

echo "apiVersion: v1
kind: Pod
metadata:
  name: $client_pod_name
  namespace: $namespace
  labels:
    app: networkpolicy-client
spec:
  restartPolicy: Never
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: curl-client
    image: quay.io/curl/curl:latest
    command: [\"sh\", \"-c\", \"sleep 30 && curl -s --connect-timeout 10 --max-time 15 http://$service_name:$port\"]
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
      runAsNonRoot: true
      seccompProfile:
        type: RuntimeDefault" > $temp_client_yaml

gum style --foreground="#0000ff" "Creating Service YAML..."

echo "apiVersion: v1
kind: Service
metadata:
  name: $service_name
  namespace: $namespace
spec:
  selector:
    app: networkpolicy-server
  ports:
  - port: $port
    targetPort: $port
  type: ClusterIP" > $temp_service_yaml

echo "Server Pod YAML content:"
cat $temp_server_yaml
echo
echo "Client Pod YAML content:"
cat $temp_client_yaml
echo
echo "Service YAML content:"
cat $temp_service_yaml

gum style --foreground="#0000ff" "Applying Server Pod manifest..."
set server_output (oc apply -f $temp_server_yaml 2>&1)

if test $status -ne 0
    gum style --foreground="#ff0000" "Error: Failed to create server pod"
    echo "Command that failed:"
    echo "oc apply -f $temp_server_yaml"
    echo "oc output:"
    echo $server_output
    rm -f $temp_server_yaml $temp_client_yaml $temp_service_yaml
    exit 1
end

gum style --foreground="#0000ff" "Applying Service manifest..."
set service_output (oc apply -f $temp_service_yaml 2>&1)

if test $status -ne 0
    gum style --foreground="#ff0000" "Error: Failed to create service"
    echo "oc output:"
    echo $service_output
    oc delete pod $server_pod_name --namespace=$namespace >/dev/null 2>&1
    rm -f $temp_server_yaml $temp_client_yaml $temp_service_yaml
    exit 1
end

gum style --foreground="#0000ff" "Applying Client Pod manifest..."
set client_output (oc apply -f $temp_client_yaml 2>&1)

if test $status -ne 0
    gum style --foreground="#ff0000" "Error: Failed to create client pod"
    echo "oc output:"
    echo $client_output
    oc delete pod $server_pod_name --namespace=$namespace >/dev/null 2>&1
    oc delete service $service_name --namespace=$namespace >/dev/null 2>&1
    rm -f $temp_server_yaml $temp_client_yaml $temp_service_yaml
    exit 1
end

# Wait for both pods to be ready
gum style --foreground="#0000ff" "Waiting for server pod to be ready..."
oc wait --for=condition=Ready pod/$server_pod_name --namespace=$namespace --timeout=60s >/dev/null 2>&1
if test $status -ne 0
    gum style --foreground="#ff0000" "Error: Server pod failed to become ready"
    oc delete pod $server_pod_name --namespace=$namespace >/dev/null 2>&1
    oc delete pod $client_pod_name --namespace=$namespace >/dev/null 2>&1
    oc delete service $service_name --namespace=$namespace >/dev/null 2>&1
    rm -f $temp_server_yaml $temp_client_yaml $temp_service_yaml
    exit 1
end

gum style --foreground="#0000ff" "Waiting for client pod to be ready..."
oc wait --for=condition=Ready pod/$client_pod_name --namespace=$namespace --timeout=60s >/dev/null 2>&1
if test $status -ne 0
    gum style --foreground="#ff0000" "Error: Client pod failed to become ready"
    oc delete pod $server_pod_name --namespace=$namespace >/dev/null 2>&1
    oc delete pod $client_pod_name --namespace=$namespace >/dev/null 2>&1
    oc delete service $service_name --namespace=$namespace >/dev/null 2>&1
    rm -f $temp_server_yaml $temp_client_yaml $temp_service_yaml
    exit 1
end

# Wait for client pod to complete its curl attempt (it has a 30s delay built in)
gum style --foreground="#0000ff" "Waiting for client pod to complete curl attempt to server service..."
sleep 45

# Check client pod exit code to determine if curl succeeded
set pod_exit_code (oc get pod $client_pod_name --namespace=$namespace -o jsonpath='{.status.containerStatuses[0].state.terminated.exitCode}' 2>/dev/null)
set curl_result (oc logs $client_pod_name --namespace=$namespace 2>/dev/null)

# If pod is still running, curl might be hanging (network blocked)
set pod_phase (oc get pod $client_pod_name --namespace=$namespace -o jsonpath='{.status.phase}' 2>/dev/null)
if test "$pod_phase" = "Running"
    set curl_exit_code 1
    set curl_result "Connection timeout/blocked"
else if test -n "$pod_exit_code"
    set curl_exit_code $pod_exit_code
else
    set curl_exit_code 1
end

# Show test results before cleanup
if test $curl_exit_code -eq 0
    gum style --foreground="#ff0000" --bold "❌ TEST FAILED: NetworkPolicy is NOT blocking pod-to-pod communication"
    gum style --foreground="#ff0000" "Client pod was able to reach server pod via service"
    echo "Response: $curl_result"
    set test_result 1
else
    gum style --foreground="#00ff00" --bold "✅ TEST PASSED: NetworkPolicy is blocking pod-to-pod communication"
    gum style --foreground="#00ff00" "Client pod was unable to reach server pod via service (as expected)"
    set test_result 0
end

# Clean up all resources and temp files
gum style --foreground="#0000ff" "Cleaning up resources..."
oc delete pod $server_pod_name --namespace=$namespace >/dev/null 2>&1
oc delete pod $client_pod_name --namespace=$namespace >/dev/null 2>&1
oc delete service $service_name --namespace=$namespace >/dev/null 2>&1
rm -f $temp_server_yaml $temp_client_yaml $temp_service_yaml

exit $test_result