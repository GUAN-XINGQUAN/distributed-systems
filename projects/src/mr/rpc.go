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
	MapTask = "map"
	ReduceTask = "reduce"
)

// helper variable for task status
const (
    WaitTask   = "wait"
    ExitTask   = "exit"
)

// Worker asks the coordinator for job: "give me a task"
type GetTaskArgs struct {
	WorkerID int
}

// Worker gets response from coordinator: "here is your task"
type GetTaskReply struct {
	TaskID 			int	
	TaskType 		string		// which type of task: map vs. reduce
	FileName 		string
	NMap 			int			// total number of map tasks
	NReduce 		int			// total number of reduce tasks
	ReduceTaskLabel int			// need this to indicate "i-th" reduce task
}

// Worker needs to tell coordinator "I am done with my current job"
type ReportTaskCompleteArgs struct {
	WorkerID	int
	TaskType	string			// I am completing "map" or "reduce" task
	TaskID		int				// which task I completed
	IsComplete	bool			// always "true" when calling report
}

// Worker just needs "empty" from coordinator after it reports completes the task
type ReportTaskCompleteReply struct {
	OK bool
}
