# Farming Module

A farming module is a Cosmos SDK module that implements farming functionality, which provides farming rewards to participants called farmers. One use case is to use this module to provide incentives for liquidity pool investors for their pool participation. 

⚠ **Farming module v1 is in active development - see "master" branch for the v1 mainnet version** ⚠

## Installation
### Requirements

If you haven't already, install Golang by following the [official docs](https://golang.org/doc/install). Make sure that your `GOPATH` and `GOBIN` environment variables are properly set up.

Requirement | Notes
----------- | -----------------
Go version  | Go1.15 or higher
Cosmos SDK  | v0.44.0 or higher

### Get Farming Module source code

```bash
git clone https://github.com/tendermint/farming.git
cd farming
make install
```

## Development

### Test

```bash
make test-all
```

### Setup local testnet using script

```bash
# This script bootstraps a single local testnet.
# Note that config, data, and keys are created in the ./data/localnet folder and
# RPC, GRPC, and REST ports are all open.
$ make localnet
```

## Resources

- [Documentation about Command-line and REST API Interfaces](./docs/How-To/client.md)
- [MVP Demo](./docs/Tutorials/demo.md)