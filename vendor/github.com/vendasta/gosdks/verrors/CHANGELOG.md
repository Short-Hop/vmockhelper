# Change Log

This file includes all the change logs for verrors' module.
All notable changes to this project will be documented in this file.


The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

**When increasing the version in this file, please remember to also update it in the [VERSION.md](VERSION.md)**


--------------------------------------------------------------------------------

## 1.4.0
- Add `WithInternalMessage` and `GetInternalMessage` to service errors.
- `WrapError` now returns an explicit ServiceError type.
## 1.3.0
- IsError on nil error now always returns `false`. Previously this behaviour was undefined.
## 1.2.0
- Added `Gone` ServiceError type
- Added 410 status code mapping to StatusCodeToGRPCError to return a Gone
- Added Gone mapping to GRPC code `FailedPrecondition`
## 1.1.0
- Update gRPC to 1.31.1
## 1.0.2
Added 402 status code mapping to StatusCodeToGRPCError to return a FailedPrecondition
## 1.0.1
Added additional error conversion to FromError function. 
## 1.0.0 - 2020-01-13
- BREAKING CHANGE: publish this directory as a Go module.
