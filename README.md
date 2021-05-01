# Greenlight API

## Introduction
Greenlist is a JSON API built with Go for retrieving and managing information about movies.

## Project Structure
- `bin` - Contain compiled application binaries, ready for deployment to a production server.
- `cmd/api` - Contain application-specifc code for the Greenlight API application.
- `internal` - Contain reusable packages used by the API. For example, code for interacting with database, data validation, and so on.
- `migrations` - Contain the SQL migration files.
- `remote` - Contain the configuration files and setup scripts for production server.
- `go.mod` file - Declare project dependencies, versions and module path.
- `Makefile` - Contain recipes for automating common administrative tasks (Go code auditing, building binaries and execute database migrations).

## API Routes
| Method | Route | Description |
| ------ | ----- | ----------- |
| GET    | /api/healthcheck | Show application health and version information |