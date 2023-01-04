package types

import (
	"errors"
)

// KeyPrefixEpoch defines prefix key for storing epochs.
var KeyPrefixEpoch = []byte{0x01}

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