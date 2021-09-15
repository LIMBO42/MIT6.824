package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"strings"
	"time"
)

//第几个reduce需要将输出写入第几个文件
//需要将中间map文件写入当前的目录
/*
*	job完成后work需要退出，可以利用call的返回值处理，如果work
* 	没能联系到coordinator，可以假设coordinate已经退出(代表任务已完成)
 */

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

type worker struct {
	workID int //存储coordinator分配的ID
	//map接受的输入为 文件名和文件内容，输出keyvalue对
	//不断添加到keyvalue对的数组((k1,v1),(k2,v2))
	mapf func(string, string) []KeyValue

	//map到reduce是对keyvalue数组按key进行排序
	//之后将key相同的values都放到一个数组中
	//按照partition的概念分发到reduce

	//第一个参数为key，第二个为key相同的value构成的数组
	//key,(v1,v2,v3)
	//返回值是对value进行处理得到的结果，以wordcount为例就是sum
	reducef func(string, []string) string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
//将kv值经过hash后分配到对应的reduce
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

//
// main/mrworker.go calls this function.
//

func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.

	// uncomment to send the Example RPC to the coordinator.
	//首先从这里开始 发送RPC给coordinator
	//然后修改coordinator回应一个还没有开始的map task的文件名
	//然后worker读文件调用map，可以参考mrsequential
	//中间文件的命名：mr-X-Y（X是map task，Y是reduce task）
	w := worker{}
	w.mapf = mapf
	w.reducef = reducef
	//发送RPC给coordinator
	//CallExample()
	w.CallRegister()
	w.running()
}

//
// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
//

//先register获取worker的id
func (w *worker) CallRegister() {

	// declare an argument structure.
	args := &RegisterArgs{}

	// declare a reply structure.
	reply := &RegisterReply{}

	// send the RPC request, wait for the reply.
	if ok := call("Coordinator.WorkerRegister", args, reply); !ok {
		log.Fatal("error: register failed")
	}
	w.workID = reply.ID
}

//不断运行，
func (w *worker) running() {
	//不断接受任务
	for {
		args := &RequestArgs{w.workID}
		reply := &RequestReply{}
		//获得coordinator给的信息
		if ok := call("Coordinator.RequestTask", args, reply); !ok {
			log.Fatal("error: request failed")
		}
		//查看任务种类
		switch reply.ReceivedTask.TaskType {
		case MAPTASK:
			w.mapTask(reply)
		case REDUCETASK:
			w.reduceTask(reply)
		//遇到NoneTask的情况，休眠一秒
		case NONETASK:
			time.Sleep(time.Second)
		case EXITTASK:
			os.Exit(0)
		}
	}

}

func (w *worker) mapTask(reply *RequestReply) {
	toDoTask := reply.ReceivedTask
	//这里的reduceNum用来将map的结果按reduce的个数分成reduceNum个bucket
	reduceNum := reply.NMap_Reduce
	mapTaskID := reply.Map_Reduce_ID

	//和sequential里面的一样 读取文件
	filename := toDoTask.Filename
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot open %v", filename)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", filename)
	}
	file.Close()

	//一样的，map返回key value数组
	kva := w.mapf(filename, string(content))
	//将key value arr分成reduceNum个bucket
	//二维数组，一维存bucket
	intermediate := make([][]KeyValue, reduceNum)
	for _, kv := range kva {
		//是按照key的值进行划分成若干buckets
		reducerID := ihash(kv.Key) % reduceNum
		intermediate[reducerID] = append(intermediate[reducerID], kv)
	}

	//临时文件存储 key value对
	tmpFileName := make([]string, reduceNum)
	// 先写成临时文件的形式，之后master检查任务后再修改成真正的中间结果文件
	interFileName := make([]string, reduceNum)

	//遍历buckets
	for num, kvs := range intermediate {
		//中间文件的名字的生成
		tmpStr := fmt.Sprint("mr-%d-%d", mapTaskID, num)
		interFileName[num] = tmpStr

		//先生成临时文件，之后再检查
		//tmpfile函数第二个参数是个pattern，这样写之后某位会以一串随机数字结尾
		file, err := ioutil.TempFile("./", "tmp_map_")
		if err != nil {
			log.Fatal("error: cannot create tmp file")
		}
		tmpFileName[num] = file.Name()
		file.Close()

		//打开临时文件
		fh, err := os.OpenFile(file.Name(), os.O_APPEND|os.O_RDWR, os.ModePerm)
		if err != nil {
			log.Fatal("error: open tempFiles failed.")
		}

		encoder := json.NewEncoder(fh)
		for _, kv := range kvs {
			if err := encoder.Encode(kv); err != nil {
				log.Fatal("error: encoding kv failed")
			}
		}
		fh.Close()
	}

	//map的工作结束，通知coordinator
	//因为worker可能处理多个mapTask和多个reduceTask,所以workerID和mapTaskID不一样
	w.notifyTaskDone(MAPTASK, mapTaskID, w.workID, tmpFileName, interFileName)
}

