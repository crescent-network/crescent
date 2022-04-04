<!--
order: 1
-->

# Concepts

## The Minting Mechanism: Constant Inflation

Unlike the `mint` module in Cosmos SDK that allows for a flexible (dynamic) inflation rate determined by market demand targeting a particular bonded-stake ratio, this `mint` module is cutomized to use a constant inflation rate. The module mints in relative to the block time with the pre-defined inflation schedule in params. It is possible that the actual minted amount for the schedule is less than the pre-defined inflation schedule amount due to the block time delay and decimal loss.

