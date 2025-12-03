<div align="center">

# S3Local

**A lightweight, S3-compatible local storage server with event-driven architecture**

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Container](https://img.shields.io/badge/container-ghcr.io-purple)](https://github.com/tkasuz/s3local/pkgs/container/s3local%2Fs3local)

[Features](#features) • [Quick Start](#quick-start) • [Architecture](#architecture) • [Examples](#examples) • [Development](#development)

</div>

---

## Overview

S3Local is a lightweight, self-contained S3-compatible storage server designed for local development and testing. It provides a complete S3 API implementation with event notification support, making it perfect for developing and testing applications that rely on AWS S3 without needing actual AWS infrastructure.

## Features

### Core Capabilities
- **S3-Compatible API** - Full compatibility with AWS S3 SDKs and CLI tools
- **Event-Driven Architecture** - Real-time event notifications for S3 operations
- **Multi-Format Support** - Works with any S3 client library
- **Lightweight & Fast** - Single binary with minimal dependencies
- **SQLite Backend** - Reliable local storage with zero configuration

### Event Notifications
- **Lambda Integration** - HTTP webhook support for serverless functions
- **SQS Integration** - Queue-based event processing
- **SNS Integration** - Pub/sub event notifications
- **Event Types** - ObjectCreated, ObjectRemoved, and more

### Developer Experience
- **Easy Setup** - Run with Docker or standalone binary
- **Web Console** - Modern React-based UI for bucket management
- **Hot Reload** - Fast development iteration with Air
- **Test-Friendly** - Perfect for integration tests and CI/CD

## Quick Start

### Using Docker

```bash
# Pull and run the API server
docker pull ghcr.io/tkasuz/s3local/s3local:latest
docker run -p 8080:8080 ghcr.io/tkasuz/s3local/s3local:latest

# Pull and run the web console
docker pull ghcr.io/tkasuz/s3local/s3local-web:latest
docker run -p 3000:80 ghcr.io/tkasuz/s3local/s3local-web:latest
```

### Using Docker Compose

```yaml
version: '3.8'
services:
  s3local:
    image: ghcr.io/tkasuz/s3local/s3local:latest
    ports:
      - "8080:8080"
    volumes:
      - s3data:/data

  web:
    image: ghcr.io/tkasuz/s3local/s3local-web:latest
    ports:
      - "3000:80"
    environment:
      - S3LOCAL_API_URL=http://s3local:8080

volumes:
  s3data:
```

### Using AWS CLI

```bash
# Configure AWS CLI for local use
aws configure set aws_access_key_id test
aws configure set aws_secret_access_key test
aws configure set region us-east-1

# Create a bucket
aws s3 mb s3://my-bucket --endpoint-url http://localhost:8080

# Upload a file
aws s3 cp myfile.txt s3://my-bucket/ --endpoint-url http://localhost:8080

# List objects
aws s3 ls s3://my-bucket/ --endpoint-url http://localhost:8080
```

## Architecture

S3Local is built with a modern, modular architecture:

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  S3 Client  │────▶│   S3 API    │────▶│   Storage   │
│  (AWS SDK)  │     │   Handler   │     │  (SQLite)   │
└─────────────┘     └──────┬──────┘     └─────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │   Event      │
                    │ Dispatcher   │
                    └──────┬───────┘
                           │
        ┌──────────────────┼──────────────────┐
        ▼                  ▼                  ▼
   ┌────────┐         ┌────────┐        ┌────────┐
   │ Lambda │         │  SQS   │        │  SNS   │
   └────────┘         └────────┘        └────────┘
```

For detailed architecture documentation, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Examples

### Event-Driven Processing

Check out the [example](example/) directory for a complete setup with:
- S3Local server
- ElasticMQ (local SQS)
- Python Lambda function
- Docker Compose configuration

```bash
cd example
docker-compose up -d
```

See the [example README](example/README.md) for detailed usage instructions.

## Development

### Prerequisites
- Go 1.23+
- Node.js 22+ (for web console)
- Docker (optional)

### API Development

```bash
cd api

# Install dependencies
go mod download

# Run with hot reload
make dev

# Run tests
make test

# Build binary
make build
```

### Web Console Development

```bash
cd web

# Install dependencies
npm install

# Start dev server
npm run dev

# Build for production
npm run build
```

### Building Container Images

```bash
# Build API image with ko
cd api
ko build ./cmd/s3local --bare

# Build web image with Docker
cd web
docker build -t s3local-web .
```

## Project Structure

```
s3local/
├── api/              # Go-based S3 API server
│   ├── cmd/         # Application entrypoints
│   ├── internal/    # Internal packages
│   └── pkg/         # Public packages
├── web/             # React-based web console
│   └── src/         # Source code
├── example/         # Example configurations
└── docs/            # Documentation
```

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

---

<div align="center">

**Built with Go, React, and SQLite**

[Report Bug](https://github.com/tkasuz/s3local/issues) • [Request Feature](https://github.com/tkasuz/s3local/issues)

</div>
