package balancer

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type balancerPoolPretty struct {
	Address            sdk.AccAddress `json:"address" yaml:"address"`
	Id                 uint64         `json:"id" yaml:"id"`
	PoolParams         PoolParams     `json:"pool_params" yaml:"pool_params"`
	FuturePoolGovernor string         `json:"future_pool_governor" yaml:"future_pool_governor"`
	TotalWeight        sdk.Dec        `json:"total_weight" yaml:"total_weight"`
	TotalShares        sdk.Coin       `json:"total_shares" yaml:"total_shares"`
	PoolAssets         []PoolAsset    `json:"pool_assets" yaml:"pool_assets"`
}

func (p Pool) String() string {
	out, err := p.MarshalJSON()
	if err != nil {
		panic(err)
	}
	return string(out)
}

// MarshalJSON returns the JSON representation of a Pool.
func (p Pool) MarshalJSON() ([]byte, error) {
	accAddr, err := sdk.AccAddressFromBech32(p.Address)
	if err != nil {
		return nil, err
	}

	decTotalWeight := sdk.NewDecFromInt(p.TotalWeight)

	return json.Marshal(balancerPoolPretty{
		Address:            accAddr,
		Id:                 p.Id,
		PoolParams:         p.PoolParams,
		FuturePoolGovernor: p.FuturePoolGovernor,
		TotalWeight:        decTotalWeight,
		TotalShares:        p.TotalShares,
		PoolAssets:         p.PoolAssets,
	})
}