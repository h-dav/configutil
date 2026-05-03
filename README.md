# configutil

Package `configutil` populates a Go struct from config sources.

[![Go Report Card](https://goreportcard.com/badge/github.com/h-dav/configutil)](https://goreportcard.com/report/github.com/h-dav/configutil)

## Installation

```bash
go get github.com/h-dav/configutil
```

## Quick Start

```go
type Config struct {
    Service string `config:"SERVICE,required"`
    Port    int    `config:"PORT,default=8080"`
}

var cfg Config
if err := configutil.Set(&cfg); err != nil {
    log.Fatal(err)
}
```

## Features

### Options

| Option                         | Description                        |
|--------------------------------|------------------------------------|
| `WithFilepath("config.env")`   | Load values from a `.env` file.    |

### Struct Tags

Format: `config:"NAME[,option,...]"`

| Option            | Description                                          |
|-------------------|------------------------------------------------------|
| `required`        | Error if no source explicitly sets the field. A `default` does not satisfy `required` — use one or the other. |
| `default=<value>` | Fallback value when no source provides one.          |
| `prefix=<prefix>` | Namespace for nested structs.                        |

### Text Replacement

Use `${VAR}` inside values to reference any key that has been discovered from any source (env vars, flags, or other file values):

```env
HOST=localhost
URL=http://${HOST}:8080
```

References are resolved against the fully merged source map, so a `.env` file value can reference an environment variable:

```env
# .env file — references the HOST env var set in the shell
URL=http://${HOST}:8080
```

Only identifiers matching `[A-Za-z_][A-Za-z0-9_]*` are treated as replacement targets. Any `${...}` pattern that does not match a valid identifier is passed through as literal text.

### Nested Structs

```go
type Config struct {
    Server struct {
        Port int `config:"PORT"`
    } `config:",prefix=SERVER_"`
}
// Reads SERVER_PORT from sources.
```

## Precedence

Sources are evaluated in order. Later sources overwrite earlier ones.

| Priority | Source                  | Notes                                              |
|----------|------------------------|----------------------------------------------------|
| 1 (low)  | Defaults               | `default=...` in struct tags.                      |
| 2        | `.env` files           | Loaded via `WithFilepath()`.                       |
| 3        | Environment variables  | Process environment.                               |
| 4 (high) | Flags                  | App must call `flag.Parse()` before `Set()`.       |

## Error Handling

All errors support `errors.Is` and `errors.As`:

```go
err := configutil.Set(&cfg)

if errors.Is(err, configutil.ErrRequired) {
    // handle missing required field
}

var convErr *configutil.FieldConversionError
if errors.As(err, &convErr) {
    fmt.Printf("field %s: %v\n", convErr.FieldName, convErr.Err)
}
```

### Sentinel Errors

| Sentinel           | Meaning                                       |
|--------------------|-----------------------------------------------|
| `ErrInvalidConfig` | Argument is not a pointer to a struct.         |
| `ErrUnsupported`   | Field type not supported.                      |
| `ErrRequired`      | Required field has no value.                   |
| `ErrFile`          | File open/read/extension error.                |
| `ErrParse`         | Syntax error in `.env` file.                   |
| `ErrConversion`    | String-to-type conversion failed.              |
| `ErrReplacement`   | `${VAR}` references an unknown key.            |
| `ErrTag`           | Malformed struct tag.                          |

## Benchmarking

```bash
go test -bench=. -benchmem ./...
```
