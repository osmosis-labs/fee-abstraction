package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func (s *KeeperTestSuite) TestGetDecTWAPFromBytes() {
	s.SetupTest()
	// represent the payload 0 0 [10 19 50 49 52 50 56 53 55 49 52 48 48 48 48 48 48 48 48 48 48] [] <nil> 0
	data := []byte{10, 19, 50, 49, 52, 50, 56, 53, 55, 49, 52, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48}

	twap, err := s.feeAbsKeeper.GetDecTWAPFromBytes(data)
	require.NoError(s.T(), err)
	require.Equal(s.T(), sdkmath.LegacyMustNewDecFromStr("2.142857140000000000"), twap)
}
