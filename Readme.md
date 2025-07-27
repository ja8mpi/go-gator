# gator CLI

`gator` is a CLI tool that requires **PostgreSQL** and **Go** to be installed to run. It helps you manage feeds and user accounts easily.

## Prerequisites

- [PostgreSQL](https://www.postgresql.org/download/) installed and running on your machine.
- [Go 1.20+](https://go.dev/dl/) installed.

## Installation

To install the `gator` CLI globally, run this command:

```bash
go install github.com/yourusername/go-gator@latest
```

Replace `github.com/yourusername/go-gator` with the actual repository path.

This will compile gator and place the binary in your Go bin directory (usually `$GOPATH/bin`). Make sure that directory is in your system PATH so you can run gator from anywhere.

## Configuration

Before running gator, create a `config.json` file in the root directory of the project with your PostgreSQL connection info, for example:

```json
{
    "db_host": "localhost",
    "db_port": 5432,
    "db_user": "your_pg_user",
    "db_password": "your_pg_password",
    "db_name": "gator_db"
}
```

Make sure the database `gator_db` exists and is accessible with the credentials provided.

## Running the Program

For development, you can run the program using:

```bash
go run .
```

But this is only for development.

For production, run the installed CLI binary directly:

```bash
gator <command> [arguments]
```

Because Go programs are statically compiled binaries, after building or installing, you do not need Go installed to run the program.

## Common Commands

Here are a few example commands you can run with gator:

Register a new user:
```bash
gator register alice
```

Login as a user:
```bash
gator login alice
```

Add a new feed:
```bash
gator addfeed "My Blog" https://myblog.com/rss.xml
```

List feeds:
```bash
gator listfeeds
```

## GitHub Repository

You can find the source code and updates at:

https://github.com/yourusername/go-gator

Replace with your actual GitHub username and repo name.

## Summary

1. Install PostgreSQL and Go.
2. Install gator with `go install`.
3. Create `config.json` with your DB credentials.
4. Use `go run .` for development.
5. Use the installed `gator` binary for production commands.
