package lock

import (
	"6.5840/kvtest1"
	"6.5840/kvsrv1/rpc"
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
	lk.lockName = lockname
	return lk
}

func (lk *Lock) Acquire() {
	// Your code here
	for {
		// get current state
		value, version, _ := lk.ck.Get(lk.lockName)

		// if locked, do nothing but to wait
		if value == LOCK {
			continue
		}
		
		// all other cases: attempt to lock
		putErr := lk.ck.Put(lk.lockName, LOCK, version)

		// if success return
		if putErr == rpc.OK {
			return
		}
	}
}

func (lk *Lock) Release() {
	// Your code here
	_, version, _ := lk.ck.Get(lk.lockName)
	lk.ck.Put(lk.lockName, UNBLOCK, version)
}
