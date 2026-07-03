package mr

import "log"
import "net"
import "os"
import "net/rpc"
import "net/http"

import (
	"sync"
	"time"
)

type TaskState string

const (
	Idle TaskState = "idle"
	InProgress TaskState = "in_progress"
	Completed TaskState = "completed"
)

type Task struct {
	State TaskState
	StartTime time.Time
	FileName string
}

type Coordinator struct {
	// Your definitions here.
	mu sync.Mutex
	mapTasks []Task
	reduceTasks []Task

	nMap int
	nReduce int

	allMapTaskComplete bool
	allReduceTaskComplete bool
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) RequestTask (args *RequestTaskArgs, reply *RequestTaskReply) error {
	// concurrency lock
	c.mu.Lock()
	defer c.mu.Unlock()  // this equals put "c.mu.Unlock" right before each return

	// phase - 1: attempt to find a map task and assign to worker
	if c.allMapTaskComplete == false {
		for i, task := range c.mapTasks {
			if task.State == Idle {
				// idle map task found --> assign to worker
				reply.TaskType = MapTask
				reply.TaskID = i
				reply.FileName = task.FileName
				reply.NReduce = c.nReduce
				reply.NMap = c.nMap

				// update the task status at coordinator memory
				c.mapTasks[i].State = InProgress
				c.mapTasks[i].StartTime = time.Now()
				return nil  // mutex is release because we use "defer" syntax
			}
		}
		// no idle map task but we have NOT complete all map task, so we cannot let this worker get any job (no reduce job before all map job done)
		reply.TaskType = WaitTask
		return nil
	}

	// phase - 2: now all map task complets. why? if not, we have TWO returns above and the function won't come here
	for i, task := range c.reduceTasks {
		if task.State == Idle{
			// reduce task found --> assign to the current worker
			reply.TaskType = ReduceTask
			reply.TaskID = i
			// reply.FileName = task.FileName  // reduce does not need explicit file name
			reply.NReduce = c.nReduce
			reply.NMap = c.nMap
			
			// update the task tracking at coordinator side
			c.reduceTasks[i].State = InProgress
			c.reduceTasks[i].StartTime = time.Now()
			return nil 
		}
	}

	// phase - 3: no available map task; no available reduce task
	if c.allReduceTaskComplete == true {
		reply.TaskType = ExitTask
	} else {
		reply.TaskType = WaitTask
	}
	return nil
}

func (c *Coordinator) ReportTask (args *ReportTaskCompleteArgs, reply *ReportTaskCompleteReply) error {
	// concurrency lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// when this function is called, task type can only be map or reduce
	if args.TaskType == MapTask {
		c.mapTasks[args.TaskID].State = Completed
		c.allMapTaskComplete = c.CheckAllMapTaskComplete()
	} else {
		c.reduceTasks[args.TaskID].State = Completed
		c.allReduceTaskComplete = c.CheckAllReduceTaskComplete()
	}
	return nil
}

func (c *Coordinator) CheckAllMapTaskComplete() bool {
	for _, task := range c.mapTasks {
		if task.State != Completed {
			return false
		}
	}
	return true
}

func (c *Coordinator) CheckAllReduceTaskComplete() bool {
	for _, task := range c.reduceTasks {
		if task.State != Completed {
			return false
		}
	}
	return true
}

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}


// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server(sockname string) {
	rpc.Register(c)
	rpc.HandleHTTP()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatalf("listen error %s: %v", sockname, e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	// ret := false

	// Your code here.
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.allMapTaskComplete && c.allReduceTaskComplete
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(sockname string, files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	// initialize map tasks
	for _, f := range(files) {
		task := Task{
			State: Idle,
			FileName: f,
		}
		c.mapTasks = append(c.mapTasks, task)
	}

	// initialize reduce tasks
	for i := 0; i < nReduce; i++ {
		task := Task{
			State: Idle,
		}
		c.reduceTasks = append(c.reduceTasks, task)
	}

	// other server state variables
	c.nMap = len(files)
	c.nReduce = nReduce
	c.allMapTaskComplete = false
	c.allReduceTaskComplete = false

	c.server(sockname)
	return &c
}
