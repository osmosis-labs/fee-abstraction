package types

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	ErrInvalidExchangeRate          = sdkerrors.Register(ModuleName, 1, "invalid exchange rate")
	ErrInvalidIBCFees               = sdkerrors.Register(ModuleName, 2, "invalid ibc fees")
	ErrHostZoneConfigNotFound       = sdkerrors.Register(ModuleName, 3, "host chain config not found")
	ErrDuplicateHostZoneConfig      = sdkerrors.Register(ModuleName, 4, "duplicate config")
	ErrHostZoneFrozen               = sdkerrors.Register(ModuleName, 5, "hostzone frozen")
	ErrNotEnoughFundInModuleAddress = sdkerrors.Register(ModuleName, 6, "not have funding yet")
)
