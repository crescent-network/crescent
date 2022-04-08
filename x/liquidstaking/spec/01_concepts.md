<!-- order: 1 -->

# Concepts

## Liquid Staking Module

Liquidstaking module provides a way for delegators to benefit with greater yields at the same time that improves capital efficiency and decentralization. The module allows users to delegate their native token(BondDenom) that stakes for them without losing access to their funds. Users can liquid stake any amount of native token and receive a synthetic version of the original token called “bToken” as a staking representation. When delegators want to liquid unstake, the module burns the bToken and give them back the native token. Delegators must wait for the unbonding period. Technically, it is worth noting that the module unbonds the amount of bToken that values the native token at the mint rate that considers both accumulated rewards and slashing existence.

## Liquid Validators

Liquid validators are determined through governance process. They must be defined in governance parameter called WhitelistedValidators and as long as they comply the active condition (you can find more details about what active condition is in the next document), they become liquid validators. When users liquid stake or unstake, there is a proxy account that acts as a middleman. When users liquid stake, the amount of liquid staking is transferred to the proxy account which distributes amount to all active liquid validators in correspondence with their weight. When users liquid unstake, the proxy account undelegates weighted shares from the active liquid validators. Note that gas usage can increase as there are more liquid validators. 

## Liquid Governance

The module acknowledges voting power of all voters by aggregating the core activities that is related to the bToken. These are the following core activities:

- Balance of `bToken`
- Balance of `PoolCoin(s)` that includes `bToken`
- Farming position of `bToken`
- Farming position of `PoolCoin(s)` that include `bToken`

## Rebalancing

The module rebalances liquid tokens of active liquid validators by redelegating from one liquid validator to another. Some cases include when there is a change in whitelisted validators and a liquid validator gets slashed. Technically, it is worth noting that some redelegation may fail due to redelegation hopping restriction in the staking module of Cosmos SDK. In that case, the module retries at the beginning of next block until it gets resolved.

A redelegation object is created every time a redelegation occurs. To prevent "redelegation hopping" redelegations may not occur under the situation that:

- the (re)delegator already has another immature redelegation in progress with a destination to a validator (let's call it Validator X)
- and, the (re)delegator is attempting to create a new redelegation where the source validator for this new redelegation is Validator X.

## Restake

The module restakes amount to all active liquid validators that corresponds to their weight when an accumulated reward is over `RewardTrigger` value. 

## Unbonding Period

Liquid stakers who unbond their delegation must wait for the duration of the `UnbondingTime`. It is a chain-specific parameter. During the unbonding period, they are still exposed to being slashed for any liquid validator’s misbehavior.

## Slashing

A liquid validator must comply slashing rules of the slashing module in Cosmos SDK. They must keep up their liveness and stay away from any other infraction related attributes. If a liquid validator fails to comply the slashing rules, the module burns some amount of liquid tokens from all liquid validators. This results to having the value of bToken decreased. Therefore, it is crucial for the community to choose and elect the most secure and responsible liquid validators.
