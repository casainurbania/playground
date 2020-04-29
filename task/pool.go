package task

import (
	"fmt"
	"sync"

	"github.com/gomodule/redigo/redis"
)

const (
	TASK_UNSTART = iota
	TASK_WAITING
	TASK_RUNNING
)

// Gorouting instance which can accept client jobs

type worker struct {
	workerPool chan *worker
	jobChannel chan Job
	stop       chan struct{}

	RedisClient redis.Conn
}

type Job func()

type Pool struct {
	JobQueue   chan Job
	dispatcher *dispatcher
	wg         sync.WaitGroup

	RedisClient redis.Conn
}

// Accepts jobs from clients, and waits for first free worker to deliver job
type dispatcher struct {
	workerPool chan *worker
	jobQueue   chan Job
	stop       chan struct{}

	RedisClient redis.Conn
}

func (w *worker) start() {
	go func() {
		var job Job
		for {
			// worker free, add it to pool
			w.workerPool <- w

			select {
			case job = <-w.jobChannel:
				fmt.Println(w.RedisClient.Do("LPOP", "WAITING"))
				fmt.Println(w.RedisClient.Do("LPUSH", "RUNNING", "1"))
				job()
				fmt.Println(w.RedisClient.Do("LPOP", "RUNNING"))
				fmt.Println(w.RedisClient.Do("LPUSH", "DONE", "1"))
			case <-w.stop:
				w.stop <- struct{}{}
				return
			}
		}
	}()
}

func newWorker(pool chan *worker, c redis.Conn) *worker {
	return &worker{
		workerPool:  pool,
		jobChannel:  make(chan Job),
		stop:        make(chan struct{}),
		RedisClient: c,
	}
}

func (d *dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			fmt.Println(d.RedisClient.Do("LPUSH", "WAITING", "1"))

			worker := <-d.workerPool
			worker.jobChannel <- job
		case <-d.stop:
			for i := 0; i < cap(d.workerPool); i++ {
				worker := <-d.workerPool

				worker.stop <- struct{}{}
				<-worker.stop
				fmt.Println(d.RedisClient.Do("LPOP", "WAITING"))
				fmt.Println(d.RedisClient.Do("LPUSH", "STOPED", "1"))

			}

			d.stop <- struct{}{}
			return
		}
	}
}

func newDispatcher(workerPool chan *worker, jobQueue chan Job, c redis.Conn) *dispatcher {
	d := &dispatcher{
		workerPool:  workerPool,
		jobQueue:    jobQueue,
		stop:        make(chan struct{}),
		RedisClient: c,
	}

	for i := 0; i < cap(d.workerPool); i++ {
		worker := newWorker(d.workerPool, c)
		worker.start()
	}

	go d.dispatch()
	return d
}

// Represents user request, function which should be executed in some worker.

// Will make pool of gorouting workers.
// numWorkers - how many workers will be created for this pool
// queueLen - how many jobs can we accept until we block
//
// Returned object contains JobQueue reference, which you can use to send job to pool.
func NewPool(numWorkers int, jobQueueLen int, c redis.Conn) *Pool {
	jobQueue := make(chan Job, jobQueueLen)
	workerPool := make(chan *worker, numWorkers)

	pool := &Pool{
		JobQueue:    jobQueue,
		dispatcher:  newDispatcher(workerPool, jobQueue, c),
		RedisClient: c,
	}
	//  WAITING +1
	fmt.Println(pool.RedisClient.Do("LPUSH", "WAITING", "1"))
	return pool
}

// In case you are using WaitAll fn, you should call this method
// every time your job is done.
//
// If you are not using WaitAll then we assume you have your own way of synchronizing.
func (p *Pool) JobDone() {
	p.wg.Done()
}

// How many jobs we should wait when calling WaitAll.
// It is using WaitGroup Add/Done/Wait
func (p *Pool) WaitCount(count int) {
	p.wg.Add(count)
}

// Will wait for all jobs to finish.
func (p *Pool) WaitAll() {
	p.wg.Wait()
}

// Will release resources used by pool
func (p *Pool) Release() {
	p.dispatcher.stop <- struct{}{}
	<-p.dispatcher.stop
}
