# agent-prxy

<p align="center">
  <img src="assets/gopher.png" alt="agent-prxy gopher" width="180"><br>
  <strong>A minimal Go web service for proxying agent interactions.</strong><br>
  <sub>Gopher artwork by Renee French</sub>
</p>

## Overview

`agent-prxy` sits in front of agent or LLM workloads and exposes a small HTTP surface for routing and controlling agent-initiated requests.

## Features

- Minimal HTTP service
- Environment-based configuration
- Health endpoint

> Status: early-stage; interfaces may change.

## Direction

This project is in its early stages. Future work may include isolated execution and deeper inspection of agent-driven tool calls.

## Getting Started

### Prerequisites

- Go `1.24.9`

### Run

```bash
go run ./cmd
```

### Health Check
```bash
curl http://localhost:8080/healthz
```

## Configuration

Configuration is provided via environment variables.

| Variable            | Required | Default                | Description |
|---------------------|----------|------------------------|-------------|
| `LISTEN_ADDR`       | No       | `:8080`                | HTTP bind address |
| `UPSTREAM_BASE_URL` | No       | `https://api.openai.com` | Upstream LLM base URL |
| `UPSTREAM_API_KEY`  | Yes      | —                      | API key for upstream |
| `TOOL_DIR`          | No       | `./tools.d`            | Directory containing tool definition YAML files |
| `TELEMETRY_PATH`    | No       | `./runs.jsonl`         | Path to telemetry output file (JSONL) |

> UPSTREAM_API_KEY must be provided via environment variables and should not be committed.

### Example

```bash
export LISTEN_ADDR=":8080"
export UPSTREAM_BASE_URL="https://api.openai.com"
export UPSTREAM_API_KEY="..."
export TOOL_DIR="./tools.d"
export TELEMETRY_PATH="./runs.jsonl"
```

## License
[Apache License](./LICENSE)
