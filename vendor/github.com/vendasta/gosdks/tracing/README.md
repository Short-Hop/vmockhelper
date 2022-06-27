#Package `tracing`

Distributed Tracing package for Vendasta microservices, wraps go.opencensus.io.

Currently, the library implements tracing via opencensus and ships traces to Google Cloud Trace.

Tracing is provided automatically to our microservices by the `serverconfig` and `logging` SDKs.  