package types

import (
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgSendQueryIbcDenomTWAP{}

// Route Implements Msg.
func (m MsgSendQueryIbcDenomTWAP) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgSendQueryIbcDenomTWAP) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners returns the expected signers for a MsgMintAndAllocateExp .
func (m MsgSendQueryIbcDenomTWAP) GetSigners() []sdk.AccAddress {
	daoAccount, err := sdk.AccAddressFromBech32(m.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{daoAccount}
}

// GetSignBytes Implements Msg.
func (m MsgSendQueryIbcDenomTWAP) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// ValidateBasic does a sanity check on the provided data.
func (m MsgSendQueryIbcDenomTWAP) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.FromAddress)
	if err != nil {
		return sdkerrors.Wrap(err, "from address must be valid address")
	}
	return nil
}

func NewMsgSendQueryIbcDenomTWAP(fromAddr sdk.AccAddress) *MsgSendQueryIbcDenomTWAP {
	return &MsgSendQueryIbcDenomTWAP{
		FromAddress: fromAddr.String(),
	}
}

var _ sdk.Msg = &MsgSwapCrossChain{}

// Route Implements Msg.
func (m MsgSwapCrossChain) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgSwapCrossChain) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners returns the expected signers for a MsgMintAndAllocateExp .
func (m MsgSwapCrossChain) GetSigners() []sdk.AccAddress {
	daoAccount, err := sdk.AccAddressFromBech32(m.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{daoAccount}
}

// GetSignBytes Implements Msg.
func (m MsgSwapCrossChain) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// ValidateBasic does a sanity check on the provided data.
func (m MsgSwapCrossChain) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.FromAddress)
	if err != nil {
		return sdkerrors.Wrap(err, "from address must be valid address")
	}
	return nil
}

func NewMsgSwapCrossChain(fromAddr sdk.AccAddress, ibcDenom string) *MsgSwapCrossChain {
	return &MsgSwapCrossChain{
		FromAddress: fromAddr.String(),
		IbcDenom:    ibcDenom,
	}
}
