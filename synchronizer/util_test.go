package synchronizer

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SequenceBatchesValidiumMethodIDs_Equality(t *testing.T) {
	var (
		expectedSequenceBatchesValidiumEtrog      = "2d72c248"
		expectedSequenceBatchesValidiumElderberry = "db5b0ed7"
		expectedSequenceBatchesValidiumBanana     = "165e8a8d"
	)

	require.Equal(t, expectedSequenceBatchesValidiumEtrog, hex.EncodeToString(methodIDSequenceBatchesValidiumEtrog))
	require.Equal(t, expectedSequenceBatchesValidiumElderberry, hex.EncodeToString(methodIDSequenceBatchesValidiumElderberry))
	require.Equal(t, expectedSequenceBatchesValidiumBanana, hex.EncodeToString(methodIDSequenceBatchesValidiumBanana))
}
