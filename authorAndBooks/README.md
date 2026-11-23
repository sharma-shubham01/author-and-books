# authorAndBooks

A REST API in Go for managing **authors** and their **books**, backed by MySQL.

It is split into layers that each handle one job: the HTTP handlers parse requests, the services apply the business rules, and the stores run the SQL. Each layer depends on interfaces, so every layer is unit-tested in isolation.

## Contents

- [Tech stack](#tech-stack)
- [Project structure](#project-structure)
- [Getting started](#getting-started)
- [Configuration](#configuration)
- [API reference](#api-reference)
- [Business rules](#business-rules)
- [Error handling](#error-handling)
- [Testing](#testing)
- [Linting](#linting)

## Tech stack

| Concern  | Choice                                                          |
| -------- | --------------------------------------------------------------- |
| Language | Go 1.18+                                                        |
| Router   | [gorilla/mux](https://github.com/gorilla/mux)                   |
| Database | MySQL, via [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) |
| Testing  | [testify](https://github.com/stretchr/testify), [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock) |
| Linting  | [golangci-lint](https://golangci-lint.run/) (`.golangci.yaml`)  |

## Project structure

```
.
├── main.go              # Wiring: DB connection, stores, services, handlers, routes
├── schema.sql           # MySQL schema (Author and Book tables)
├── apperrors/           # Domain errors (Validation, NotFound, Conflict)
├── database/            # MySQL config (from env) and connection
├── model/               # Author and Book types
├── handler/
│   ├── rest/            # Shared JSON / path-param / error-response helpers
│   ├── author/          # /author HTTP handlers
│   └── book/            # /book HTTP handlers
├── service/             # Business rules and validation
└── store/               # MySQL persistence and the transaction helper
```

A request goes through these layers:

```
HTTP request → handler → service → store → MySQL
                  ↑         │
                  └─ apperrors mapped to HTTP status codes
```

- **Handlers** decode JSON, parse path and query parameters, and turn errors into HTTP responses.
- **Services** validate input, enforce uniqueness and existence rules, and decide when to use a transaction.
- **Stores** run SQL only. They do no business validation. `store.Transactor` runs a function against author and book stores bound to a single `*sql.Tx`.

## Getting started

### Prerequisites

- Go 1.18 or newer
- A running MySQL server

### 1. Create the database schema

The schema is in [`schema.sql`](schema.sql). It creates the `test` database and the `Author` and `Book` tables:

```sh
mysql -u root -p < schema.sql
```

You can run it more than once, because it uses `IF NOT EXISTS`. If you use a database other than `test`, change the `CREATE DATABASE` / `USE` lines and set `DB_NAME` to match.

> Tip: to start MySQL quickly with Docker:
>
> ```sh
> docker run --name authors-mysql -e MYSQL_ROOT_PASSWORD=secret -e MYSQL_DATABASE=test -p 3306:3306 -d mysql:8
> ```

### 2. Run the server

```sh
git clone <repo-url>
cd authorAndBooks
go mod download

DB_PASSWORD=secret go run .
```

The server listens on **`:8000`**:

```
server started at :8000
```

## Configuration

The database connection is configured through environment variables (see `database/db.go`):

| Variable      | Default     | Description       |
| ------------- | ----------- | ----------------- |
| `DB_USER`     | `root`      | MySQL user        |
| `DB_PASSWORD` | *(empty)*   | MySQL password    |
| `DB_HOST`     | `127.0.0.1` | MySQL host        |
| `DB_PORT`     | `3306`      | MySQL port        |
| `DB_NAME`     | `test`      | Database name     |

The server pings the database at startup and exits if it can't connect.

## API reference

Base URL: `http://localhost:8000`. All request and response bodies are JSON.

### Data models

**Author**

```json
{
  "ID": 1,
  "first_name": "Ruskin",
  "last_name": "Bond",
  "dob": "19/05/1934",
  "pen_name": "Ruskin"
}
```

**Book**

```json
{
  "ID": 1,
  "title": "The Blue Umbrella",
  "author": { "ID": 1, "first_name": "Ruskin", "last_name": "Bond", "dob": "19/05/1934", "pen_name": "Ruskin" },
  "publication": "Penguin",
  "published_date": "12/08/1980"
}
```

### Endpoints

| Method   | Path           | Description                                   | Success |
| -------- | -------------- | --------------------------------------------- | ------- |
| `GET`    | `/book`        | List books (optional filters, see below)      | `200`   |
| `GET`    | `/book/{id}`   | Get one book, including its author            | `200`   |
| `POST`   | `/book`        | Create a book (and its author if needed)      | `201`   |
| `PUT`    | `/book/{id}`   | Replace a book                                | `200`   |
| `DELETE` | `/book/{id}`   | Delete a book                                 | `204`   |
| `POST`   | `/author`      | Create an author                              | `201`   |
| `PUT`    | `/author/{id}` | Replace an author                             | `200`   |
| `DELETE` | `/author/{id}` | Delete an author **and all of their books**   | `204`   |

`{id}` must be a positive integer.

#### `GET /book`

| Query param     | Type    | Description                                                         |
| --------------- | ------- | ------------------------------------------------------------------- |
| `title`         | string  | Return only books with this exact title                             |
| `includeAuthor` | boolean | `true` returns full author details; by default only `author.ID` is set |

```sh
curl "http://localhost:8000/book?title=The%20Blue%20Umbrella&includeAuthor=true"
```

#### `POST /book`

There are two ways to link a book to an author:

1. **Existing author:** set `author.ID`. That author must already exist.
2. **Inline author:** leave out `author.ID` and give the author's details. If an author with exactly the same details exists, that author is reused. Otherwise a new author is created. The author and the book are written in one transaction.

```sh
curl -X POST http://localhost:8000/book \
  -H "Content-Type: application/json" \
  -d '{
        "title": "The Blue Umbrella",
        "publication": "Penguin",
        "published_date": "12/08/1980",
        "author": { "first_name": "Ruskin", "last_name": "Bond", "dob": "19/05/1934", "pen_name": "Ruskin" }
      }'
```

#### `PUT /book/{id}`

Replaces the whole book. `author.ID` is required and must refer to an existing author.

```sh
curl -X PUT http://localhost:8000/book/1 \
  -H "Content-Type: application/json" \
  -d '{ "title": "The Blue Umbrella", "publication": "Scholastic", "published_date": "1/1/1981", "author": { "ID": 1 } }'
```

#### `POST /author`

```sh
curl -X POST http://localhost:8000/author \
  -H "Content-Type: application/json" \
  -d '{ "first_name": "Ruskin", "last_name": "Bond", "dob": "19/05/1934", "pen_name": "Ruskin" }'
```

#### `PUT /author/{id}`

Replaces the whole author. Every field is required.

#### `DELETE /author/{id}` / `DELETE /book/{id}`

```sh
curl -X DELETE http://localhost:8000/author/1
```

Deleting an author also deletes all of their books, in one transaction.

## Business rules

**Authors**

- Creating an author requires `first_name` and `dob`.
- Updating an author requires all fields: `first_name`, `last_name`, `dob`, `pen_name`.
- Two authors can't have exactly the same details. A duplicate returns `409 Conflict`.

**Books**

- `title` is required.
- `publication` must be one of `Penguin`, `Scholastic` or `Arihanth`.
- `published_date` must use the `d/m/yyyy` format (zero-padding is optional). It can't be earlier than **1880** or later than today.
- An author can't have two books with the same title. A duplicate returns `409 Conflict`.
- When a request has several validation problems, they are all reported together, separated by `; `.

## Error handling

Errors are returned as JSON:

```json
{ "error": "missing required fields: first_name, dob" }
```

| Domain error             | HTTP status                 | Example                                     |
| ------------------------ | --------------------------- | ------------------------------------------- |
| `apperrors.Validation`   | `400 Bad Request`           | invalid body, bad id, failed validation     |
| `apperrors.NotFound`     | `404 Not Found`             | `book with id 42 not found`                 |
| `apperrors.Conflict`     | `409 Conflict`              | `author already exists`                     |
| anything else            | `500 Internal Server Error` | details are logged, not returned to the client |

Request bodies are limited to 1 MiB.

## Testing

Each layer is tested in isolation:

- **Handlers:** `httptest` with mocked services
- **Services:** hand-written fake stores (`service/fakes_test.go`)
- **Stores:** `go-sqlmock` checks the exact SQL that runs

```sh
go test ./...                 # run all tests
go test -cover ./...          # with coverage
go test -race ./...           # with the race detector
```

The tests don't need a running database.

## Linting

The linter configuration is in `.golangci.yaml`:

```sh
golangci-lint run
```
