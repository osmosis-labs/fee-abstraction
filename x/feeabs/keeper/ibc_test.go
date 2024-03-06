package keeper_test

import (
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
)

func (s *KeeperTestSuite) TestGetDecTWAPFromBytes() {
	s.SetupTest()

	data := []byte{10, 19, 50, 49, 52, 50, 56, 53, 55, 49, 52, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48}
	twap, err := s.feeAbsKeeper.GetDecTWAPFromBytes(data)
	require.NoError(s.T(), err)
	require.Equal(s.T(), sdkmath.LegacyMustNewDecFromStr("2.142857140000000000"), twap)
}
