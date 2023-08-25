package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	v1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	amino = codec.NewLegacyAmino()

	// AminoCdc references the global x/relationships module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/relationships and
	// defined at the application level.
	AminoCdc = codec.NewAminoCodec(amino)
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSendQueryIbcDenomTWAP{}, "feeabs/SendQueryIbcDenomTWAP", nil)
	cdc.RegisterConcrete(&MsgSwapCrossChain{}, "feeabs/SwapCrossChain", nil)
	cdc.RegisterConcrete(&AddHostZoneProposal{}, "feeabs/AddHostZoneProposal", nil)
	cdc.RegisterConcrete(&DeleteHostZoneProposal{}, "feeabs/DeleteHostZoneProposal", nil)
	cdc.RegisterConcrete(&SetHostZoneProposal{}, "feeabs/SetHostZoneProposal", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSendQueryIbcDenomTWAP{},
		&MsgSwapCrossChain{},
		&MsgFundFeeAbsModuleAccount{},
	)

	registry.RegisterImplementations(
		(*v1beta1types.Content)(nil),
		&AddHostZoneProposal{},
		&DeleteHostZoneProposal{},
		&SetHostZoneProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

func init() {
	RegisterCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)
}
