package types

import (
	sdkerrors "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgFundFeeAbsModuleAccount{}

// Route Implements Msg.
func (m MsgFundFeeAbsModuleAccount) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgFundFeeAbsModuleAccount) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners returns the expected signers for a MsgMintAndAllocateExp .
func (m MsgFundFeeAbsModuleAccount) GetSigners() []sdk.AccAddress {
	daoAccount, err := sdk.AccAddressFromBech32(m.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{daoAccount}
}

// GetSignBytes Implements Msg.
func (m MsgFundFeeAbsModuleAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// ValidateBasic does a sanity check on the provided data.
func (m MsgFundFeeAbsModuleAccount) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.FromAddress)
	if err != nil {
		return sdkerrors.Wrap(err, "from address must be valid address")
	}
	return nil
}

func NewMsgFundFeeAbsModuleAccount(fromAddr sdk.AccAddress, amount sdk.Coins) *MsgFundFeeAbsModuleAccount {
	return &MsgFundFeeAbsModuleAccount{
		FromAddress: fromAddr.String(),
		Amount:      amount,
	}
}
