package balancer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/notional-labs/fee-abstraction/tests/interchaintest/osmosistypes/gamm/types"
)

const (
	TypeMsgCreateBalancerPool = "create_balancer_pool"
	TypeMsgMigrateShares      = "migrate_shares"
)

var _ sdk.Msg = &MsgCreateBalancerPool{}

func (msg MsgCreateBalancerPool) Route() string        { return types.RouterKey }
func (msg MsgCreateBalancerPool) Type() string         { return TypeMsgCreateBalancerPool }
func (msg MsgCreateBalancerPool) ValidateBasic() error { return nil }

func (msg MsgCreateBalancerPool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreateBalancerPool) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgMigrateSharesToFullRangeConcentratedPosition{}

func (msg MsgMigrateSharesToFullRangeConcentratedPosition) Route() string        { return types.RouterKey }
func (msg MsgMigrateSharesToFullRangeConcentratedPosition) Type() string         { return TypeMsgMigrateShares }
func (msg MsgMigrateSharesToFullRangeConcentratedPosition) ValidateBasic() error { return nil }

func (msg MsgMigrateSharesToFullRangeConcentratedPosition) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgMigrateSharesToFullRangeConcentratedPosition) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
