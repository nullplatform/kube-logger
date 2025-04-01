# kube-logger

A command-line tool for fetching logs from Kubernetes pods using nullplatform selectors.

## Overview

kube-logger allows you to fetch logs from Kubernetes pods based on various selectors like application ID, scope ID, and deployment ID. It provides easy filtering, pagination, and output formatting options.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/your-username/kube-logger.git
cd kube-logger

# Build the tool
make build

# The binary will be available in the build directory
./build/kube-logger --help
```

### Pre-built Binaries

Pre-built binaries are available on the [Releases page](https://github.com/your-username/kube-logger/releases).

## Usage

```bash
kube-logger -namespace <namespace> [options]
```

### Required Arguments

- `-namespace`: Kubernetes namespace to fetch logs from

### Optional Arguments

- `-application-id`: Filter by application ID
- `-scope-id`: Filter by scope ID
- `-deployment-id`: Filter by deployment ID
- `-limit`: Maximum number of log entries to return (default: 100)
- `-filter`: Text pattern to filter logs
- `-start-time`: Start time for logs in RFC3339 format
- `-token`: Pagination token for fetching next batch of logs
- `-kubeconfig`: Path to kubeconfig file (defaults to in-cluster or ~/.kube/config)
- `-help`: Show help

## Examples

### Basic Usage

```bash
kube-logger -namespace nullplatform
```

### Filter by Application

```bash
kube-logger -namespace nullplatform -application-id 1691688910
```

### Complete Example with Multiple Filters

```bash
kube-logger -namespace nullplatform -application-id 1691688910 -scope-id 760499159 -deployment-id 1705961777 -limit 10
```

### Using Pagination

To fetch the next page of logs:

```bash
kube-logger -namespace nullplatform -application-id 1691688910 -token <token_from_previous_output>
```

## How It Works

kube-logger uses Kubernetes selectors to find pods matching your criteria and fetches their logs. It supports:

1. Filtering logs by application, scope, and deployment IDs
2. Limiting the number of returned log entries
3. Pagination for handling large log volumes
4. Text pattern filtering within logs

## Development

### Prerequisites

- Go 1.16 or higher
- Access to a Kubernetes cluster

### Building

```bash
make build
```

### Running Tests

```bash
make test
```

## License

[License details]