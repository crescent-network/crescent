<!-- order: 0 title: Liquid Staking Overview parent: title: "liquidstaking" -->

# `liquidstaking`

## Abstract

This document specifies the `liquidstaking` module. Proof of Stake consensus mechanism requires validators to lock their coins in blockchain network to have a change of validating next blocks and delegators delegate their tokens to their choice of a validator to receive rewards. If they want to unbond their delegations, then they generally have to wait for the unbonding period, which is usually 21 days in cosmos ecosystem. The `liquidstaking` module solves that problem by providing a synthetic version of the native token when they liquid stake their coins. This benefits stakers with greater yields and improve capital efficieny.

## Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[State Transitions](03_state_transitions.md)**
4. **[Messages](04_messages.md)**
5. **[Begin-Block](05_begin_block.md)**
6. **[Hooks](06_hooks.md)**
7. **[Events](07_events.md)**
8. **[Parameters](08_params.md)**
