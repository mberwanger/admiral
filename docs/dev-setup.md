# Admiral Development Environment Setup

This guide walks you through setting up a local development environment for the `admiral` project. Whether you're contributing code, testing features, or exploring the app, these steps will get you up and running.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Docker**: Required for running dependencies (e.g., databases, services) via Docker Compose.
  - Install: [Docker Desktop](https://www.docker.com/products/docker-desktop/) (includes Docker Compose).
  - Verify: `docker --version` and `docker compose version`.
- **Go**: Needed to build and run the backend.
  - Install: [Go Installation Guide](https://go.dev/doc/install).
  - Recommended version: 1.24 or later (check `go.mod` in the repo for specifics).
  - Verify: `go version`.

## Getting Started

1. **Set Up Dependencies with Docker Compose**

- The project uses Docker Compose to provision external services (e.g., database).
- Run the following command to start these services in the background:

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml up -d
```

- Verify services are running:

```bash
docker ps
```

_Note: Check `deploy/docker-compose/docker-compose.yaml` for details on services (e.g., ports, credentials)._

2. **Run Database Migrations**

- The backend requires a database schema. Use the Go CLI to apply migrations:

```bash
go run main.go migrate --config config.yaml --env .env.dev
```

- Ensure `config.yaml` and `.env.dev` exist in the root directory:
  - `config.yaml`: Backend configuration (e.g., DB connection details).
  - `.env.dev`: Environment variables for development (e.g., DB_HOST=localhost).

3. **Start the Development Environment**

- Use the make dev target to run both the frontend and backend simultaneously:

```bash
make dev
```

_This launches the full stack in development mode with hot reloading (if configured)._

## Docker Compose Components

The `docker-compose.yaml` file in `deploy/docker-compose/` sets up the following services:

### Keycloak (Local Identity Provider)

- **Purpose:** Provides authentication and user management via an OpenID Connect-compatible IdP.
- **Port:** 9090
- **Admin Console:** http://localhost:9090/admin/master/console/
- **Admin Credentials:**
  - Username: `admiral`
  - Password: `secret`
- **Realm:** A realm named `admiral` is pre-configured.
- **Users in `admiral` Realm:**
  - **Admin User:**
    - Username: `admin`
    - Password: `admin`
    - Role: Administrator privileges.
  - **Demo User:**
    - Username: `demo`
    - Password: `demo`
    - Role: Typical user for testing.

_Notes: Use the admin console to manage users, roles, or clients as needed._

### Mailpit (Email Sink)

- **Purpose:** Captures and displays outgoing emails in development, preventing real emails from being sent.
- **Web Interface:** http://localhost:8025/
- **Port:** 8025 (web UI), 1025 (SMTP)

_Notes: Check the UI to view emails triggered by the app (e.g., registration, password resets)._

### PostgreSQL Database

- **Purpose:** Persistent storage for the application.
- **Port:** 5432
- **Database Name:** admiral
- **Credentials:**
  - Username: `admiral`
  - Password: `secret`

_Notes: The schema is applied via migrations in Step 2. Connect manually with `psql -h localhost -U admiral -d admiral` if needed (password: `secret`)._

## Running Components Separately (Optional)

If you prefer to work on the backend or frontend independently:

- **Backend Only**

```bash
make server-dev
```

_Runs the Go server, typically on localhost:8080 (check config.yaml for port)._

- **Frontend Only**

```bash
make web-dev
```

_Runs the web app, often on localhost:8888 (confirm in frontend config or Makefile)._

## Useful Commands

- **List All Make Targets**

```bash
make help
```

_Displays available make commands and their purposes._

- **Stop Docker Services**

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml down
```

## Troubleshooting

- Docker Errors: Ensure Docker is running and you have permissions (e.g., add your user to the docker group on Linux).
- Go Command Fails: Verify Go modules are downloaded (go mod download).
- Port Conflicts: Check if ports (e.g., 9090, 5432, 8025) are in use (lsof -i :9090) and adjust configs if needed.

## Next Steps

- Explore the codebase.
- Check README.md or other docs in the repo for app-specific details.
- Submit issues or PRs to https://github.com/mberwanger/admiral!
