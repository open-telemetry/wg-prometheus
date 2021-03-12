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
  difference causes how the state is maintained throughout the
  collection including how it’s handled in the OpenTelemetry
  collector.
* **Cumulative vs delta**: OpenTelemetry supports delta temporality
  whereas Prometheus always expects absolute/cumulative values. This
  breaks some components that produce and communicate deltas where
  cumulatives cannot be rebuilt before being exported to Prometheus.
* **Histogram boundaries**: Prometheus histogram boundaries are by
  lower equal (le) while OpenTelemetry histogram boundaries are
  greater equal (ge). (This difference will be resolved via #18)
* **Semantic conventions**: OpenTelemetry predefines semantic
  conventions to collect additional metadata with telemetry data.
  Prometheus users don’t follow the same conventions and Prometheus
  client library provided data may lack the semantic conventions
  available in OpenTelemetry libraries.

## Compatibility Requirements

Given the number of fundamental design goals between OpenTelemetry
and Prometheus, our aim is to close the gaps where possible and
make the right compromises to meet the goals defined in this document.
OpenTelemery and Prometheus won’t be fully compatible but we will
enable important use cases to enable OpenTelemetry for Prometheus
users. The following sections summarizes the expectations from
the collector and the libraries.

### Collector

* Collector will implement a Prometheus remote write exporter.
  Publishing a pull-based metrics handler with all collected
  metrics is not a scalable approach.
* Collector will support scraping and ingesting cumulative metrics.
  Prometheus doesn’t support deltas and there are cases where
  rebuilding the cumulatives from deltas is not possible/easy.
* Collector will support all discovery and scraping configuration
  options in the Prometheus server. Collector will ignore the
  alerting rules.
* Collector will support exporting cumulative series to
  OTLP-compatible exporters.
* At each scrape, collector will produce an "up" metric as a gauge
  internally in the reciever and exporter will translate it to a
  Prometheus up metric in the export time.
* Collector will produce "instance" and "job" labels similar
  to the Prometheus server.
* If a target disappears from the scrape, the collector will
  write explicit staleness markers for the respective timeseries.
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
