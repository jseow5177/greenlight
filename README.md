# Greenlight API

## Introduction
Greenlist is a JSON API built with Go for retrieving and managing information about movies.

This project is from the book <a href="https://lets-go-further.alexedwards.net/" target="_blank">Let's Go Further!</a> by Alex Edwards.

## Software Versions
1. PostgreSQL v13.2
2. Go v1.16.3 windows/amd64
3. migrate v4.14.1
4. curl

## Getting started

TODO: Add instructions to setup development environment.

## Project Structure
- `bin` - Contain compiled application binaries, ready for deployment to a production server.
- `cmd/api` - Contain application-specifc code for the Greenlight API application.
- `internal` - Contain reusable packages used by the API. For example, code for interacting with database, data validation, and so on.
- `internal/data` - Contain all the custom data types used in this application.
- `migrations` - Contain the SQL migration files.
- `remote` - Contain the configuration files and setup scripts for production server.
- `go.mod` file - Declare project dependencies, versions and module path.
- `Makefile` - Contain recipes for automating common administrative tasks (Go code auditing, building binaries and execute database migrations).
- `bash` - Contain bash scripts that execute curl commands to test handlers

## API Routes
| Method | Route | Description |
| ------ | ----- | ----------- |
| GET    | /v1/healthcheck | Show application health and version information |
| GET    | /v1/movies      | Show the details of all movies |
| POST   | /v1/movies      | Create a new movie |
| GET    | /v1/movies/:id  | Show the details of a specific movie |
| PUT    | /v1/movies/:id  | Update the details of a specific movie |
| DELETE | /v1/movies/:id  | Delete a specific movie |

## Database Pool Configuration

Go's `sql.DB` connection pool contains two types of connections - 'in-use' and 'idle' connections.

An 'in-use' connection is one where it is used to perform a database task such as performing a query. Once the task is done, the connection will be marked as 'idle'.

We can configure the behavior of Go's connection pool with the following four settings.

| Database Setting | Detail | Default Value | Application Setting |
| ------ | ------ | ------ | ----- | 
| MaxIdleConns | The maximum number of idle connections in the pool. | 2 |  25 | 
| MaxOpenConns | The maximum number of open connections (in-use + idle) in the pool. | Unlimited | 25 |
| ConnMaxLifetime | The maximum length of time that a connection can be reused for. | Unlimited | Default |
| ConnMaxIdleTime | The maximum length of time that a connection can be idle. | Unlimited | 15 mins |

## Database Models

### Movie

| Field | Description | 
| ----- | ------ | 
| id | Unique identifier |
| title | Title of movie |
| year | Movie release year |
| runtime | Movie runtime in minutes |
| genres | Movies genres |
| version | The version of movie data. Incremented on each update |

## Filtering, Sorting and Pagination

The API `GET /v1/movies` supports query parameters that implement filtering, sorting, and pagination.

### Filtering

This application uses reductive filtering and performs a basic full-text partial searches.

```
// List all movies
/v1/movies

// List movies where the title is a case-insenstive exact match for 'black panther'
/v1/movies?title=black+panther

// List movies where the genres includes 'adventure'
/v1/movies?genres=adventure

// List movies where the title is a case-insensitive exact match for 'moana' AND the
// genres include both 'animation' AND 'adventure'
/v1/movies?title=moana&genres=animation,adventure
```

## API Test Scripts

Run the following scripts in sequence.

Run all 'up' migration files

`sh ./bash/migrateup.sh`

Test handler to create movies

`sh ./bash/createmovie.sh`

Test handler to show movies

`sh ./bash/showmovie.sh`

Test handler to update movies

`sh ./bash/updatemovie.sh`

Test handler to delete movies

`sh ./bash/deletemovie.sh`

Run all 'down' migration files

`sh ./bash/migratedown.sh`




