package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_DefaultsConstructor(t *testing.T) {
	dflt, err := Default()
	require.NoError(t, err)
	require.Equal(t, "committee_user", dflt.DB.User)
}
