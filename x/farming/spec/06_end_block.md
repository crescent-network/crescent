<!-- order: 6 -->

 # End-Block

- Termination of Farming Plan
    - Private Plan
        - distribution stops
        - remove plan states
        - keep stake, reward states for unstakable stakes and claimable rewards each farmers
        - rest of the fund in `farmingPoolAddress` sent to `terminationAddress`, but in Private Plan case, `farmingPoolAddress` == `terminationAddress`, so the fund is not moved
    - Public Plan
        - distribution stops
        - remove plan states
        - keep stake, reward states for unstakable stakes and claimable rewards each farmers
        - rest of the fund in `farmingPoolAddress` sent to `terminationAddress`