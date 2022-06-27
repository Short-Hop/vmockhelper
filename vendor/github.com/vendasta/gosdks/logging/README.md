#Package `logging`

Provide a set of functions to flow your applications' logs to Google Cloud Console

## Getting Started

### Prerequisites

You have to define environment variable `HOSTNAME` in your deployment yaml file in order to user this package 
([mscli](https://github.com/vendasta/mscli) may already did this for you):

```yaml
- name: HOSTNAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
```

### Example

```go
package main
import (
    "log"
    "context"
    
    "github.com/vendasta/gosdks/logging"
)

func main() {
    namespace := "Your app's namespace"
    podName := "Your app's pod name"
    appName := "Your app's name"
    ctx := context.Background()

    var opts []logging.LoggerOption
    if config.IsLocal() {
        opts = []logging.LoggerOption{logging.LocalLoggingOnly()}
    }
    err := logging.Initialize(namespace, podName, appName, opts...)
    if err != nil {
        log.Fatalln("error initializing logging module, err: %s", err.Error())
    }

    logging.Infof(ctx, "info blahblah...")
    logging.Warningf(ctx, "warning blahblah...")
    logging.Errorf(ctx, "error blahblah...")
}
```

## Setting log inclusion level
It is useful to control what severity of logs a service is shipping, for example, we may want all debug and above logs. What severity of logs that are shipped to the cloud console is controlled by the `LOG_INCLUSION_LEVEL` environment variable.
This variable can be set under `podEnv` in the `microservice.yaml` file
```yaml
      podEnv:
      - key: LOG_INCLUSION_LEVEL
        value: info
```

appropriate values for this environment variable are:
- debug
- info
- warning
- error

## Normalizing request paths
The logging interceptor sends Histogram statsd data on the metric `{microserviceName}.gRPC.Latency`, and one of the tags is the path of the request. When your paths may contain dynamic values use the `NormalizedPathFromRequest` logger option to normalize them.

Some use cases for path normalizing:

REST normalizing to method and normalized path: GET `/order/ORD-123` => `GET /order/{orderId}`
REST normalizing to operation id: GET `/order/ORD-123` => `orders-get`
vstatic normalizing cache-ids out of file paths: `/some-client.123456.prod/main.123456.js` => `/some-client/main.js`

## Tagging Your Logs

It might be useful to add tags to your logs so that you can filter them easier.

```go
logging.Tag(ctx, "key", "value")
```

Will attach a label `"key": "value"` to entries logged with that particular `ctx`.

Tagging contexts that originate from a GRPC or HTTP entrypoint will just work. In order to tag a context that is not associated with a request (such as a background job), you will need to use the context returned by `NewTaggedContext`.

```go
// this can be any context, not just context.Background()
ctx := logging.NewTaggedContext(context.Background())
logging.Tag(ctx, "key", "value")
```

If you need to differentiate between different contexts operating in the same function, you may find it useful to have a UUID in order to tie together a single stream of execution. You can use the helper:
```go
ctx := logging.NewWorkerContext(context.Background())
```
