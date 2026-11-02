# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 0.1.0 (2026-11-02)


### Bug Fixes

* **deps:** bump golang.org/x/net to v0.62.0 for http2 CVEs ([64e795f](https://github.com/manovaspace/orbit-observability/commit/64e795f6967e95a93f67ee42dd458a19b4563f0d))
* **observability:** remove empty critical section left after cfg removal ([c581343](https://github.com/manovaspace/orbit-observability/commit/c581343c275e3dc3a99747c293d4849034ff3108))
* **observability:** remove unused cfg package-level variable ([cd4d521](https://github.com/manovaspace/orbit-observability/commit/cd4d521b726cec77f61f9d2f1d97d1c772de359a))
* **observability:** use resource.NewSchemaless to prevent schema url conflicts across OTEL updates ([d7e0a5d](https://github.com/manovaspace/orbit-observability/commit/d7e0a5df8e7c63e2964e829f11937362ba71c9c3))


### Code Refactoring

* adopt the Go library standard ([#23](https://github.com/manovaspace/orbit-observability/issues/23)) ([612d766](https://github.com/manovaspace/orbit-observability/commit/612d766a8e370de771c55ef1cf21e2739616f973))

## [Unreleased]

### Added

- Initial public MIT release.