func (w *worker) reduceTask(reply *RequestReply) {
	//? 为什么需要mapNum？
	//列举所有可能的中间文件，mapID-reduceID
	mapNum := reply.NMap_Reduce
	reduceTaskID := reply.Map_Reduce_ID
	//根据reduceTaskID去encode获得key value对

	interFileNames := []string{}

	//这里的map是key对应所有value
	interMaps := make(map[string][]string)

	for i := 0; i < mapNum; i++ {
		interFileName := fmt.Sprintf("mr-%d-%d", i, reduceTaskID)
		interFileNames = append(interFileNames, interFileName)
		//打开interfilename文件
		file, err := os.Open(interFileName)
		if err != nil {
			log.Fatalf("error: Can't open interFile: %v", interFileName)
		}
		decoder := json.NewDecoder(file)
		for {
			var kv KeyValue
			//json解码，存在kv中
			if err := decoder.Decode(&kv); err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatal(err)
				}
			}
			//这样获取到了keyvalue
			//如果key不存在
			if _, ok := interMaps[kv.Key]; !ok {
				interMaps[kv.Key] = []string{kv.Value}
			} else {
				// 被统计过了
				interMaps[kv.Key] = append(interMaps[kv.Key], kv.Value)
			}
		}

	}
	//mapf 参数为string,[]string
	reduceResult := []string{}
	//传给notifyTaskDone
	tmpFileName := make([]string, 1)
	finalFileName := make([]string, 1)

	//interMaps
	for k, v := range interMaps {
		reduceResult = append(reduceResult, w.reducef(k, v))
	}

	//写入临时文件
	file, err := ioutil.TempFile("./", "tmp_reduce_")
	if err != nil {
		log.Fatalf("error: open tmp_reduce_ failed.")
	}
	wErr := ioutil.WriteFile(file.Name(), []byte(strings.Join(reduceResult, "")), 0600)
	if wErr != nil {
		log.Fatal("error: write reduce result to tmp_reduce_ failed. ")
	}

	tmpFileName[0] = file.Name()
	finalFileName[0] = fmt.Sprintf("mr-out-%d", reduceTaskID)

	//到这里reduce完成，询问coordinator要不要处理文件名做最后的结果
	w.notifyTaskDone(REDUCETASK, reduceTaskID, w.workID, tmpFileName, finalFileName)
	// 如果执行到这，说明本reduce在之前没有crach，因此本reduce要处理的中间结果可以删除
	for _, fn := range interFileNames {
		os.Remove(fn)
	}
}

//完成map/reduce之后通知coordinator，因此需要RPC call
func (w *worker) notifyTaskDone(taskType int, taskID int, workerID int, tmpFileName []string, interFileName []string) {
	args := &NotifyTaskDoneArgs{taskType, taskID, workerID, tmpFileName, interFileName}
	reply := &NotifyTaskDoneReply{}
	if ok := call("Coordinator.NotifyMasterTaskDone", args, reply); !ok {
		log.Fatal("error: notify master failed.")
	}
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
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
