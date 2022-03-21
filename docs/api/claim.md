---
Title: Claim
Description: A high-level overview of what gRPC-gateway REST routes are supported in the claim module.
---

# Claim Module

## Synopsis

This document provides a high-level overview of what gRPC-gateway REST routes are supported in the `claim` module.

## gRPC-gateway REST Routes

<!-- markdown-link-check-disable -->
++https://github.com/crescent-network/crescent/blob/main/proto/crescent/claim/v1beta1/query.proto 

- [Airdrops](#Airdrops)
- [Airdrop](#Airdrop)
- [ClaimRecord](#ClaimRecord)


## Airdrops

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/claim/v1beta1/airdrops
```

Example Response

```json
{
  "airdrops": [
    {
      "id": "1",
      "source_address": "cre15rz2rwnlgr7nf6eauz52usezffwrxc0mxajpmw",
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
http://localhost:1317/crescent/claim/v1beta1/airdrops/1
```

Example Response

```json
{
  "airdrop": {
    "id": "1",
    "source_address": "cre15rz2rwnlgr7nf6eauz52usezffwrxc0mxajpmw",
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
http://localhost:1317/crescent/claim/v1beta1/airdrops/1/claim_records/cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p
```

Example Response

```json
{
  "claim_record": {
    "airdrop_id": "1",
    "recipient": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
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