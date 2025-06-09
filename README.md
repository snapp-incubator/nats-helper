# NATS Helper

A command-line tool for monitoring and managing NATS streams and consumers, with Prometheus metrics support.

## Features

- Monitor NATS stream and consumer events in real-time
- Export events as Prometheus metrics for monitoring and alerting
- Track stream and consumer operations, configurations, and state changes
- Configurable via environment variables, config file, or command-line flags
- Support for NATS authentication

## Installation

```bash
go install github.com/snapp-incubator/nats-helper@latest
```

## Configuration

The tool can be configured using one of the following methods:

1. Environment variables:
   ```bash
   export NATS_HELPER_NATS_URL="nats://localhost:4222"
   export NATS_HELPER_NATS_USER="user"     # optional
   export NATS_HELPER_NATS_PASSWORD="pass" # optional
   ```

2. Configuration file (default: `$HOME/.nats-helper.yaml`):
   ```yaml
   nats:
     url: "nats://localhost:4222"
     user: "user"     # optional
     password: "pass" # optional
   ```

3. Command-line flags:
   ```bash
   nats-helper export --config /path/to/config.yaml --metrics-addr :9090
   ```

## Usage

### Export Events as Prometheus Metrics

```bash
nats-helper export --metrics-addr :9090
```

This will start the event exporter and expose Prometheus metrics at `http://localhost:9090/metrics`.

#### Available Metrics

The following Prometheus metrics are exposed:

**Stream Metrics:**
- `nats_stream_operations_total`: Counter of stream operations
  - Labels: `stream`, `operation`
  - Operations tracked:
    - `create`: Stream creation
    - `delete`: Stream deletion
    - `update`: Stream configuration update
    - `NEW_CONSUMER`: New consumer added to stream
    - `purge`: Stream purge operation
    - `seal`: Stream seal operation

- `nats_stream_config_changes_total`: Counter of stream configuration changes
  - Labels: `stream`, `config_field`
  - Fields tracked:
    - `max_msgs`: Maximum messages limit
    - `max_bytes`: Maximum bytes limit
    - `max_age`: Maximum age limit

- `nats_stream_state`: Gauge of current stream state
  - Labels: `stream`, `metric`
  - Metrics tracked:
    - `messages`: Current number of messages
    - `bytes`: Current size in bytes
    - `first_seq`: First message sequence
    - `last_seq`: Last message sequence
    - `consumer_count`: Number of consumers

**Consumer Metrics:**
- `nats_consumer_operations_total`: Counter of consumer operations
  - Labels: `stream`, `consumer`, `operation`
  - Operations tracked:
    - `create`: Consumer creation
    - `delete`: Consumer deletion
    - `update`: Consumer configuration update

- `nats_consumer_config_changes_total`: Counter of consumer configuration changes
  - Labels: `stream`, `consumer`, `config_field`
  - Fields tracked:
    - `max_deliver`: Maximum delivery attempts
    - `ack_policy`: Acknowledgment policy
    - `replay_policy`: Message replay policy

- `nats_consumer_state`: Gauge of current consumer state
  - Labels: `stream`, `consumer`, `metric`
  - Metrics tracked:
    - `num_pending`: Number of pending messages
    - `num_ack_pending`: Number of unacknowledged messages
    - `num_redelivered`: Number of redelivered messages

#### Example Prometheus Queries

```promql
# Total number of stream operations by type
sum(nats_stream_operations_total) by (stream, operation)

# Current number of messages in each stream
nats_stream_state{metric="messages"}

# Number of consumers per stream
nats_stream_state{metric="consumer_count"}

# Consumer operations by type
sum(nats_consumer_operations_total) by (stream, consumer, operation)

# Streams with most configuration changes
topk(5, sum(nats_stream_config_changes_total) by (stream))

# Consumers with pending messages
nats_consumer_state{metric="num_pending"} > 0
```

#### Alerting Examples

```yaml
# Alert for streams with high message count
groups:
- name: nats_alerts
  rules:
  - alert: HighStreamMessageCount
    expr: nats_stream_state{metric="messages"} > 1000000
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Stream {{ $labels.stream }} has high message count"
      description: "Stream {{ $labels.stream }} has {{ $value }} messages"

# Alert for consumers with many pending messages
  - alert: HighConsumerPendingMessages
    expr: nats_consumer_state{metric="num_pending"} > 1000
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Consumer {{ $labels.consumer }} in stream {{ $labels.stream }} has high pending messages"
      description: "Consumer has {{ $value }} pending messages"
```

## Event Types

The tool monitors and processes the following NATS event types:

### Stream Events
- Stream creation and deletion
- Stream configuration updates
- Stream state changes
- Stream purge and seal operations
- Consumer addition to streams

### Consumer Events
- Consumer creation and deletion
- Consumer configuration updates
- Consumer state changes
- Message acknowledgment events

## Development

### Prerequisites

- Go 1.21 or later
- NATS server with JetStream enabled

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/snapp-incubator/nats-helper.git
   cd nats-helper
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the project:
   ```bash
   go build
   ```

## License

MIT