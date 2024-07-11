package synchronizer

import (
	"sync"
	"sync/atomic"

	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/ethereum/go-ethereum/common"
)

// CommitteeMapSafe represents a thread-safe implementation for the data availability committee members map.
type CommitteeMapSafe struct {
	members      sync.Map
	membersCount int32
}

// NewCommitteeMapSafe creates a new CommitteeMapSafe.
func NewCommitteeMapSafe() *CommitteeMapSafe {
	return &CommitteeMapSafe{members: sync.Map{}}
}

// Store sets the value for a key.
func (t *CommitteeMapSafe) Store(member etherman.DataCommitteeMember) {
	_, exists := t.members.LoadOrStore(member.Addr, member)
	if !exists {
		atomic.AddInt32(&t.membersCount, 1)
	} else {
		t.members.Store(member.Addr, member)
	}
}

// StoreBatch sets the range of values and keys.
func (t *CommitteeMapSafe) StoreBatch(members []etherman.DataCommitteeMember) {
	for _, m := range members {
		t.Store(m)
	}
}

// Load returns the value stored in the map for a key, or false if no value is present.
func (t *CommitteeMapSafe) Load(addr common.Address) (etherman.DataCommitteeMember, bool) {
	rawValue, exists := t.members.Load(addr)
	if !exists {
		return etherman.DataCommitteeMember{}, false
	}
	return rawValue.(etherman.DataCommitteeMember), exists //nolint:forcetypeassert
}

// Delete deletes the value for a key.
func (t *CommitteeMapSafe) Delete(key common.Address) {
	_, exists := t.members.LoadAndDelete(key)
	if exists {
		atomic.AddInt32(&t.membersCount, -1)
	}
}

// AsSlice returns a slice of etherman.DataCommitteeMembers.
func (t *CommitteeMapSafe) AsSlice() []etherman.DataCommitteeMember {
	membersSlice := make([]etherman.DataCommitteeMember, 0, atomic.LoadInt32(&t.membersCount))
	t.members.Range(func(_, rawMember any) bool {
		membersSlice = append(membersSlice, rawMember.(etherman.DataCommitteeMember)) //nolint:forcetypeassert

		return true
	})
	return membersSlice
}

// Length returns the current length of the map.
func (t *CommitteeMapSafe) Length() int {
	return int(atomic.LoadInt32(&t.membersCount))
}
