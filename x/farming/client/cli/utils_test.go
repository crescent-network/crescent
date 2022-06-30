package cli_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"

	"github.com/crescent-network/crescent/v2/app/params"
	"github.com/crescent-network/crescent/v2/x/farming/client/cli"
)

func TestParsePrivateFixedPlan(t *testing.T) {
	okJSON := testutil.WriteToNewTempFile(t, `
{
  "name": "This plan intends to provide incentives for Cosmonauts!",
  "staking_coin_weights": [
    {
      "denom": "PoolCoinDenom",
      "amount": "1.000000000000000000"
    }
  ],
  "start_time": "2021-07-15T08:41:21Z",
  "end_time": "2022-07-16T08:41:21Z",
  "epoch_amount": [
    {
      "denom": "uatom",
      "amount": "1"
    }
  ]
}
`)

	plan, err := cli.ParsePrivateFixedPlan(okJSON.Name())
	require.NoError(t, err)
	require.NotEmpty(t, plan.String())

	require.Equal(t, "This plan intends to provide incentives for Cosmonauts!", plan.Name)
	require.Equal(t, "1.000000000000000000PoolCoinDenom", plan.StakingCoinWeights.String())
	require.Equal(t, "2021-07-15T08:41:21Z", plan.StartTime.Format(time.RFC3339))
	require.Equal(t, "2022-07-16T08:41:21Z", plan.EndTime.Format(time.RFC3339))
	require.Equal(t, "1uatom", plan.EpochAmount.String())
}

func TestParsePrivateRatioPlan(t *testing.T) {
	okJSON := testutil.WriteToNewTempFile(t, `
{
  "name": "This plan intends to provide incentives for Cosmonauts!",
  "staking_coin_weights": [
    {
      "denom": "PoolCoinDenom",
      "amount": "1.000000000000000000"
    }
  ],
  "start_time": "2021-07-15T08:41:21Z",
  "end_time": "2022-07-16T08:41:21Z",
  "epoch_ratio": "1.000000000000000000"
}
`)

	plan, err := cli.ParsePrivateRatioPlan(okJSON.Name())
	require.NoError(t, err)
	require.NotEmpty(t, plan.String())

	require.Equal(t, "This plan intends to provide incentives for Cosmonauts!", plan.Name)
	require.Equal(t, "1.000000000000000000PoolCoinDenom", plan.StakingCoinWeights.String())
	require.Equal(t, "2021-07-15T08:41:21Z", plan.StartTime.Format(time.RFC3339))
	require.Equal(t, "2022-07-16T08:41:21Z", plan.EndTime.Format(time.RFC3339))
	require.Equal(t, "1.000000000000000000", plan.EpochRatio.String())
}

func TestParsePublicPlanProposal(t *testing.T) {
	encodingConfig := params.MakeTestEncodingConfig()

	okJSON := testutil.WriteToNewTempFile(t, `
{
  "title": "Public Farming Plan",
  "description": "Are you ready to farm?",
  "add_plan_requests": [
    {
      "name": "First Public Farming Plan",
      "farming_pool_address": "cosmos1mzgucqnfr2l8cj5apvdpllhzt4zeuh2cshz5xu",
      "termination_address": "cosmos1mzgucqnfr2l8cj5apvdpllhzt4zeuh2cshz5xu",
      "staking_coin_weights": [
        {
          "denom": "PoolCoinDenom",
          "amount": "1.000000000000000000"
        }
      ],
      "start_time": "2021-07-15T08:41:21Z",
      "end_time": "2022-07-16T08:41:21Z",
      "epoch_amount": [
        {
          "denom": "uatom",
          "amount": "1"
        }
      ]
    }
  ]
}
`)

	proposal, err := cli.ParsePublicPlanProposal(encodingConfig.Marshaler, okJSON.Name())
	require.NoError(t, err)

	require.Equal(t, "Public Farming Plan", proposal.Title)
	require.Equal(t, "Are you ready to farm?", proposal.Description)
}
