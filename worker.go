package routinepool

import (
	"time"
)

type Worker struct {
	taskQueue chan Runable

	lastTime       time.Time
	expireInterval time.Duration
	reservedFlag   bool

	destroyChan chan bool
	closeChan   chan struct{}
}

func NewWorker(tq chan Runable, ei time.Duration, rf bool,
	dc chan bool, cc chan struct{}) (w *Worker) {

	w = &Worker{
		taskQueue:      tq,
		expireInterval: ei,
		reservedFlag:   rf,
		destroyChan:    dc,
		closeChan:      cc}

	return
}

func (w *Worker) Start() {

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case task := <-w.taskQueue:
				if task == nil {
					continue
				}

				//do task
				task.Run()
				//update lastTime
				w.lastTime = time.Now()

			case <-ticker.C:
				if (!w.reservedFlag) && (time.Now().Sub(w.lastTime) > w.expireInterval) {

					w.destroyChan <- true
					return
				}

			case <-w.closeChan:
				return
			}
		}
	}()
}
