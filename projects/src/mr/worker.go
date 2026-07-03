package mr

import "fmt"
import "log"
import "net/rpc"
import "hash/fnv"
import "os"

import (
	"encoding/json"
	"io/ioutil"
	"sort"
	"time"
)

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

var coordSockName string // socket for coordinator


// main/mrworker.go calls this function.
func Worker(sockname string, mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	coordSockName = sockname

	// Your worker implementation here.
	// keep asking --> receiving --> doing --> report done --> asking next one
	for {
		// ask for a task
		reply := requestTask()

		switch reply.TaskType {
		case MapTask:
			doMap(reply, mapf)
			reportDone(MapTask, reply.TaskID)
		case ReduceTask:
			doReduce(reply, reducef)
			reportDone(ReduceTask, reply.TaskID)
		case WaitTask:
			time.Sleep(time.Second)
		case ExitTask:
			return
		}
	}
	// uncomment to send the Example RPC to the coordinator.
	// CallExample()
}

func doMap(reply RequestTaskReply, mapf func(string, string) []KeyValue) {
	// step - 1: read the file
	file, err := os.Open(reply.FileName)
	if err != nil {
		log.Fatalf("cannot open %v", reply.FileName)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", reply.FileName)
	}
	file.Close()

	// step - 2: call mapf: get a list of key-value pairs (value is always "1")
	kva := mapf(reply.FileName, string(content))

	// step - 3: partition into nReduce buckets: an array of array
	// e.g.: [("apple": "1"), ("banana": "1"), ...]
	buckets := make([][]KeyValue, reply.NReduce)
	for _, kv := range kva {
		y := ihash(kv.Key) % reply.NReduce
		buckets[y] = append(buckets[y], kv)
	}

	// step - 4: write each bucket into intermediate file: "mr-X-Y"
	// X: the i-th map task; Y: the i-th reduce
	for y := 0; y < reply.NReduce; y++ {
		filename := fmt.Sprintf("mr-%d-%d", reply.TaskID, y)
		// optimize for atomic writing: write to temp file then rename operation which is atomic
		tempFile, _ := os.CreateTemp("", "mr-tmp-")
		encode := json.NewEncoder(tempFile)
		for _, kv := range buckets[y] {
			encode.Encode(&kv)
		}
		tempFile.Close()
		os.Rename(tempFile.Name(), filename)  // atomic!!! avoid crashing during writing
	}
	return
}

func doReduce(reply RequestTaskReply, reducef func(string, []string) string) {
	// the reduce task ID is the "i-th" bucket
	y := reply.TaskID

	// step - 1: read the intermediate files (>= 1) within the current bucket
	// recall the intermdiate file naming conventions: "mr-X-Y"
	kva := []KeyValue{}
	for x := 0; x < reply.NMap; x++ {
		filename := fmt.Sprintf("mr-%d-%d", x, y)
		file, err := os.Open(filename)
		if err != nil {
			// current bucket does not have file. why? nothing is mapped to this bucket during doMap hash % NReduce
			continue
		}

		decoder := json.NewDecoder(file)
		for {
			var kv KeyValue  // just declare as i will keep using this again and again
			err := decoder.Decode(&kv)
			if err != nil {
				break
			}
			kva = append(kva, kv)
		}
		file.Close() 
	}
	
	// step - 2: sort by key and call reducef
	// the intermediate files from the current reduce bucket may render word-count in messy order
	// we need to put adjacent word next to each other: [("apple": "1"), ("apple": "1"), ("banana": "1")]
	// then we can just consolidate two "apple" easily
	sort.Sort(ByKey(kva))

	// step - 3: group by key and call reducef
	// using temp file again to avoid crash in the middle of writing
	// even crash happens, this is temp file, so won't affect final outcome as coordinator will re-assign the job if it does not receive report done
	ofile, _ := os.CreateTemp("", "mr-out-tmp-")

	// copy from the single process reduce function
	i := 0
	for i < len(kva) {
		j := i + 1
		for j < len(kva) && kva[j].Key == kva[i].Key {
			j++
		}
		values := []string {}
		for k := i; k < j; k++ {
			values = append(values, kva[k].Value)
		}

		output := reducef(kva[i].Key, values)

		fmt.Fprintf(ofile, "%v %v\n", kva[i].Key, output)

		i = j
	}
	ofile.Close()

	// atomic operation! important for distributed system to account for crash in the middle of writing
	os.Rename(ofile.Name(), fmt.Sprintf("mr-out-%d", y))
	return
}

func requestTask() RequestTaskReply {
	args := RequestTaskArgs{}
	reply := RequestTaskReply{}
	call("Coordinator.RequestTask", &args, &reply)
	return reply
}

func reportDone(taskType string, taskID int) {
	args := ReportTaskCompleteArgs{}
	args.TaskType = taskType
	args.TaskID = taskID
	reply := ReportTaskCompleteReply{}
	call("Coordinator.ReportTask", &args, &reply)
	return
} 


// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	c, err := rpc.DialHTTP("unix", coordSockName)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	if err := c.Call(rpcname, args, reply); err == nil {
		return true
	}
	log.Printf("%d: call failed err %v", os.Getpid(), err)
	return false
}
