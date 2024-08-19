package types

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
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

func TestRemoveDuplicateOffChainData(t *testing.T) {
	type args struct {
		ods []OffChainData
	}
	tests := []struct {
		name string
		args args
		want []OffChainData
	}{
		{
			name: "no duplicates",
			args: args{
				ods: []OffChainData{
					{
						Key: common.BytesToHash([]byte("key1")),
					},
					{
						Key: common.BytesToHash([]byte("key2")),
					},
				},
			},
			want: []OffChainData{
				{
					Key: common.BytesToHash([]byte("key1")),
				},
				{
					Key: common.BytesToHash([]byte("key2")),
				},
			},
		},
		{
			name: "with duplicates",
			args: args{
				ods: []OffChainData{
					{
						Key: common.BytesToHash([]byte("key1")),
					},
					{
						Key: common.BytesToHash([]byte("key1")),
					},
				},
			},
			want: []OffChainData{
				{
					Key: common.BytesToHash([]byte("key1")),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, RemoveDuplicateOffChainData(tt.args.ods), "RemoveDuplicateOffChainData(%v)", tt.args.ods)
		})
	}
}
