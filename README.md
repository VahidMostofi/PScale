## Wise Autoscaler
Uses configurations found by [swarmmanager](https://github.com/vahidmostofi/swarmmanager) to efficiently scale a microservice application in Kubernetes.

### Evaluate configs
set these environment variables:

- ``` WA_EVALUATE_INTERVAL ``` in seconds
- ``` WA_EVALUATE_REPORT_PATH ``` the directory which the report will be saved into, ending with \, the file name will be ```SystemName + - + $(RandomString) + .yml```

### auto-scale
- to evaluate while auto-scaling, set ```WA_EVALUATE_ENABLE``` to ```true``` (default to true)

- ```WA_AUTOSCALE_INTERVAL``` specifies the monitor interval for autoscaler in seconds

### only evaluate
to only evaluate, set the evaluate configs and run

``` go run main.go devaluate ```

or

``` go run main.go eval ```