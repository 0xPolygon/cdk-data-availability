package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsHexValid(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "valid hex without 0x prefix",
			s:    "76616c69642068657820776974686f757420307820707265666978",
			want: true,
		},
		{
			name: "valid hex with 0x prefix",
			s:    "0x76616c696420686578207769746820307820707265666978",
			want: true,
		},
		{
			name: "invalid hex without 0x prefix",
			s:    "76616c696invalid07769746820307820707265666978",
			want: false,
		},
		{
			name: "invalid hex with 0x prefix",
			s:    "0x76616c696invalid07769746820307820707265666978",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, IsHexValid(tt.s))
		})
	}
}
