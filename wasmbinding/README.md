# CosmWasm Support

This package contains CosmWasm integration points.

This package provides first class support for:

- Queries
  - Pairs
  - Pools

- Messages / Execution
  - LimitOrder

## Command line interface (CLI)

- Commands

```sh
  crescentd tx wasm -h
```

- Query

```sh
  crescentd query wasm -h
```

## Tests

This contains a few high level tests that `x/wasm` is properly
integrated.

Since the code tested is not in this repo, and we are just testing the
application integration (app.go), I figured this is the most suitable
location for it.