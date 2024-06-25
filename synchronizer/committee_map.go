package synchronizer

import (
	"sync"

	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/ethereum/go-ethereum/common"
)

// CommitteeMapSafe represents a thread-safe implementation for the data availability committee members map
type CommitteeMapSafe struct {
	mu sync.RWMutex

	members map[common.Address]etherman.DataCommitteeMember
}

// NewCommitteeMapSafe creates a new CommitteeMapSafe.
func NewCommitteeMapSafe() *CommitteeMapSafe {
	return &CommitteeMapSafe{
		members: make(map[common.Address]etherman.DataCommitteeMember),
	}
}

// Store sets the value for a key.
func (t *CommitteeMapSafe) Store(key common.Address, value etherman.DataCommitteeMember) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.members[key] = value
}

// StoreBatch sets the range of values and keys.
func (t *CommitteeMapSafe) StoreBatch(members []etherman.DataCommitteeMember) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, m := range members {
		t.members[m.Addr] = m
	}
}

// Load returns the value stored in the map for a key, or false if no value is present.
func (t *CommitteeMapSafe) Load(key common.Address) (etherman.DataCommitteeMember, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	value, ok := t.members[key]
	return value, ok
}

// Delete deletes the value for a key.
func (t *CommitteeMapSafe) Delete(key common.Address) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.members, key)
}

// AsSlice returns a slice of data committee members.
func (t *CommitteeMapSafe) AsSlice() []etherman.DataCommitteeMember {
	t.mu.RLock()
	defer t.mu.RUnlock()

	membersSlice := make([]etherman.DataCommitteeMember, 0, len(t.members))
	for _, m := range t.members {
		membersSlice = append(membersSlice, m)
	}
	return membersSlice
}

// Length returns the current length of the map.
func (t *CommitteeMapSafe) Length() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.members)
}
