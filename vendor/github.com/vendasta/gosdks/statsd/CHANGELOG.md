# Change Log

This file includes all the change logs for statsd's module.
All notable changes to this project will be documented in this file.


The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

**When increasing the version in this file, please remember to also update it in the [VERSION.md](VERSION.md)**


--------------------------------------------------------------------------------
## 1.4.0 - 2022-01-07
- Update datadog-go client from 4.8.3 to 5.0.2

## 1.3.0 - 2021-10-09
- Update datadog-go client from 3.3.1 to 4.8.3

## 1.2.0 - 2021-07-19
- Change `InitializeCustomClient` to return an interface from the Datadog statsd library instead of a concrete pointer. Technically this is a breaking change, but the existing function is used in only one place.
- Return a no op client when `InitializeCustomClient` is called on local environments
- Add some cautionary docstrings about how this package parameterizes itself

## 1.1.0 - 2021-04-08
- Add method to create a custom statsd client, that is not the global default

## 1.0.0 - 2020-01-13
- BREAKING CHANGE: publish this directory as a Go module.

## 0.3.0 - 2018-11-19
* Add tag if available to outgoing statsd metrics

## 0.2.0 - 2018-08-22
* Add support for Distribution metrics
* See https://docs.datadoghq.com/graphing/metrics/distributions/ for info

## 0.1.0 - 2018-04-26
* Add `WithMonitoring` helper for tracking function calls; see docstring

## 0.0.1 - 2017-12-13
* Fixed a bug where global tags were not being set

## 0.0.0 - 2017-12-13
* History missing
