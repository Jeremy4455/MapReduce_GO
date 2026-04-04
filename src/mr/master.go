package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

type taskStatus int

const (
	idle taskStatus = iota
	running
	finished
	failed
)

type MapTaskInfo struct {
	TaskId    int
	Status    taskStatus
	StartTime int64
}

type ReduceTaskInfo struct {
	Status    taskStatus
	StartTime int64
}

type Master struct {
	NReduce       int
	MapTasks      map[string]*MapTaskInfo
	MapSuccess    bool
	muMap         sync.Mutex
	ReduceTasks   []*ReduceTaskInfo
	ReduceSuccess bool
	muReduce      sync.Mutex
	// Your definitions here.
}

// start a thread that listens for RPCs from worker.go
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	// sockname := masterSock()
	// os.Remove(sockname)
	// l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrmaster.go calls Done() periodically to find out if the entire job has finished.
func (m *Master) Done() bool {
	for _, t := range m.MapTasks {
		if t.Status != finished {
			return false
		}
	}

	for _, t := range m.ReduceTasks {
		if t.Status != finished {
			return false
		}
	}

	log.Println("[MASTER] All tasks finished. System exiting...")
	return true
}

func (m *Master) NoticeResult(args *MyArgs, reply *MyReply) error {
	switch args.MsgType {
	case MapSuccess:
		{
			m.muMap.Lock()
			// defer m.muMap.Unlock()
			for _, v := range m.MapTasks {
				if v.TaskId == args.TaskID && v.Status == running {
					v.Status = finished
					m.muMap.Unlock()
					log.Printf("[MAP] Task %d finished successfully.\n", args.TaskID)
					return nil
				}
			}
		}
	case ReduceSuccess:
		{
			m.muReduce.Lock()
			// defer m.muReduce.Unlock()
			m.ReduceTasks[args.TaskID].Status = finished
			m.muReduce.Unlock()
			log.Printf("[REDUCE] Task %d finished successfully.\n", args.TaskID)
			return nil
		}
	case MapFailed:
		{
			m.muMap.Lock()
			// defer m.muReduce.Unlock()
			log.Printf("[MAP] Task %d reported FAILED by worker.\n", args.TaskID)
			for _, v := range m.MapTasks {
				if v.TaskId == args.TaskID && v.Status == running {
					v.Status = failed
					m.muMap.Unlock()
					return nil
				}
			}
		}
	case ReduceFailed:
		{
			m.muReduce.Lock()
			// defer m.muMap.Unlock()
			log.Printf("[REDUCE] Task %d reported FAILED by worker.\n", args.TaskID)
			if m.ReduceTasks[args.TaskID].Status == running {
				m.ReduceTasks[args.TaskID].Status = failed
			}
			m.muReduce.Unlock()
			return nil
		}
	}
	return nil
}

const maxTimePeriod = 10

func (m *Master) AskForTask(args *MyArgs, reply *MyReply) error {
	if args.MsgType != AskForTask {
		return BadMsgType
	}

	if !m.MapSuccess {
		m.muMap.Lock()

		countMapSuccess := 0
		for fileName, taskInfo := range m.MapTasks {
			alloc := false

			if taskInfo.Status == idle || taskInfo.Status == failed {
				alloc = true
			} else if taskInfo.Status == running {
				curTime := time.Now().Unix()
				if curTime-taskInfo.StartTime > maxTimePeriod {
					log.Printf("[TIMEOUT] Map Task %d timed out, reassigning...\n", taskInfo.TaskId)
					alloc = true
				}
			} else {
				countMapSuccess++
			}

			if alloc {
				reply.MsgType = MapTaskAlloc
				reply.TaskName = fileName
				reply.NReduce = m.NReduce
				reply.TaskID = taskInfo.TaskId

				taskInfo.Status = running
				taskInfo.StartTime = time.Now().Unix()
				log.Printf("[ALLOC] Assigned Map Task %d for file %s\n", taskInfo.TaskId, fileName)
				m.muMap.Unlock()
				return nil
			}
		}
		m.muMap.Unlock()

		if countMapSuccess < len(m.MapTasks) {
			reply.MsgType = Wait
			return nil
		} else {
			m.MapSuccess = true
			log.Println("[PHASE] Map phase completed. Starting Reduce phase...")
		}
	}

	if !m.ReduceSuccess {
		m.muReduce.Lock()

		countReduceSuccess := 0
		for idx, taskInfo := range m.ReduceTasks {
			alloc := false
			if taskInfo.Status == idle || taskInfo.Status == failed {
				alloc = true
			} else if taskInfo.Status == running {
				curTime := time.Now().Unix()
				if curTime-taskInfo.StartTime > maxTimePeriod {
					log.Printf("[TIMEOUT] Reduce Task %d timed out, reassigning...\n", idx)
					alloc = true
				}
			} else {
				countReduceSuccess++
			}

			if alloc {
				reply.MsgType = ReduceTaskAlloc
				reply.TaskID = idx

				taskInfo.Status = running
				taskInfo.StartTime = time.Now().Unix()

				log.Printf("[ALLOC] Assigned Reduce Task %d\n", idx)
				m.muReduce.Unlock()
				return nil
			}
		}

		m.muReduce.Unlock()
		if countReduceSuccess < len(m.ReduceTasks) {
			reply.MsgType = Wait
			return nil
		} else {
			m.ReduceSuccess = true
			log.Println("[PHASE] Reduce phase completed.")
		}
	}

	reply.MsgType = Shutdown
	return nil
}

func (m *Master) initTask(files []string) {
	for idx, file := range files {
		m.MapTasks[file] = &MapTaskInfo{
			TaskId: idx,
			Status: idle,
		}
	}

	for idx := range m.ReduceTasks {
		m.ReduceTasks[idx] = &ReduceTaskInfo{
			Status: idle,
		}
	}
}

// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{
		NReduce:     nReduce,
		MapTasks:    make(map[string]*MapTaskInfo),
		ReduceTasks: make([]*ReduceTaskInfo, nReduce),
	}

	m.initTask(files)

	m.server()
	return &m
}
