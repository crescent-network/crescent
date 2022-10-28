package cli_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v3/app/params"
	"github.com/crescent-network/crescent/v3/x/lpfarm/client/cli"
)

func TestParseFarmingPlanProposal(t *testing.T) {
	okJSON := testutil.WriteToNewTempFile(t, `
{
  "title": "Farming Plan Proposal",
  "description": "Let's start farming",
  "create_plan_requests": [
    {
      "description": "New Farming Plan",
      "farming_pool_address": "cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3",
      "reward_allocations": [
        {
          "pair_id": "1",
          "rewards_per_day": [
            {
              "denom": "stake",
              "amount": "100000000"
            }
          ]
        },
        {
          "denom": "pool2",
          "rewards_per_day": [
            {
              "denom": "stake",
              "amount": "200000000"
            }
          ]
        }
      ],
      "start_time": "2022-01-01T00:00:00Z",
      "end_time": "2023-01-01T00:00:00Z"
    }
  ],
  "terminate_plan_requests": [
    {
      "plan_id": "1"
    },
    {
      "plan_id": "2"
    }
  ]
}
`)

	encodingConfig := params.MakeTestEncodingConfig()

	plan, err := cli.ParseFarmingPlanProposal(encodingConfig.Marshaler, okJSON.Name())
	require.NoError(t, err)
	require.NotEmpty(t, plan.String())
}
