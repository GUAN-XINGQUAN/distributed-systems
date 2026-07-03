package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

// helper variable for ENUM task type
const (
	MapTask = "Map"
	ReduceTask = "Reduce"
    WaitTask = "Wait"
    ExitTask = "Exit"
)

// Worker asks the coordinator for job: "give me a task"
type RequestTaskArgs struct {
	// empty; this RPC sends nothing but to let coordinator know
}

// Worker gets response from coordinator: "here is your task"
type RequestTaskReply struct {
	TaskType string
	TaskID int
	FileName string
	NReduce int
	NMap int
}

// Worker needs to tell coordinator "I am done with my current job"
type ReportTaskCompleteArgs struct {
	TaskType string
	TaskID int
}

// Worker just needs "empty" from coordinator after it reports completes the task
type ReportTaskCompleteReply struct {
	// nothing; the worker does not need anything from the coordinator but to acknowledge
}
