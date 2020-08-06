package routinepool

import (
	"sync"
	"time"
)

//routinePool represents a go routine pool manager
type RoutinePool struct {
	mainLock sync.RWMutex

	TaskQueueSize    int64         // task queue
	WaitInterval     time.Duration //queue wait duration,when queue is full
	NormalWorkerSize int64
	MaxWorkerSize    int64
	ExpireInterval   time.Duration //worker max idle duration

	WorkerSize int64

	TaskQueue         chan Runable
	workerDestroyChan chan bool
	closeChan         chan struct{}
	closed            bool
}

//create a fixed routine pool
func NewFixedRoutinePool(taskQueueSize int64, waitInterval time.Duration,
	normalWorkerSize int64) (rp *RoutinePool) {

	rp = newRoutinePool(taskQueueSize, waitInterval, normalWorkerSize, normalWorkerSize)
	return
}

//create a cached routine pool
func NewCachedRoutinePool(taskQueueSize int64, waitInterval time.Duration,
	normalWorkerSize int64, maxWorkerSize int64) (rp *RoutinePool) {

	rp = newRoutinePool(taskQueueSize, waitInterval, normalWorkerSize, maxWorkerSize)
	return
}

//create a new routine pool
func newRoutinePool(taskQueueSize int64, waitInterval time.Duration, normalWorkerSize int64,
	maxWorkerSize int64) (rp *RoutinePool) {
	rp = &RoutinePool{
		TaskQueueSize:    taskQueueSize,
		WaitInterval:     waitInterval,
		NormalWorkerSize: normalWorkerSize,
		MaxWorkerSize:    maxWorkerSize,
		ExpireInterval:   time.Second * 60}

	rp.TaskQueue = make(chan Runable, taskQueueSize)
	rp.workerDestroyChan = make(chan bool)
	rp.closeChan = make(chan struct{})

	go rp.workerMonitor()
	return
}

//put a task into the routinepool
func (rp *RoutinePool) Execute(task Runable) bool {
	rp.mainLock.RLock()
	if rp.closed {
		rp.mainLock.RUnlock()
		return false
	}

	if rp.WorkerSize < rp.NormalWorkerSize {

		worker := NewWorker(rp.TaskQueue, rp.ExpireInterval, true, rp.workerDestroyChan, rp.closeChan)
		worker.Start()
		rp.WorkerSize++

	} else if (rp.WorkerSize < rp.MaxWorkerSize) && (len(rp.TaskQueue) >= int(rp.TaskQueueSize)) {

		worker := NewWorker(rp.TaskQueue, rp.ExpireInterval, false, rp.workerDestroyChan, rp.closeChan)
		worker.Start()
		rp.WorkerSize++
	}
	rp.mainLock.RUnlock()

	timer := time.NewTimer(rp.WaitInterval)
	defer timer.Stop()

	select {
	case rp.TaskQueue <- task:
		return true
	case <-timer.C:
		return false
	}

	return false
}

type innerTask struct{ call func() }

func (t *innerTask) Run() {
	t.call()
}

//put a Func into the routinepool
func (rp *RoutinePool) ExecuteFunc(call func()) bool {
	t := new(innerTask)
	t.call = call

	return rp.Execute(t)
}

//monitor routinepool worker
func (rp *RoutinePool) workerMonitor() {
	var isDestroy bool

	for {
		select {
		case isDestroy = <-rp.workerDestroyChan:

			if !isDestroy {
				continue
			}

			rp.mainLock.Lock()
			rp.WorkerSize--
			rp.mainLock.Unlock()

			//fmt.Printf("worker size:%d task queue:%d\n", rp.WorkerSize, len(rp.TaskQueue))
		case <-rp.closeChan:

			return
		}
	}
}

//release resource
func (rp *RoutinePool) Close() {
	rp.mainLock.Lock()
	defer rp.mainLock.Unlock()
	if rp.closed {
		return
	}

	close(rp.closeChan)
	close(rp.workerDestroyChan)
	close(rp.TaskQueue)
}
