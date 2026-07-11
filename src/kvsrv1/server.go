package kvsrv

import (
	"log"
	"sync"

	"6.5840/kvsrv1/rpc"
	"6.5840/labrpc"
	"6.5840/tester1"
)

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type Entry struct {
	Value string
	Version rpc.Tversion
}

type KVServer struct {
	mu sync.Mutex

	// Your definitions here.
	table map[string]Entry
}

func MakeKVServer() *KVServer {
	kv := &KVServer{}
	// Your code here.
	kv.table = make(map[string]Entry)
	return kv
}

// Get returns the value and version for args.Key, if args.Key
// exists. Otherwise, Get returns ErrNoKey.
func (kv *KVServer) Get(args *rpc.GetArgs, reply *rpc.GetReply) {
	// Your code here.
	// always remember to lock before changing server variable
	kv.mu.Lock()
	defer kv.mu.Unlock()

	// directly fetch the key
	entry, ok := kv.table[args.Key]

	// key exist --> different return
	if ok {
		reply.Value = entry.Value
		reply.Version = entry.Version
		reply.Err = rpc.OK
	} else {
		reply.Err = rpc.ErrNoKey
	}
}

// Update the value for a key if args.Version matches the version of
// the key on the server. If versions don't match, return ErrVersion.
// If the key doesn't exist, Put installs the value if the
// args.Version is 0, and returns ErrNoKey otherwise.
func (kv *KVServer) Put(args *rpc.PutArgs, reply *rpc.PutReply) {
	// Your code here.
	// lock
	kv.mu.Lock()
	defer kv.mu.Unlock()

	// update value
	entry, ok := kv.table[args.Key]

	if !ok {
		if args.Version == 0 {
			kv.table[args.Key] = Entry{
				Value: args.Value,
				Version: args.Version + 1,
			}
			reply.Err = rpc.OK
		} else {
			reply.Err = rpc.ErrNoKey
		}

	} else {
		if args.Version == entry.Version {
			kv.table[args.Key] = Entry{
				Value: args.Value,
				Version: entry.Version + 1,
			}
			reply.Err = rpc.OK
		} else {
			reply.Err = rpc.ErrVersion
		}
	}

	return 
}



// You can ignore all arguments; they are for replicated KVservers
func StartKVServer(tc *tester.TesterClnt, ends []*labrpc.ClientEnd, gid tester.Tgid, srv int, persister *tester.Persister) []any {
	kv := MakeKVServer()
	return []any{kv}
}
