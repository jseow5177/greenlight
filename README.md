# Greenlight API

## Introduction
Greenlist is a JSON API built with Go for retrieving and managing information about movies.

## Project Structure
- `bin` - Contain compiled application binaries, ready for deployment to a production server.
- `cmd/api` - Contain application-specifc code for the Greenlight API application.
- `internal` - Contain reusable packages used by the API. For example, code for interacting with database, data validation, and so on.
- `internal/data` - Contain all the custom data types used in this application.
- `migrations` - Contain the SQL migration files.
- `remote` - Contain the configuration files and setup scripts for production server.
- `go.mod` file - Declare project dependencies, versions and module path.
- `Makefile` - Contain recipes for automating common administrative tasks (Go code auditing, building binaries and execute database migrations).

## API Routes
| Method | Route | Description |
| ------ | ----- | ----------- |
| GET    | /v1/healthcheck | Show application health and version information |
| GET    | /v1/movies      | Show the details of all movies |
| POST   | /v1/movies      | Create a new movie |
| GET    | /v1/movies/:id  | Show the details of a specific movie |
| PUT    | /v1/movies/:id  | Update the details of a specific movie |
| DELETE | /v1/movies/:id  | Delete a specific movie |