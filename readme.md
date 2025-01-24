1. Create test Project:
```
oc new-project test
```
2. Create new builds
```
oc new-build registry.access.redhat.com/ubi9/go-toolset --name=sleep --binary -n test
oc new-build registry.access.redhat.com/ubi9/go-toolset --name=metrics --binary -n test
```
3. Launch builds with local code
```
oc start-build metrics --from-dir=./metrics
oc start-build sleep --from-dir=./sleep
```
4. Check when the builds are completed
```
watch oc get pods -n test
```
5. Apply manifests
```
oc apply -f manifests
```
6. Once the sleep pod is running, scale metrics deployment:
```
oc scale deployment/metrics --replicas=5
```
7. Once everything is deployed, you can check the http_request_errors_total metric in Observe -> Metrics until the first pod shows a peak above 1 second. By accessing the log of that pod, you will be able to see the timestamp and the source port.
