package app

import (
	ibcante "github.com/cosmos/ibc-go/v7/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errorstypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	feeabsante "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/ante"
	feeabskeeper "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/keeper"
)

// HandlerOptions extends the SDK's AnteHandler options by requiring the IBC
type HandlerOptions struct {
	ante.HandlerOptions

	IBCKeeper    *ibckeeper.Keeper
	FeeAbskeeper feeabskeeper.Keeper
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(errorstypes.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(errorstypes.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(errorstypes.ErrLogic, "sign mode handler is required for ante builder")
	}

	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.RejectExtensionOptionsDecorator{},
		feeabsante.NewFeeAbstrationMempoolFeeDecorator(options.FeeAbskeeper),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		feeabsante.NewFeeAbstractionDeductFeeDecorate(options.AccountKeeper, options.BankKeeper, options.FeeAbskeeper, options.FeegrantKeeper),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
