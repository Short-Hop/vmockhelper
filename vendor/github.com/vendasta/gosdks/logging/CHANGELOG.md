# Change Log

This file includes all the change logs for logging's module.
All notable changes to this project will be documented in this file.


The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

**When increasing the version in this file, please remember to also update it in the [VERSION.md](VERSION.md)**


--------------------------------------------------------------------------------

## 1.15.0 - 2022-04-26
- Renamed GKE logger to just Cloud Logger and added the ability to provide a CloudLoggingStrategy in the options (that defaults to gke logging).
This was done to facilitate logging in CloudFunctions bedder so there's a CloudFunctionLoggingStrategy as well

## 1.14.1 - 2022-04-18
- Fix actual request being provided to `NormalizedPathFromRequest`, previously a fake empty request was being provided

## 1.14.0 - 2022-04-08
- Add `NormalizedPathFromRequest` option for client to handle normalizing path from request, appearing in Datadog tags
- Add logging tag of `logging.normalized_path` on requests

## 1.13.3 - 2022-03-29
- Ensure RequestID always generates positive IDs
- Ensure random bigflake workerID is valid

## 1.13.2 - 2022-03-24
- Fix potential dupliciate requestID's in GKE log bundler for requests on pods on the same GKE VM node instance

## 1.13.1 - 2022-03-10
- Replace uses of opencensus trace with our tracing SDK which wraps it

## 1.13.0 - 2022-01-07
- increase statsd version from 1.3.0 to 1.4.0

## 1.12.0 - 2021-11-10
- increase statsd version from 1.0.0 to 1.3.0

## 1.11.1 - 2021-10-29 
- Fix issue where path is not being set properly on tags

## 1.11.0 - 2021-09-28
- Extract function that creates GCE labels

## 1.10.0 - 2021-09-10
- Add `ActiveRequestId` and `ActiveTraceId` functions which can be used to get values that can be displayed to
  users in error messages allowing us to lookup the relevant logs.
- Add `labels.request_ID` to the always logged data
- Refactor the calculation of the trace label. It should maintain the original logic:
  - Use the opencensus trace id when present
  - else bundle based on the request id
  - else no bundling
- Only apply the `projects/repcore-prod/traces/` prefix to `trace` and not `labels.appengine.googleapis.com/trace_id`
  This matches the behaviour of the GCP Trace List tool when linking to logs.

## 1.9.1 - 2021-06-16
- Add fallback path for logging with bundling if the path provided is invalid

## 1.9.0 - 2021-05-31
- Add docker image id to common labels based off of MSCLI environment variable

## 1.8.1 - 2021-05-26
- Add log level inclusion to the request logs to filter out info level requests

## 1.8.0 - 2021-05-25
- Add log level inclusion to allow developers to specify what level of logs are actually logged to cloud console

## 1.7.3 - 2021-05-03
- clarify deprecation message for `ClientInterceptor()`, `ClientStreamInterceptor()`

## 1.7.2 - 2021-03-09
- Stop attaching untraced logs to a traceID of "00000000000000000000000000000000"

## 1.7.1 - 2021-02-04
- Fix missing log format directive on warnings in logging interceptor

## 1.7.0 - 2021-02-04
- Log out InternalMessage from verrors on erroring requests.

## 1.6.0 - 2021-02-01
- Add the filename and line number to log messages of error and above, with the ability to set this to any level.

## 1.5.0 - 2020-12-09
- Now the originating user agent has been passed through, tag it on the request

## 1.4.1 - 2020-12-08
- Add a nil check for config in postRequestHook

## 1.4.0 - 2020-09-18
- Tag `origin`, `referer`, and `host` in logs via logging interceptors for grpc and http.

## 1.3.0 - 2020-08-26
- Upgrade gRPC to 1.31.1

## 1.2.2 - 2020-07-28
- Adds `LocalLoggingOnly` `LoggerOption` which bypasses the namespace and pod
  requirement for logging.Initialize. This allows logging to be initialized
  when running outside of Google Compute Engine.

## 1.2.1 - 2020-06-26
- Downgrade grpc due to problems with 1.30.0

## 1.2.0 - 2020-06-23
- Remove legacy GKE logging support.

## 1.1.0 - 2020-06-22
- Realize support of new GKE logging setting (Kubernetes Engine Logging).

## 1.0.1 - 2020-06-09
- Log literal strings when formatting argument is not specified.

## 1.0.0 - 2020-01-13
- BREAKING CHANGE: publish this directory as a Go module.

## 0.6.0 - 2019-11-25
- Add tag to Datadog request tick with status series (i.e. 4xx for status 404)

## 0.5.1 - 2019-03-08
- Change format of TraceID to what Google expects

## 0.5.0 - 2018-06-28
- Successful requests (non-500s) will not include debug level log lines or below

## 0.4.0 - 2018-04-26
- Logging interceptor () now reports proper HTTP Status codes for both grpc errors and ServiceErrors

## 0.3.0 - 2018-03-29
- Adds logging.WithBundling which may be used to bundle logs for arbitrary tasks (pubsub, taskqueues, etc.)
- Removes GetBundlingCtxForPubsub and GetBundlingCtxForTaskQueues as they are now unused (use logging.WithBundling instead).

## 0.2.0 - 2018-03-01
- Removing hand-rolled tracing from `logging`. Use opencensus library.
- `logging.FromContext` should be replaced with:

```
import "go.opencensus.io/trace"

ctx, span := trace.StartSpan(ctx, "GetMulti")
span.AddAttributes(
    trace.StringAttribute("attr", "my-attr"),
)
defer span.End()
```

## 0.1.1 - 2018-01-30
- fix issue with pubsub contexts ignoring labels from parent contexts

## 0.1.0 - 2018-01-30
- Add StreamInterceptor() for gRPC request logs


## 0.0.0 - 2018-01-30
- First documented release
- Updated HTTPMiddleware to batch logs in Google Cloud Log Viewer
