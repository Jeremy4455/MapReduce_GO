package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"errors"
	"os"
	"strconv"
)

var (
	BadMsgType = errors.New("bad message type")
	NoMoreTask = errors.New("no more task left")
)

type MsgType int

const (
	AskForTask MsgType = iota
	MapTaskAlloc
	ReduceTaskAlloc
	MapSuccess
	ReduceSuccess
	MapFailed
	ReduceFailed
	Shutdown
	Wait
)

// Add your RPC definitions here.

type MyArgs struct {
	MsgType MsgType
	TaskID  int
}

type MyReply struct {
	MsgType  MsgType
	TaskID   int
	NReduce  int
	TaskName string
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
