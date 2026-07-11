package lock

import (
	"6.5840/kvtest1"
	"6.5840/kvsrv1/rpc"
)

const (
	// lock state
	LOCK = "LOCK"
	UNLOCK = "UNLOCK"
)

type Lock struct {
	// IKVClerk is a go interface for k/v clerks: the interface hides
	// the specific Clerk type of ck but promises that ck supports
	// Put and Get.  The tester passes the clerk in when calling
	// MakeLock().
	ck kvtest.IKVClerk
	// You may add code here
	lockName string
	clientID string  // unique identity for each client
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
	lk.clientID = kvtest.RandValue(8)
	return lk
}

func (lk *Lock) Acquire() {
	// Your code here
	for {
		// get current state
		value, version, _ := lk.ck.Get(lk.lockName)

		if value == lk.clientID {  // if hold the lock and owner is me
			return
		}

		// if locked but owner is not me, do nothing but to wait
		if value != "" && value != UNLOCK && value != lk.clientID {
			continue
		}
		
		// all other cases: attempt to lock --> put my ID there as a lock state so that others know it's me
		putErr := lk.ck.Put(lk.lockName, lk.clientID, version)

		// if success return
		if putErr == rpc.OK {
			return
		}
		if putErr == rpc.ErrMaybe {  // if error case: confirm whether I have already gained the lock
			value, _, _ := lk.ck.Get(lk.lockName)
			if value == lk.clientID {
				return
			}
		}
	}
}

func (lk *Lock) Release() {
	// Your code here
	for {
		value, version, _ := lk.ck.Get(lk.lockName)
		if value == UNLOCK {  // already released: nothing to do
			return
		}
		putErr := lk.ck.Put(lk.lockName, UNLOCK, version)
		if putErr == rpc.OK {
			return
		}
		if putErr == rpc.ErrMaybe {
			value, _, _ = lk.ck.Get(lk.lockName)
			if value == UNLOCK {
				return
			}
		}
	}
}
