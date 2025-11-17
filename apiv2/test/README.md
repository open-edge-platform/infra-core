# Integration Tests

## Requirements

It requires golang and kubectl installed, besides an operational Kubernetes cluster
with EMF (Edge Manageability Framework) deployed.

The tests consider to have access to the cluster.

```bash
# Set variables needed for the tests
export KEYCLOAK_URL="https://keycloak.kind.internal"
export PASSWORD='<REPLACE_WITH_PASSWORD>'
export USERNAME="sample-project-api-user"
export API_URL="http://127.0.0.1:8080/"
export CA_PATH="orch-ca.crt"

PROJECT_ID=$(kubectl get projects.project -o json | jq -r ".items[0].status.projectStatus.uID")

JWT_TOKEN=$(curl -k --location --request POST ${KEYCLOAK_URL}/realms/master/protocol/openid-connect/token --header 'Content-Type: application/x-www-form-urlencoded' --data-urlencode 'grant_type=password' --data-urlencode 'client_id=system-client' --data-urlencode username=${USERNAME} --data-urlencode password=${PASSWORD} --data-urlencode 'scope=openid profile email groups' | jq -r '.access_token')
```

The integration tests require access to the Infrastructure Manager REST API,
the the project ID which is associated with the USERNAME,
the path of the cluster CA file, and the JWT token associated with the USERNAME credentials.

Notice: change the names of the variables according to your deployment setup.

## Run

Then enable a port-forward to have interface with the API component via port 8080.

```bash
kubectl port-forward svc/apiv2-proxy -n orch-infra --address 0.0.0.0 8080:8080 &
```

Run the integration tests:

```bash
# A test case can be specified after the `go test` statement, such as: -run TestHostCustom
JWT_TOKEN=${JWT_TOKEN} PROJECT_ID=${PROJECT_ID}  go test -v -count=1 ./test/client/ -apiurl=${API_URL} -caPath=${CA_PATH} 
```

or using Make target:

```bash
make int-test JWT_TOKEN=${JWT_TOKEN} PROJECT_ID=${PROJECT_ID} API_URL=${API_URL} CA_PATH=${CA_PATH}
```

Kill the port-forward command:

```bash
kill $(ps -eaf | grep 'kubectl' | grep 'port-forward svc/apiv2-proxy' | awk '{print $2}')
```
