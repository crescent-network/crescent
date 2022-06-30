<p>&nbsp;</p>
<p align="center">

<img src="crescent_core_image.png" width=700 height=280>

</p>

<p align="center">
Crescent Core - Expanding DeFi capabilities through InterBlockchain Technology<br/><br/>

<a href="https://pkg.go.dev/github.com/crescent-network/crescent">
    <img src="https://pkg.go.dev/badge/github.com/crescent-network/crescent">
</a>
<a href="https://codecov.io/gh/crescent-network/crescent">
    <img src="https://codecov.io/gh/crescent-network/crescent/branch/main/graph/badge.svg">
</a>
<img src="https://github.com/crescent-network/crescent/actions/workflows/test.yml/badge.svg">
</p>


## What is Crescent?

Crescent is a DeFi Hub that provides innovative and powerful tools that empower users’ digital assets for maximizing their financial returns while managing associated risks in the most efficient way. In the base layer, Crescent core has a DEX functionality equipped with several unique characteristics.

- A Hybrid DEX : a hybrid system of orderbook and AMM
- Tick System : standardization of order price
- Batch Execution : all orders included in same block are fairly executed
- Ranged Pools : liquidity pools providing liquidity within predefined price range
- Optimized Liquidity Incentives Strategy
- Synergy with Crescent Boost

## Installation

### Use binaries

This is the easiest way to get started. Download a pre-built binary for your operating system. You can find the latest binaries on the [releases](https://github.com/crescent-network/crescent/releases) page.

### Build from source

**Step 1. Install Golang**

Go version [1.17](https://go.dev/doc/go1.17) or higher is required for Crescent Core.

If you haven't already, install Go by following the [official docs](https://golang.org/doc/install). Make sure that your `GOPATH` and `GOBIN` environment variables are properly set up.

**Step 2. Get Crescent Core source code**

Use `git` to retrieve Crescent Core from the [official repo](https://github.com/crescent-network/crescent/) and checkout the `release/v2.0.x` branch. This branch contains the latest release, which will install the `crescentd` binary.

```bash
git clone https://github.com/crescent-network/crescent.git
cd crescent && git checkout release/v2.0.x
make install
```

**Step 3. Verify your installation**

Verify that you have installed `crescentd` successfully by running the following command:

```bash
crescentd version --long
```

## Dependency

Crescent Core uses a customized Cosmos SDK. Please check the differences on [here](https://github.com/crescent-network/cosmos-sdk/compare/v0.45.3...v1.1.0-sdk-0.45.3).

| Requirement           | Notes             |
|-----------------------|-------------------|
| Go version            | Go1.17 or higher  |
| customized cosmos-sdk | v1.1.0-sdk-0.45.3 |

## Documentation

The documentation is available in [docs](docs) directory. If you are a developer interested in technical specification, see inside each `x/{module}`'s `spec` directory.

* [Crescent Official Docs](https://docs.crescent.network/)
* [Swagger API Docs](https://app.swaggerhub.com/apis-docs/crescent/crescent/2.0.0)

## Community

* [Official Website](https://crescent.network/)
* [Medium Blog](https://crescentnetwork.medium.com/)
* [Discord](https://discord.com/invite/vmjfqHy4UA)
* [Telegram](https://t.me/+5lJ33oeqV2QwYzQ1)
* [Twitter](https://twitter.com/CrescentHub)

## Contributing

Crescent is a public and open-source blockchain protocol. We welcome contributions from everyone. If you are interested in contributing to Crescent Core, please review our [CONTRIBUTING](CONTRIBUTING.md) guide. Thank you to all those who have contributed to Crescent Core.

## License

This software is licensed under the Apache 2.0 license. Read more about it [here](LICENSE).
