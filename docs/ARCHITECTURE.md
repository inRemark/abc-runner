# Servers Module Architecture

## Purpose

The servers/ module is an independent collection of simple protocol servers meant to serve as realistic targets for the abc-runner benchmarking and integration tests. It is intentionally isolated so it can be built, run, and deployed separately from the main benchmarking application.

## High-level Structure

- cmd/: per-protocol server entrypoints (http-server, tcp-server, udp-server, grpc-server, multi-server).
- internal/: shared implementation details for servers (configuration loading, monitoring, logging, common utilities).
- pkg/: public packages that expose server interfaces and protocol implementations.
- config/: per-server configuration files and examples.
- scripts/: convenience scripts to start/stop servers and run health checks.
- docs/: deployment / configuration / API docs for the servers.
- test/: integration and performance tests targeting each server implementation.

## Protocol Support

- HTTP: configurable responses (JSON/XML/text), delay simulation, status code control, payload templates.
- TCP: echo services, long-lived connections, concurrency control.
- UDP: unicast/multicast/broadcast patterns, packet validation, drop rate simulation.
- gRPC: unary and streaming modes, TLS support, token-based auth support.

## Build & Run

- The module contains a separate go.mod allowing it to be built independently.
- Typical build: `cd servers && go build ./cmd/http-server` (or the relevant cmd subpackage).
- Provided scripts: start-all.sh stops/starts multiple server binaries for local testing.

## Configuration

- YAML-based configs per server (network address, ports, concurrency, buffer sizes, simulated failure modes).
- Config examples under config/examples and config/servers.
- Servers support override via command-line flags for path to config file.

## Observability & Health

- Health endpoints and health-check scripts are provided (scripts/health-check.sh).
- Servers expose runtime metrics and structured logs. Recommend Prometheus-compatible metrics where applicable.

## Testing

- test/ contains integration and performance tests used by the main project to validate adapters.
- Servers are designed to be deterministic and configurable for repeatable test scenarios.

## Deployment

- Servers can be containerized individually; Dockerfiles are not present by default but recommended to add.
- For local development, scripts/start-all.sh launches all configured servers.

## Extensibility & API

- pkg/interfaces define server initialization and lifecycle hooks so the main project or other consumers can embed servers programmatically.
- new protocol implementations should follow the `pkg/<protocol>` directory structure and `cmd/<protocol>-server` layout.

## Operational Concerns

- Resource limits should be configured to avoid noisy-neighbor effects during performance runs (connection caps, per-client rate limits).
- TLS certificate management for gRPC/HTTP should be externalized (env/vault/filesystem) and clearly documented.

## Strengths & Recommendations

- Strengths: isolated module, multi-protocol coverage, configurable behavior for reproducible tests.
- Recommendations:
  - Add Dockerfiles and a docker-compose or Kubernetes manifest for reproducible environments.
  - Standardize Prometheus metrics endpoints and add documentation for metrics naming.
  - Add simple examples showing how to run a single protocol server with a sample config.

Generated from servers/ README, directory layout and available config files. For deeper analysis I can scan specific server implementations and tests to enumerate exact endpoints, ports and configuration options—tell me which protocol to inspect first if you want that.
