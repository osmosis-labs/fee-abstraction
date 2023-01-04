package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidExchangeRate = sdkerrors.Register(ModuleName, 1, "invalid exchange rate")
)
