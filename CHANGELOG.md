<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking Protobuf, gRPC and REST routes used by end-users.
"CLI Breaking" for breaking CLI commands.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.
Ref: https://keepachangelog.com/en/1.0.0/
-->
<!-- markdown-link-check-disable -->

# Changelog

## [Unreleased]

## [v1.0.0] - 2022-04-12

### Features

* (crescentd) feat: add `x/liquidity` module
* (crescentd) feat: add `x/liquidstaking` module
* (crescentd) feat: add `x/farming` module
* (crescentd) feat: add `x/mint`(constant inflation) module
* (crescentd) feat: add `x/claim` module
* (sdk) Crescent Core uses a customized Cosmos SDK [v1.0.2-sdk-0.44.5](https://github.com/crescent-network/cosmos-sdk/releases/v1.0.2-sdk-0.44.5). Please check the differences on [here](https://github.com/crescent-network/cosmos-sdk/compare/v0.44.5...v1.0.2-sdk-0.44.5).
  * `x/staking` fix: allow delegate only spendable coins
  * `x/gov` feat: add additional voting powers hook on tally (liquid governance)
  * `x/vesting` feat: periodic vesting msg
  * `x/bank` feat: Add dynamic blockedAddrs
  
[Unreleased]: https://github.com/crescent-network/crescent/compare/v1.0.0...HEAD
[v1.0.0]: https://github.com/crescent-network/crescent/releases/tag/v1.0.0