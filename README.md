<p>&nbsp;</p>
<p align="center">

<img src="assets/banner.png" width=700 height=280>

</p>

<p align="center">
Expanding DeFi capabilities through InterBlockchain Technology<br/><br/>

<a href="https://pkg.go.dev/github.com/crescent-network/crescent">
    <img src="https://pkg.go.dev/badge/github.com/crescent-network/crescent">
</a>
<a href="https://codecov.io/gh/crescent-network/crescent">
    <img src="https://codecov.io/gh/crescent-network/crescent/branch/main/graph/badge.svg">
</a>
<img src="https://github.com/crescent-network/crescent/actions/workflows/test.yml/badge.svg">
</p>


## What is Crescent Network?

Crescent Network is a DeFi Hub in Cosmos ecosystem with a goal of empowering users’ digital assets for maximizing their financial returns while managing associated risks in the most efficient way by providing innovative and sophisticated DeFi products. In the base layer, Crescent core has the following unique characteristics.

- Hybrid DEX: a combination of Automated Market Maker (AMM) and Order Book models.
- Ranged Pool: next generation Automated Market Maker that increases capital efficiency. Liquidity is allocated within a predefined price range.
- Batch Execution : all deposits, withdrawals, and orders are accumulated in a batch and they are fairly executed at the same time.
- Novel DeFi products are on the way!

## Installation

### Use binaries

Download a pre-built binary for your operating system. You can find the latest binaries in this [releases](https://github.com/crescent-network/crescent/releases) page.

### Build from source

**Step 1. Install Golang**

Go version [1.18](https://go.dev/doc/go1.18) or higher is required.

If you haven't already, install Go by following the installation guide in [the official docs](https://golang.org/doc/install). Make sure that your `GOPATH` and `GOBIN` environment variables are properly set up.

**Step 2. Get source code**

Use `git` to retrieve Crescent Core from [the official repository](https://github.com/crescent-network/crescent) and checkout latest release, which will install the `crescentd` binary.

```bash
git clone https://github.com/crescent-network/crescent.git
cd crescent && git checkout release/v5.0.x
make install
```

**Step 3. Verify your installation**

Verify the commit hash to see if you have installed `crescentd` correctly.

```bash
crescentd version --long
```

## Dependency

Crescent core uses a fork of [cosmos-sdk](https://github.com/crescent-network/cosmos-sdk) and [ibc-go](https://github.com/crescent-network/ibc-go). If you would like to know which ones customized from the original `cosmos-sdk` and `ibc-go`, please reference the release notes in the respective repository.

| Requirement         | Notes                |
|---------------------|----------------------|
| cosmos-sdk (forked) | v1.2.0-sdk-0.45.10   |
| ibc-go (forked)     | v3.4.0-crescent-v4-2 |

## Documentation

The documentation is available in [docs](docs) directory. If you are a developer interested in technical specification, go to each `x/{module}`'s `spec` directory.

* [Crescent Official Docs](https://docs.crescent.network/)
* [Swagger API Docs](https://app.swaggerhub.com/apis-docs/crescent/crescent)

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
