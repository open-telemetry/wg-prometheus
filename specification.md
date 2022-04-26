# OpenTelemetry/Prometheus Compatibility Specification

Status: [Experimental](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/document-status.md)

## Abstract

OpenTelemetry is aiming to provide compatibility with
Prometheus and OpenMetrics. This document explains the
extent of the support. OpenTelemetry collector and libraries
will align with the compatibility requirements defined in
this spec.

## Goals

* OpenTelemetry collector can be used as a drop-in replacement
  for Prometheus server to scrape and export metrics data.
* OpenTelemetry collector should export OTLP-compatible Prometheus
  time series to OpenTelemetry metrics exporters.
* OpenTelemetry libraries should implement exporters to publish
  Prometheus metrics.
* To the OpenTelemetry collector, an OpenTelemetry target and
  a Prometheus instrumented target is indistinguishable.

## Differences and Limitations

OpenTelemetry and Prometheus/OpenMetrics have different design
goals and this reflects in the data model and implementation
details. This section summarizes a few key differences.

* **Pull vs push**: Prometheus is mainly designed for pull
  whereas OpenTelemetry primarily is designed for push. This
  difference changes how the state is maintained throughout the
  collection pipeline, including how it’s handled in the
  OpenTelemetry collector.
* **Cumulative vs delta**: OpenTelemetry supports delta temporality
  whereas Prometheus always expects absolute/cumulative values. This
  may result in deltas (collected by OTel client libraries) not being
  able to exported to Prometheus, but Prometheus instrumented metrics
  will be fully supported because they are cumulative.
* **Histogram boundaries**: Prometheus histogram boundaries are by
  lower equal (le) while OpenTelemetry histogram boundaries are
  greater equal (ge). This soon will be fixed by
  [opentelemetry-proto#262](https://github.com/open-telemetry/opentelemetry-proto/pull/262).
* **Semantic conventions**: OpenTelemetry predefines semantic
  conventions to collect additional metadata with telemetry data.
  Prometheus users don’t follow the same conventions and Prometheus
  client library provided data may lack the semantic conventions
  available in OpenTelemetry libraries.

## Implementation Requirements

### Collector

* Collector will implement a Prometheus remote write exporter.
  Collector is a common metrics sink in collection pipelines where
  metric data points are recieved and quickly "forwarded" to exporters.
  Implementing a pull-based metrics handler will require additional
  design work in this model to be efficient, we may follow-up
  with improvements to enable a metrics handler.

* Collector will support scraping and ingesting cumulative metrics.
  Collector will not try to rebuild the cumulatives from deltas
  at this moment but we may improve this case in the future.
* Collector will support all global, discovery and scraping configuration
  options in the Prometheus server. Collector will ignore the
  alerting rules.
* Collector MUST pass the [Prometheus Remote Write Sender Compliance Tests](https://github.com/prometheus/compliance/tree/main/remote_write_sender).
* Collector will support exporting cumulative series to
  OTLP-compatible exporters.
* Collector won’t assume any OpenTelemetry semantic conventions
  might be in place in the scraped data. Collector may decorate
  the samples with some semantic convention attributes available
  in the collector.
* Collector will support Promeheus WAL if it has access
  to persistent storage.
* Collector will provide remote write fine tuning options similar
  to Prometheus server.

### Libraries

* Libraries should implement a Prometheus metrics handler that
  will listen to a user-specified host:port.
* Libraries should provide Prometheus metrics in text format and
  may have protobuf support.
