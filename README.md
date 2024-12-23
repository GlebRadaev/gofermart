[![Go Report Card](https://goreportcard.com/badge/github.com/glebradaev/gofermart)](https://goreportcard.com/report/github.com/glebradaev/gofermart) [![gophermart](https://github.com/glebradaev/gofermart/actions/workflows/gofermart.yml/badge.svg)](https://github.com/glebradaev/gofermart/actions/workflows/gofermart.yml) [![go version](https://img.shields.io/badge/golang-v1.22.7-lightblue)](https://go.dev/)

# Gofermart Accrual System

## Table of Contents

- [Project Purpose](#project-purpose)
- [Architecture](#architecture)
- [Requirements](#requirements)
- [Setup Instructions](#setup-instructions)
  - [Environment Setup](#environment-setup)
  - [Running with Docker Compose](#running-with-docker-compose)
- [API Description](#api-description)
- [Testing](#testing)
  - [Running Tests](#running-tests)
  - [Test Coverage](#test-coverage)
- [Notes](#notes)
  - [Available Commands in Makefile](#available-commands-in-makefile)

## Project Purpose

The project is designed to automate the accounting and accrual of bonus points, providing an interface for managing orders and calculating rewards. The system utilizes two services:

- **Gofermart** — the main service for order management.
- **Accrual** — the bonus accrual service.

## Architecture

The project consists of two services:

- **Gofermart** — the primary service that interacts with the user via a REST API and manages data.
- **Accrual** — an auxiliary service that performs accrual calculations.

Technologies:

- Go version 1.22.7
- PostgreSQL for data storage.
- Docker Compose for container orchestration.

Functional Features:

- Service health checks using `healthcheck`.
- Automatic container restarts upon failures via the `restart` parameter.
- All environment variables are defined in the `.env` file.
- Communication between services via HTTP.

## Requirements

- Go version 1.22.7
- Docker and Docker Compose
- PostgreSQL

## Setup Instructions

### Environment Setup

Clone the repository:

```bash
git clone https://github.com/GlebRadaev/gofermart.git
```

Navigate to the project root directory:

```bash
cd gofermart
```

Copy the `env.example` file and rename it to `.env`:

```bash
cp env.example .env
```

Ensure that the `.env` file contains the correct values for the following variables:

- `DATABASE_USERNAME` — the PostgreSQL username.
- `DATABASE_PASSWORD` — the PostgreSQL password.
- `DATABASE_NAME` — the database name.
- `DATABASE_PORT` — the PostgreSQL port (default: 5432).
- `ACCRUAL_PORT` — the accrual service port.
- `GOFERMART_PORT` — the Gofermart service port.

### Running with Docker Compose

Start the services:

```bash
docker compose up --build
```

After successful startup, the services will be available at the specified ports in the `.env` file.

## API Description

A detailed description of the endpoints is available in the `SPECIFICATION.md` file.

## Testing

Unit tests are implemented to verify the correctness of the project.

### Running Tests

Make sure the project is set up according to the Setup Instructions.

To run tests, use the following Makefile command:

```bash
make test
```

### Test Coverage

For convenient viewing of the test coverage, use the command:

```bash
make coverage
```

This command runs tests with coverage and generates an HTML coverage report.

## Notes

### Available Commands in Makefile

- **lint**: Runs code linting with golangci-lint.

```bash
make lint
```

- **test**: Runs the project's unit tests.

```bash
make test
```

- **vet**: Runs a code check using `go vet`.

```bash
make vet
```

- **coverage**: Runs tests with coverage and generates a report.

```bash
make coverage
```

- **build**: Builds the project's binary.

```bash
make build
```

- **run**: Runs the project's binary.

```bash
make run
```

- **swagger**: Generates Swagger documentation for the project.

```bash
make swagger
```
