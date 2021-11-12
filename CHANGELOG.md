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

## [Unreleased] v1.0.0 - 2021-11-26

* [\#64](https://github.com/tendermint/farming/pull/64) docs: improve documentation for audit release

* [\#187](https://github.com/tendermint/farming/pull/187) edit farming.proto
* [\#196](https://github.com/tendermint/farming/pull/196) doc: fix to be broken links by renaming branch name
* [\#189](https://github.com/tendermint/farming/pull/189) feat: Create CODEOWNERS file and update contributing.md
* [\#198](https://github.com/tendermint/farming/pull/198) feat: change naming for consistency in public proposals
* [\#191](https://github.com/tendermint/farming/pull/191) docs: edit farming spec to improve documentation quality
* [\#201](https://github.com/tendermint/farming/pull/201) build: bump liquidity to v1.4.1
* [\#186](https://github.com/tendermint/farming/pull/186) feat: add and refactor invariant checks
* [\#169](https://github.com/tendermint/farming/pull/169) tests: add detailed and randomized cases for public plan proposals simulation
* [\#205](https://github.com/tendermint/farming/pull/205) fix: change the initial epoch to 1
* [\#210](https://github.com/tendermint/farming/pull/210) docs: edit budget_with_farming.md demo text
* [\#214](https://github.com/tendermint/farming/pull/214) fix: use default String implementation for types
* [\#218](https://github.com/tendermint/farming/pull/218) test: implement test for PositiveTotalStakingsAmountInvariant
* [\#209](https://github.com/tendermint/farming/pull/209) fix: resolve TODO comments
* [\#202](https://github.com/tendermint/farming/pull/202) fix: apply suggestions from the audit review
* [\#223](https://github.com/tendermint/farming/pull/223) fix: remove PlanTerminationStatusInvariant
* [\#216](https://github.com/tendermint/farming/pull/216) test: add tests for gov proposal
* [\#225](https://github.com/tendermint/farming/pull/225) fix: add plan type validation for public plan proposal
* [\#206](https://github.com/tendermint/farming/pull/206) build: bump cosmos-sdk version to v0.44.3
* [\#228](https://github.com/tendermint/farming/pull/228) test: add test cases for reward allocation of ratio plan

## [v0.1.2](https://github.com/tendermint/farming/releases/tag/v0.1.2) - 2021-10-18

* [\#181](https://github.com/tendermint/farming/pull/181) fix: emit rewards_withdrawn only when positive amount is withdrawn
* [\#180](https://github.com/tendermint/farming/pull/180) fix: withdraw rewards only when staked amount changes
* [\#178](https://github.com/tendermint/farming/pull/178) docs: improve documentation for audit release
* [\#177](https://github.com/tendermint/farming/pull/177) feat: bump budget to v0.1.1

## [v0.1.1](https://github.com/tendermint/farming/releases/tag/v0.1.1) - 2021-10-15

* [\#135](https://github.com/tendermint/farming/pull/135) fix: Fix comparison bug when allocating all balances of the farming pool
* [\#133](https://github.com/tendermint/farming/pull/133) feat: fix update public plan proposal & disable codecov patch
* [\#140](https://github.com/tendermint/farming/pull/140) docs: add demo for budget with farming
* [\#141](https://github.com/tendermint/farming/pull/141) docs: update spec documentation and add events
* [\#145](https://github.com/tendermint/farming/pull/145) fix: remove minter perm and add liquidity module
* [\#143](https://github.com/tendermint/farming/pull/143) test: add tests for HistoricalRewards
* [\#132](https://github.com/tendermint/farming/pull/132) test: Add genesis tests
* [\#146](https://github.com/tendermint/farming/pull/146) test: add CLI query tests
* [\#154](https://github.com/tendermint/farming/pull/154) test: add import and export simulation
* [\#150](https://github.com/tendermint/farming/pull/150) fix: duplicate value of the name field in plan
* [\#157](https://github.com/tendermint/farming/pull/157) test: add tests for types
* [\#156](https://github.com/tendermint/farming/pull/156) feat: add examples of multiple coins for command-line interfaces
* [\#162](https://github.com/tendermint/farming/pull/162) fix: terminate plan when deleting public plan proposal
* [\#166](https://github.com/tendermint/farming/pull/166) fix: emit rewards_withdrawn event
* [\#160](https://github.com/tendermint/farming/pull/160) fix: do not allow empty plan name
* [\#168](https://github.com/tendermint/farming/pull/168) chore: bump Cosmos SDK version to v0.44.2
* [\#158](https://github.com/tendermint/farming/pull/158) docs: add more clear and description comments
* [\#167](https://github.com/tendermint/farming/pull/167) fix: refine sentinel errors
* [\#165](https://github.com/tendermint/farming/pull/165) docs: add detailed descriptions and examples to swagger specification
* [\#175](https://github.com/tendermint/farming/pull/175) feat: impose DelayedStakingGasFee when staking

## [v0.1.0](https://github.com/tendermint/farming/releases/tag/v0.1.0) - 2021-09-17