package main

import (
	"fmt"
	"os"
	"time"

	"github.com/casainurbania/playground/task"
	"github.com/gomodule/redigo/redis"

	"github.com/casainurbania/playground/cmd"
	"github.com/casainurbania/playground/conn"
)

var r redis.Conn

func init() {
	var err error
	r, err = redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		panic(err)
	}
	// 示例
	//_, err = r.Do("LPUSH", "UNSTART", 0)
	//_, err = r.Do("LPUSH", "WAIT", 0)
	//_, err = r.Do("LPUSH", "RUNNING", 0)

}

func main() {

	// 队列长度10,worker数量3个
	pool := task.NewPool(3, 10, r)
	defer pool.Release()

	// 等待多少个worker执行结束
	pool.WaitCount(10)

	// 将任务注入队列
	for i := 0; i < 10; i++ {

		pool.JobQueue <- func() {
			time.Sleep(time.Second)
			// 任务执行结束后告诉pool其中一个任务结束了
			defer pool.JobDone()
			myTask()
		}
	}

	// 等待pool中所有任务结束
	pool.WaitAll()
}

// 模拟执行的任务
func myTask() {
	cl, err := conn.NewSSHClient("192.168.10.112", "root", "1qaz!QAZ")
	if err != nil {
		fmt.Println("ssh连接失败: ", err.Error())
		os.Exit(1)
	}
	c := &cmd.RemoteCommand{}
	c.Cmd = "date"
	c.Timeout = time.Second * 25
	c, err = cmd.NewRemoteCmd("date", 5, cl)

	if err != nil {
		fmt.Println(err)
		return
	}
	if err := c.Run(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(c.Stdout())
	fmt.Println(c.Stderr())
}
