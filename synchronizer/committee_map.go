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

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (t *CommitteeMapSafe) Range(f func(key common.Address, value etherman.DataCommitteeMember) bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for k, v := range t.members {
		if !f(k, v) {
			break
		}
	}
}

// Length returns the current length of the map.
func (t *CommitteeMapSafe) Length() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.members)
}
