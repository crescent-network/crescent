---
Title: Claim
Description: A high-level overview of what gRPC-gateway REST routes are supported in the claim module.
---

# Claim Module

## Synopsis

This document provides a high-level overview of what gRPC-gateway REST routes are supported in the claim module.

## gRPC-gateway REST Routes

<!-- markdown-link-check-disable -->
++https://github.com/cosmosquad-labs/squad/blob/main/proto/squad/claim/v1beta1/query.proto 

- [Airdrops](#Airdrops)
- [Airdrop](#Airdrop)
- [ClaimRecord](#ClaimRecord)


## Airdrops

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/claim/v1beta1/airdrops
```

Example Response

```json
{
  "airdrops": [
    {
      "id": "1",
      "source_address": "cosmos15rz2rwnlgr7nf6eauz52usezffwrxc0mz4pywr",
      "conditions": [
        "CONDITION_TYPE_DEPOSIT",
        "CONDITION_TYPE_SWAP",
        "CONDITION_TYPE_LIQUIDSTAKE",
        "CONDITION_TYPE_VOTE"
      ],
      "start_time": "2022-02-01T00:00:00Z",
      "end_time": "2022-03-21T00:00:00Z"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## Airdrop

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/claim/v1beta1/airdrops/1
```

Example Response

```json
{
  "airdrop": {
    "id": "1",
    "source_address": "cosmos15rz2rwnlgr7nf6eauz52usezffwrxc0mz4pywr",
    "conditions": [
      "CONDITION_TYPE_DEPOSIT",
      "CONDITION_TYPE_SWAP",
      "CONDITION_TYPE_LIQUIDSTAKE",
      "CONDITION_TYPE_VOTE"
    ],
    "start_time": "2022-02-01T00:00:00Z",
    "end_time": "2022-03-21T00:00:00Z"
  }
}
```


## ClaimRecord

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/claim/v1beta1/airdrops/1/claim_records/cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v
```

Example Response

```json
{
  "claim_record": {
    "airdrop_id": "1",
    "recipient": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
    "initial_claimable_coins": [
      {
        "denom": "airdrop",
        "amount": "3000000000000"
      }
    ],
    "claimable_coins": [
      {
        "denom": "airdrop",
        "amount": "3000000000000"
      }
    ],
    "claimed_conditions": [
    ]
  }
}
```