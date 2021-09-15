package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

//标识task的状态TaskState
//分别为idle=0，running=1，finished=2
const (
	IDLE int = iota
	RUNNING
	FINISHED
)

//标识task的种类TaskType
//none task是说现在所有的task都分配出去了，但是还没有做完
//此时有worker来要task，为了防止有节点出错，设置再过一段时间再来要

//exit task是指coordinator会给节点发送一个falsetask，告知worker任务完成，退出即可
const (
	MAPTASK int = iota
	REDUCETASK
	NONETASK
	EXITTASK
)

//TaskPhase:

const (
	//只分发map Task
	MAPPHASE int = iota
	//只分发reduce
	REDUCEPHASE
	//结束阶段
	EXITPHASE
)

//go中首字母大写才能被别的包访问
type Task struct {
	//用于存储task的类型map,reduce
	TaskType int
	//map对应文件，一个文件对应一个map task
	Filename string
	//记录fail掉的work id
	FailedWorks []int
	TaskState   int
	//Task的状态，分别是idle,running,finished
	//记录fail掉的work id

}

type Coordinator struct {
	workerNums int //给workers编号，这样map和reduce task有自己的编号，

	//另外map需要将文件分为nReduce个

	mapTask    []Task //存储map的队列
	reduceTask []Task //存储reduce的队列
	mutex      sync.Mutex
	nReduce    int //记录reduce个数，这样可以分成nReduce个bucket
	taskPhase  int //记录当前应该分发哪种任务

}

// Your code here -- RPC handlers for the worker to call.

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (m *Coordinator) WorkerRegister(args *RegisterArgs, reply *RegisterReply) error {
	//这里为什么要锁住？
	//因为每个线程 过来都需要对共享的数据进行处理
	m.mutex.Lock()
	defer m.mutex.Unlock()
	reply.ID = m.workerNums
	m.workerNums++
	return nil
}

//申请task的时候Coordinator要进行分配
func isInSlice(arr *[]int, id int) bool {
	for _, v := range *arr {
		if v == id {
			return true
		}
	}
	return false
}

//返回任务,如果是map返回nreduce和mapTaskID
//如果是reduce返回mapnums和reduceTaskID
func (m *Coordinator) getTask(workerID int) (*Task, int, int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	//根据现在的Coordinator处于的状态分配Task
	switch m.taskPhase {
	case MAPPHASE:
		//记录是否所有map都完成
		allMapFinished := true
		for i := 0; i < len(m.mapTask); i++ {
			if m.mapTask[i].TaskState != FINISHED {
				allMapFinished = false
			}
			//空闲并且workid不在failed里面
			if m.mapTask[i].TaskState == IDLE && (!isInSlice(&m.mapTask[i].FailedWorks, workerID)) {
				m.mapTask[i].TaskState = RUNNING
				//返回任务, Reduce数量和mapTaskID
				return &m.mapTask[i], m.nReduce, i
			}
		}
		if allMapFinished {
			m.taskPhase = REDUCEPHASE
		} else {
			// map任务还没有执行完，但是上面没有返回，说明要等待
			return &Task{NONETASK, "", nil, IDLE}, m.nReduce, -1
		}
	case REDUCEPHASE:
		//map未结束 reduce未开始的phase
		allReduceFinished := true
		for i := 0; i < len(m.reduceTask); i++ {
			if m.reduceTask[i].TaskState != FINISHED {
				allReduceFinished = false
			}
			if m.reduceTask[i].TaskState == IDLE && (!isInSlice(&m.reduceTask[i].FailedWorks, workerID)) {
				m.reduceTask[i].TaskState = RUNNING
				//这里返回task,mapTask的数量和reduceID
				return &m.reduceTask[i], len(m.mapTask), i
			}
		}
		if allReduceFinished {
			m.taskPhase = EXITPHASE
		} else {
			//ReduceTask没执行完
			return &Task{NONETASK, "", nil, IDLE}, len(m.mapTask), -1
		}
		//强制执行下面的代码
		fallthrough
	case EXITPHASE:
		return &Task{EXITTASK, "", nil, IDLE}, len(m.mapTask), -1
	}
	return &Task{EXITTASK, "", nil, IDLE}, len(m.mapTask), -1
}

//这个函数Coordinator会在分发结束的时候产生一个线程，这个线程过十秒，会去查看一下对应task的状态
//如果超过十秒没有完成，会将workerID加入fail里面
func (m *Coordinator) exceedTime(task *Task, workerID int) {
	time.Sleep(10 * time.Second)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if task.TaskState != FINISHED {
		task.FailedWorks = append(task.FailedWorks, workerID)
	}
}

//请求task
func (m *Coordinator) RequestTask(args *RequestArgs, reply *RequestReply) error {
	reply.ReceivedTask, reply.NMap_Reduce, reply.Map_Reduce_ID = m.getTask(args.ID)
	if reply.ReceivedTask.TaskType == MAPTASK || reply.ReceivedTask.TaskType == REDUCETASK {
		go m.exceedTime(reply.ReceivedTask, args.ID)
	}
	return nil
}

func (m *Coordinator) NotifyMasterTaskDone(args *NotifyTaskDoneArgs, reply *NotifyTaskDoneReply) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	switch args.TaskType {
	case MAPTASK:
		//不在失败的里面
		if !isInSlice(&m.mapTask[args.TaskID].FailedWorks, args.WorkerID) {
			for i := 0; i < len(args.TmpFileName); i++ {
				os.Rename(args.TmpFileName[i], args.InterFileName[i])
			}
			m.mapTask[args.TaskID].TaskState = FINISHED
		} else {
			//删除临时文件
			for _, filename := range args.TmpFileName {
				os.Remove(filename)
			}
		}
	case REDUCETASK:
		if !isInSlice(&m.reduceTask[args.TaskID].FailedWorks, args.WorkerID) {
			for i := 0; i < len(args.TmpFileName); i++ {
				os.Rename(args.TmpFileName[i], args.InterFileName[i])
			}
			m.reduceTask[args.TaskID].TaskState = FINISHED
		} else {
			//删除临时文件
			for _, filename := range args.TmpFileName {
				os.Remove(filename)
			}
		}
	}
	return nil
}

//RPC机制是通过 这个函数实现的，修改reply返回给worker
/*
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}
*/
//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.taskPhase == EXITTASK {
		ret = true
	}
	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
//nReduce代表reduce的个数，map根据这一点将中间key划分成nreduce个bucket

func (m *Coordinator) initial(files []string, nReduce int) {
	m.workerNums = 0
	m.mutex = sync.Mutex{}
	m.nReduce = nReduce
	m.taskPhase = MAPPHASE

	for _, filename := range files {
		task := Task{MAPTASK, filename, []int{}, IDLE}
		m.mapTask = append(m.mapTask, task)
	}

	for i := 0; i < nReduce; i++ {
		task := Task{REDUCETASK, "", []int{}, IDLE}
		m.reduceTask = append(m.reduceTask, task)
	}

}

func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	c.initial(files, nReduce)
	c.server()
	return &c
}
