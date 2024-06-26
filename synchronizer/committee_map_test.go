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

	members := []etherman.DataCommitteeMember{
		{Addr: common.HexToAddress("0x1"), URL: "Member 1"},
		{Addr: common.HexToAddress("0x2"), URL: "Member 2"},
		{Addr: common.HexToAddress("0x3"), URL: "Member 3"},
		{Addr: common.HexToAddress("0x4"), URL: "Member 4"},
		{Addr: common.HexToAddress("0x5"), URL: "Member 5"},
		{Addr: common.HexToAddress("0x6"), URL: "Member 6"},
	}

	var wg sync.WaitGroup
	wg.Add(2)

	ch := make(chan etherman.DataCommitteeMember)

	go func() {
		defer wg.Done()

		for _, m := range members {
			committee.Store(m)
			ch <- m
		}
		close(ch)
	}()

	var actualMembers []etherman.DataCommitteeMember
	go func() {
		defer wg.Done()

		for m := range ch {
			member, ok := committee.Load(m.Addr)
			require.True(t, ok)
			actualMembers = append(actualMembers, member)
		}
	}()

	wg.Wait()

	require.Len(t, actualMembers, len(members))
	for i, m := range members {
		require.Equal(t, m, actualMembers[i])
	}

	// replace the single committee member
	replacedMember := etherman.DataCommitteeMember{Addr: members[0].Addr, URL: "New Member 1"}
	committee.Store(replacedMember)
	require.Equal(t, len(members), committee.Length())
	actualReplacedMember, exists := committee.Load(replacedMember.Addr)
	require.True(t, exists)
	require.Equal(t, replacedMember, actualReplacedMember)
	// skip the first member, because it is replaced and already asserted
	for i, m := range members[1:] {
		require.Equal(t, m, actualMembers[i+1])
	}
}

func TestDelete(t *testing.T) {
	committee := NewCommitteeMapSafe()

	member := etherman.DataCommitteeMember{Addr: common.HexToAddress("0x1"), URL: "Member 1"}

	committee.Store(member)
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

func TestAsSlice(t *testing.T) {
	committee := NewCommitteeMapSafe()
	committee.StoreBatch(
		[]etherman.DataCommitteeMember{
			{Addr: common.HexToAddress("0x1"), URL: "Member 1"},
			{Addr: common.HexToAddress("0x2"), URL: "Member 2"},
			{Addr: common.HexToAddress("0x3"), URL: "Member 3"},
			{Addr: common.HexToAddress("0x4"), URL: "Member 4"},
		})

	membersSlice := committee.AsSlice()

	require.Equal(t, committee.Length(), len(membersSlice))
	for _, member := range membersSlice {
		foundMember, ok := committee.Load(member.Addr)
		require.True(t, ok)
		require.Equal(t, foundMember, member)
	}
}

func TestLength(t *testing.T) {
	committee := NewCommitteeMapSafe()

	members := []etherman.DataCommitteeMember{
		{Addr: common.HexToAddress("0x1"), URL: "http://localhost:1001"},
		{Addr: common.HexToAddress("0x2"), URL: "http://localhost:1002"},
	}

	committee.StoreBatch(members)

	require.Equal(t, len(members), committee.Length())
}
