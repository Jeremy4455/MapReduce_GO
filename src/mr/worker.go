package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// ByKey 实际上就是一个 []KeyValue
type ByKey []KeyValue

// 实现 sort.Interface 接口

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(masterAddr string, mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	for {
		reply := CallForATask(masterAddr)
		if reply != nil {
			switch reply.MsgType {
			case MapTaskAlloc:
				{
					err := HandleMapTask(reply, mapf)
					if err != nil {
						CallForReportStatus(masterAddr, MapFailed, reply.TaskID)
					} else {
						CallForReportStatus(masterAddr, MapSuccess, reply.TaskID)
					}
				}
			case ReduceTaskAlloc:
				{
					err := HandleReduceTask(reply, reducef)
					if err != nil {
						CallForReportStatus(masterAddr, ReduceFailed, reply.TaskID)
					} else {
						CallForReportStatus(masterAddr, ReduceSuccess, reply.TaskID)
					}
				}
			case Wait:
				time.Sleep(time.Second * 10)
			case Shutdown:
				os.Exit(0)
			}
		}
		time.Sleep(time.Second)
	}
	// uncomment to send the Example RPC to the master.
	// CallExample()

}

func HandleReduceTask(reply *MyReply, reducef func(string, []string) string) error {
	key_id := reply.TaskID

	kvs := map[string][]string{}

	fileList, err := ReadSpecificFile(key_id, "./")
	if err != nil {
		return err
	}

	for _, file := range fileList {
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			err := dec.Decode(&kv)
			if err != nil {
				break
			}
			kvs[kv.Key] = append(kvs[kv.Key], kv.Value)
		}
		file.Close()
	}

	var keys []string
	for key := range kvs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	outputDir := "output"
	oname := filepath.Join(outputDir, "mr-out-"+strconv.Itoa(reply.TaskID))

	ofile, err := os.Create(oname)
	if err != nil {
		return err
	}
	defer ofile.Close()

	for _, key := range keys {
		output := reducef(key, kvs[key])
		_, err := fmt.Fprintf(ofile, "%v %v\n", key, output)
		if err != nil {
			return err
		}
	}

	DelFileByReduceId(reply.TaskID, "./")

	return nil
}

func HandleMapTask(reply *MyReply, mapf func(string, string) []KeyValue) error {
	// 以文件名作为任务名，根据键排序处理结果，并合并相同Key的键值对
	file, err := os.Open(reply.TaskName)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	kvs := mapf(reply.TaskName, string(content))
	sort.Sort(ByKey(kvs))

	oname_prefix := "mr-out-" + strconv.Itoa(reply.TaskID) + "-"

	key_group := map[string][]string{}
	for _, kv := range kvs {
		key_group[kv.Key] = append(key_group[kv.Key], kv.Value)
	}

	_ = DelFileByMapID(reply.TaskID, "./")

	for key, values := range key_group {
		redId := ihash(key)
		oname := oname_prefix + strconv.Itoa(redId%reply.NReduce)
		var ofile *os.File
		if _, err := os.Stat(oname); os.IsNotExist(err) {
			ofile, _ = os.Create(oname)
		} else {
			ofile, _ = os.OpenFile(oname, os.O_APPEND|os.O_WRONLY, 0644)
		}

		enc := json.NewEncoder(ofile)
		for _, value := range values {
			err := enc.Encode(&KeyValue{Key: key, Value: value})
			if err != nil {
				er := ofile.Close()
				if er != nil {
					return er
				}
				return err
			}
		}
		err := ofile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func CallForATask(masterAddr string) *MyReply {
	// 发送请求
	args := MyArgs{
		MsgType: AskForTask,
	}
	reply := MyReply{}
	ok := call(masterAddr, "Master.AskForTask", &args, &reply)
	if !ok {
		fmt.Printf("Call to master for a task failed\n")
		return nil
	}
	return &reply
}

func CallForReportStatus(masterAddr string, successType MsgType, taskID int) {
	// 报告任务执行情况
	args := MyArgs{
		MsgType: successType,
		TaskID:  taskID,
	}

	ok := call(masterAddr, "Master.NoticeResult", &args, nil)
	if !ok {
		fmt.Printf("call Master.NoticeResult failed.")
	}
}

// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(masterAddr string, rpcname string, args interface{}, reply interface{}) bool {
	if masterAddr == "" {
		masterAddr = "127.0.0.1:1234"
	}

	c, err := rpc.DialHTTP("tcp", masterAddr)
	// sockname := masterSock()
	// c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
