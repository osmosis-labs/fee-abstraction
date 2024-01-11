package types

import (
	"errors"
	"time"
)

const (
	DefaultSwapPeriod           = time.Minute * 180
	DefaultQueryPeriod          = time.Minute * 60
	DefaultSwapEpochIdentifier  = "swap"
	DefaultQueryEpochIdentifier = "query"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// Validate also validates epoch info.
func (epoch EpochInfo) Validate() error {
	if epoch.Identifier == "" {
		return errors.New("epoch identifier should NOT be empty")
	}
	if epoch.Duration == 0 {
		return errors.New("epoch duration should NOT be 0")
	}
	if epoch.CurrentEpoch < 0 {
		return errors.New("epoch CurrentEpoch must be non-negative")
	}
	if epoch.CurrentEpochStartHeight < 0 {
		return errors.New("epoch CurrentEpoch must be non-negative")
	}
	return nil
}

func NewGenesisEpochInfo(identifier string, duration time.Duration) EpochInfo {
	return EpochInfo{
		Identifier:              identifier,
		StartTime:               time.Time{},
		Duration:                duration,
		CurrentEpoch:            0,
		CurrentEpochStartHeight: 0,
		CurrentEpochStartTime:   time.Time{},
		EpochCountingStarted:    false,
	}
}
