[![codecov](https://codecov.io/gh/tendermint/farming/branch/main/graph/badge.svg)](https://codecov.io/gh/tendermint/farming?branch=main)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/tendermint/farming)](https://pkg.go.dev/github.com/tendermint/farming)

# Farming Module

The farming module is a Cosmos SDK module that implements farming functionality, which provides farming rewards to participants called farmers. A primary use case is to use this module to provide incentives for liquidity pool investors for their pool participation. 

⚠ **Farming module v1 is in active development** ⚠ 
- see the [main](https://github.com/tendermint/farming/tree/main) branch for the latest 
- see [releases](https://github.com/tendermint/farming/releases) for the latest release

## Dependencies

If you haven't already, install Golang by following the [official docs](https://golang.org/doc/install). Make sure that your `GOPATH` and `GOBIN` environment variables are properly set up.

Requirement | Notes
----------- | -----------------
Go version  | Go1.16 or higher
Cosmos SDK  | v0.44.2 or higher

## Installation

```bash
# Use git to clone farming module source code and install `farmingd`
git clone https://github.com/tendermint/farming.git
cd farming
make install
```

## Getting Started

To get started to the project, visit the [TECHNICAL-SETUP.md](./TECHNICAL-SETUP.md) docs.

## Documentation

The farming module documentation is available in [docs](./docs) folder and technical specification is available in [specs](https://github.com/tendermint/farming/blob/main/x/farming/spec/README.md) folder. 

These are some of the documents that help you to quickly get you on board with the farming module.

- [How to bootstrap a local network with farming module](./docs/Tutorials/localnet)
- [How to use Command Line Interfaces](./docs/How-To/cli)
- [How to use gRPC-gateway REST Routes](./docs/How-To)
- [Demo for how to budget and farming modules](./docs/Tutorials/demo/budget_with_farming.md)

## Contributing

We welcome contributions from everyone. The [main](https://github.com/tendermint/farming/tree/main) branch contains the development version of the code. You can branch of from main and create a pull request, or maintain your own fork and submit a cross-repository pull request. If you're not sure where to start check out [CONTRIBUTING.md](./CONTRIBUTING.md) for our guidelines & policies for how we develop farming module. Thank you to all those who have contributed to farming module!
