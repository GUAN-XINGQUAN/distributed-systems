package lock

import (
	"6.5840/kvtest1"
)

const (
	// lock state
	LOCK = "LOCK"
	UNBLOCK = "UNBLOCK"
)

type Lock struct {
	// IKVClerk is a go interface for k/v clerks: the interface hides
	// the specific Clerk type of ck but promises that ck supports
	// Put and Get.  The tester passes the clerk in when calling
	// MakeLock().
	ck kvtest.IKVClerk
	// You may add code here
	lockName string
	lockState string
}

// The tester calls MakeLock() and passes in a k/v clerk; your code can
// perform a Put or Get by calling lk.ck.Put() or lk.ck.Get().
//
// This interface supports multiple locks by means of the
// lockname argument; locks with different names should be
// independent.
func MakeLock(ck kvtest.IKVClerk, lockname string) *Lock {
	lk := &Lock{ck: ck}
	// You may add code here
	lk.lockName = lockName
	return lk
}

func (lk *Lock) Acquire() {
	// Your code here
	for {
		// get current state
		value, version, err := ck.Get(lk.lockName)

		// if locked, do nothing but to wait
		if value == LOCK {
			continue
		} 
		
		// attempt conditional put

		// if success return

		// other retry

	}
}

func (lk *Lock) Release() {
	// Your code here
}
