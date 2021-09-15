package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

//在register时候需要获得的
type RegisterArgs struct {
}

type RegisterReply struct {
	ID int
}

type RequestArgs struct {
	//发请求的时候告知ID
	ID int
}

type RequestReply struct {
	ReceivedTask *Task
	//如果是map,该值为reduce num,需要reduce num来划分buckets，按照key进行hash
	//写入中间文件,让对应的reduce来处理

	//如果是reduce，则该值对应所有可能的的map task ID，这样可以列举去查看所有map给reduce的中间文件
	NMap_Reduce int
	//如果是map,该值为mapID，用来写中间文件的名字
	//如果是reduce，该值为reduceID
	Map_Reduce_ID int
}

type NotifyTaskDoneArgs struct {
	TaskType      int
	TaskID        int
	WorkerID      int
	TmpFileName   []string
	InterFileName []string // 当reduce任务完成时，这个参数表示finalFineName
}

type NotifyTaskDoneReply struct {
}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
