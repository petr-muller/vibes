#!/usr/bin/env fish

function show_usage
    echo "Usage: test-networkpolicy-external-access.fish <namespace>"
    echo "Tests that NetworkPolicy blocks external access by deploying a netcat server and attempting to reach it via Route"
end

function check_required_binaries
    set required_binaries oc gum curl
    
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
set pod_name "networkpolicy-netcat-server"
set service_name "networkpolicy-netcat-service"
set route_name "networkpolicy-netcat-route"
set port 8080

gum style --foreground="#00ff00" --bold "Testing NetworkPolicy external access blocking in namespace: $namespace"

# Check if namespace exists
if not oc get namespace $namespace >/dev/null 2>&1
    gum style --foreground="#ff0000" "Error: Namespace '$namespace' does not exist"
    exit 1
end

# Create temporary YAML files
set temp_pod_yaml "/tmp/networkpolicy-netcat-pod.yaml"
set temp_service_yaml "/tmp/networkpolicy-netcat-service.yaml"
set temp_route_yaml "/tmp/networkpolicy-netcat-route.yaml"

gum style --foreground="#0000ff" "Creating Pod YAML..."

echo "apiVersion: v1
kind: Pod
metadata:
  name: $pod_name
  namespace: $namespace
  labels:
    app: networkpolicy-test
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
        type: RuntimeDefault" > $temp_pod_yaml

gum style --foreground="#0000ff" "Creating Service YAML..."

echo "apiVersion: v1
kind: Service
metadata:
  name: $service_name
  namespace: $namespace
spec:
  selector:
    app: networkpolicy-test
  ports:
  - port: $port
    targetPort: $port
  type: ClusterIP" > $temp_service_yaml

gum style --foreground="#0000ff" "Creating Route YAML..."

echo "apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: $route_name
  namespace: $namespace
spec:
  to:
    kind: Service
    name: $service_name
  port:
    targetPort: $port" > $temp_route_yaml

echo "Pod YAML content:"
cat $temp_pod_yaml
echo
echo "Service YAML content:"
cat $temp_service_yaml
echo
echo "Route YAML content:"
cat $temp_route_yaml

gum style --foreground="#0000ff" "Applying Pod manifest..."
set pod_output (oc apply -f $temp_pod_yaml 2>&1)

if test $status -ne 0
    gum style --foreground="#ff0000" "Error: Failed to create test pod"
    echo "Command that failed:"
    echo "oc apply -f $temp_pod_yaml"
    echo "oc output:"
    echo $pod_output
    rm -f $temp_pod_yaml $temp_service_yaml $temp_route_yaml
    exit 1
end

gum style --foreground="#0000ff" "Applying Service manifest..."
set service_output (oc apply -f $temp_service_yaml 2>&1)

if test $status -ne 0
    gum style --foreground="#ff0000" "Error: Failed to create service"
    echo "oc output:"
    echo $service_output
    oc delete pod $pod_name --namespace=$namespace >/dev/null 2>&1
    rm -f $temp_pod_yaml $temp_service_yaml $temp_route_yaml
    exit 1
end

gum style --foreground="#0000ff" "Applying Route manifest..."
set route_output (oc apply -f $temp_route_yaml 2>&1)

if test $status -ne 0
    gum style --foreground="#ff0000" "Error: Failed to create route"
    echo "oc output:"
    echo $route_output
    oc delete pod $pod_name --namespace=$namespace >/dev/null 2>&1
    oc delete service $service_name --namespace=$namespace >/dev/null 2>&1
    rm -f $temp_pod_yaml $temp_service_yaml $temp_route_yaml
    exit 1
end

# Wait for pod to be ready
gum style --foreground="#0000ff" "Waiting for pod to be ready..."
oc wait --for=condition=Ready pod/$pod_name --namespace=$namespace --timeout=60s >/dev/null 2>&1
if test $status -ne 0
    gum style --foreground="#ff0000" "Error: Pod failed to become ready"
    oc delete pod $pod_name --namespace=$namespace >/dev/null 2>&1
    oc delete service $service_name --namespace=$namespace >/dev/null 2>&1
    oc delete route $route_name --namespace=$namespace >/dev/null 2>&1
    rm -f $temp_pod_yaml $temp_service_yaml $temp_route_yaml
    exit 1
end

# Get the route URL
set route_url (oc get route $route_name --namespace=$namespace -o jsonpath='{.spec.host}' 2>/dev/null)
if test -z "$route_url"
    gum style --foreground="#ff0000" "Error: Failed to get route URL"
    oc delete pod $pod_name --namespace=$namespace >/dev/null 2>&1
    oc delete service $service_name --namespace=$namespace >/dev/null 2>&1
    oc delete route $route_name --namespace=$namespace >/dev/null 2>&1
    rm -f $temp_pod_yaml $temp_service_yaml $temp_route_yaml
    exit 1
end

gum style --foreground="#0000ff" "Attempting to curl http://$route_url from local machine (this should fail if NetworkPolicy is working)..."

# Wait a bit for the service to be fully ready
sleep 5

set curl_result (curl -s --connect-timeout 10 --max-time 15 http://$route_url 2>/dev/null)
set curl_exit_code $status

# Show test results before cleanup
if test $curl_exit_code -eq 0
    gum style --foreground="#ff0000" --bold "❌ TEST FAILED: NetworkPolicy is NOT blocking external access"
    gum style --foreground="#ff0000" "External access to http://$route_url was successful"
    echo "Response: $curl_result"
    set test_result 1
else
    gum style --foreground="#00ff00" --bold "✅ TEST PASSED: NetworkPolicy is blocking external access"
    gum style --foreground="#00ff00" "External access to http://$route_url was blocked (as expected)"
    set test_result 0
end

# Clean up all resources and temp files
gum style --foreground="#0000ff" "Cleaning up resources..."
oc delete route $route_name --namespace=$namespace >/dev/null 2>&1
oc delete service $service_name --namespace=$namespace >/dev/null 2>&1
oc delete pod $pod_name --namespace=$namespace >/dev/null 2>&1
rm -f $temp_pod_yaml $temp_service_yaml $temp_route_yaml

exit $test_result