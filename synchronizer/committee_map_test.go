package synchronizer

import (
	"sync"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestStoreAndLoad(t *testing.T) {
	committee := NewCommitteeMapSafe()

	member := etherman.DataCommitteeMember{Addr: common.HexToAddress("0x1"), URL: "Member 1"}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		committee.Store(member.Addr, member)
	}()

	wg.Wait()

	loadedMember, ok := committee.Load(member.Addr)
	require.True(t, ok)
	require.Equal(t, member, loadedMember)
}

func TestDelete(t *testing.T) {
	committee := NewCommitteeMapSafe()

	member := etherman.DataCommitteeMember{Addr: common.HexToAddress("0x1"), URL: "Member 1"}

	committee.Store(member.Addr, member)
	committee.Delete(member.Addr)

	_, ok := committee.Load(member.Addr)
	require.False(t, ok)
}

func TestStoreBatch(t *testing.T) {
	committee := NewCommitteeMapSafe()

	members := []etherman.DataCommitteeMember{
		{Addr: common.HexToAddress("0x1"), URL: "http://localhost:1001"},
		{Addr: common.HexToAddress("0x2"), URL: "http://localhost:1002"},
		{Addr: common.HexToAddress("0x3"), URL: "http://localhost:1003"},
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		committee.StoreBatch(members)
	}()

	wg.Wait()

	for _, member := range members {
		loadedMember, ok := committee.Load(member.Addr)
		require.True(t, ok)
		require.Equal(t, member, loadedMember)
	}
}

func TestRange(t *testing.T) {
	committee := NewCommitteeMapSafe()

	members := []etherman.DataCommitteeMember{
		{Addr: common.HexToAddress("0x1"), URL: "Member 1"},
		{Addr: common.HexToAddress("0x2"), URL: "Member 2"},
		{Addr: common.HexToAddress("0x3"), URL: "Member 3"},
	}

	committee.StoreBatch(members)

	foundMembers := make(map[common.Address]etherman.DataCommitteeMember)
	committee.Range(func(key common.Address, value etherman.DataCommitteeMember) bool {
		foundMembers[key] = value
		return true
	})

	require.Equal(t, len(members), len(foundMembers))
	for _, member := range members {
		require.Equal(t, member, foundMembers[member.Addr])
	}
}

// Test for Length method
func TestLength(t *testing.T) {
	committee := NewCommitteeMapSafe()

	members := []etherman.DataCommitteeMember{
		{Addr: common.HexToAddress("0x1"), URL: "http://localhost:1001"},
		{Addr: common.HexToAddress("0x2"), URL: "http://localhost:1002"},
		{Addr: common.HexToAddress("0x3"), URL: "http://localhost:1003"},
	}

	committee.StoreBatch(members)

	length := committee.Length()
	require.Equal(t, len(members), length)
}
