package raft

// The file ../raftapi/raftapi.go defines the interface that raft must
// expose to servers (or the tester), but see comments below for each
// of these functions for more details.
//
// In addition,  Make() creates a new raft peer that implements the
// raft interface.

import (
	//	"bytes"
	"math/rand"
	"sync"
	"time"

	//	"6.5840/labgob"
	"6.5840/labrpc"
	"6.5840/raftapi"
	tester "6.5840/tester1"
)

type ServerState string

const (
	Follower  = "follower"
	Candidate = "candidate"
	Leader    = "Leader"
)

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *tester.Persister   // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]

	// Your data here (3A, 3B, 3C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	currentTerm   int         // what round of election I am in
	votedFor      int         // who did I vote for (-1 means no one)
	state         ServerState // follower, candidate, or leader
	lastHeartbeat time.Time   // when did I last hear heartbeat
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (3A).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term = rf.currentTerm
	if rf.state == Leader {
		isleader = true
	} else {
		isleader = false
	}
	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
// before you've implemented snapshots, you should pass nil as the
// second argument to persister.Save().
// after you've implemented snapshots, pass the current snapshot
// (or nil if there's not yet a snapshot).
func (rf *Raft) persist() {
	// Your code here (3C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// raftstate := w.Bytes()
	// rf.persister.Save(raftstate, nil)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (3C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// how many bytes in Raft's persisted log?
func (rf *Raft) PersistBytes() int {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.persister.RaftStateSize()
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (3D).

}

// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (3A, 3B).
	Term        int
	CandidateID int
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (3A).
	Term        int
	VoteGranted bool
}

// RPC argument for AppendEntries
type AppendEntriesArgs struct {
	Term     int
	LeaderId int
}

// RPC reply for AppendEntries
type AppendEntriesReply struct {
	Term    int
	Success bool
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (3A, 3B).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// always the rule: check my term against argument term
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.state = Follower
		rf.votedFor = -1
	}

	// prepare reply variable
	reply.Term = rf.currentTerm
	reply.VoteGranted = false // default reject vote unless we are elgibile for vote (see below)

	// stale vote request: reject directly
	if args.Term < rf.currentTerm {
		return
	}

	// i will vote only if I haven't vote or i have already voted for you
	if rf.votedFor == -1 || rf.votedFor == args.CandidateID {
		reply.VoteGranted = true
		rf.votedFor = args.CandidateID
		rf.lastHeartbeat = time.Now()
	}
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	// when I receive a heartbeat request
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// always check: if I receive heartbeat, is my term older --> become follower
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.state = Follower
		rf.votedFor = -1
	}

	// prepare RPC response
	reply.Term = rf.currentTerm
	reply.Success = false // default: I will not admit your append entry

	// reject stale request
	if args.Term < rf.currentTerm {
		return
	}

	// accept the correct append entry request
	rf.state = Follower
	rf.lastHeartbeat = time.Now()
	reply.Success = true
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (3B).

	return index, term, isLeader
}

func (rf *Raft) ticker() {
	for true {

		// Your code here (3A)
		// Check if a leader election should be started.
		rf.mu.Lock()
		timeout := time.Duration(300+rand.Intn(150)) * time.Millisecond
		// Case 1: I am follower and I haven't heard heartbeat --> mark me as candidate and start election
		// Case 2: I am candidate and election timeout elapses --> start a new election
		if rf.state != Leader && time.Since(rf.lastHeartbeat) > timeout {
			rf.startElection()
		}
		rf.mu.Unlock()

		// pause for a random amount of time between 50 and 350
		// milliseconds.
		ms := 50 + (rand.Int63() % 300)
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
}

func (rf *Raft) startElection() {
	// This function kicks of the election process and ask for votes from all other servers
	// Phase 1: mark myself as candidate
	rf.state = Candidate
	rf.votedFor = rf.me           // I vote for myself
	rf.currentTerm++              // new round of election so my term goes by 1
	rf.lastHeartbeat = time.Now() // update the heartbeat time

	// prepare for vote collection
	voteCount := 1         // I already got one vote from myself
	term := rf.currentTerm // which term am I kicking off election (important!!!)

	// Phase 2: ask all others to vote for me
	for serverIdx := range rf.peers {
		// no need to ask for myself because I voted for myself already in Phase 1
		if serverIdx == rf.me {
			continue
		}

		// Use go routine to ask vote concurrently
		go func(peerId int) {
			// construct RPC call argument and reply
			args := RequestVoteArgs{
				Term:        term,
				CandidateID: rf.me,
			}
			reply := RequestVoteReply{}
			ok := rf.sendRequestVote(peerId, &args, &reply)
			if ok {
				rf.mu.Lock()
				defer rf.mu.Unlock()

				// check stale reply: if reply is not relevant --> no need to continue ask for vote
				if rf.currentTerm != term || rf.state != Candidate {
					return
				}

				// am I outdated?
				if reply.Term > rf.currentTerm {
					rf.state = Follower
					rf.currentTerm = reply.Term
					rf.votedFor = -1
					return
				}

				// otherwise, right timing to collect vote and count
				if reply.VoteGranted {
					voteCount++
					if voteCount > len(rf.peers)/2 {
						rf.becomeLeader()
					}
				}
			}
		}(serverIdx)
	}
}

func (rf *Raft) becomeLeader() {
	rf.state = Leader
	go rf.startHeatbeat() // leader needs to do heartbeat all the time
}

func (rf *Raft) startHeatbeat() {
	// send the heartbeat all the time until I am not the leader anymore
	for true {
		rf.mu.Lock()

		// I am no longer leader: stop immediately
		if rf.state != Leader {
			rf.mu.Unlock()
			return
		}

		// read the term
		term := rf.currentTerm
		rf.mu.Unlock()

		// broadcast my term to all other servers
		for serverIdx := range rf.peers {
			// avoid signaling myself
			if serverIdx == rf.me {
				continue
			}
			// talk with others using go routine in parallel
			go func(peerId int) {
				// construct RPC call request and reply
				args := AppendEntriesArgs{
					Term:     term,
					LeaderId: rf.me,
				}
				reply := AppendEntriesReply{}
				ok := rf.sendAppendEntries(peerId, &args, &reply)
				if ok {
					rf.mu.Lock()
					defer rf.mu.Unlock()
					// If RPC responses show my current term is lower: step down to become follower
					if reply.Term > rf.currentTerm {
						rf.currentTerm = reply.Term
						rf.state = Follower
						rf.votedFor = -1
					}
				}
			}(serverIdx)
		}
		// lab requires us to send heartbeat 10 times per second: sleep 10 ms
		time.Sleep(100 * time.Millisecond)
	}
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *tester.Persister, applyCh chan raftapi.ApplyMsg) raftapi.Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (3A, 3B, 3C).
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.state = Follower
	rf.lastHeartbeat = time.Now()

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()

	return rf
}
