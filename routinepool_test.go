package routinepool

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

type TaskFixed struct {
	ID   int
	Flag string
}

func (t *TaskFixed) Run() {
	fmt.Printf("task %s %d complete\n", t.Flag, t.ID)
	atomic.AddInt64(&counter, 1)
}

var counter int64

func TestNewFixedRoutinePoolExecute(t *testing.T) {
	//clear counter
	counter = 0

	rp := NewFixedRoutinePool(1, time.Millisecond*10, 5)
	defer rp.Close()

	for i := 1; i <= 999; i++ {
		task := &TaskFixed{ID: i, Flag: "fixed"}
		rp.Execute(task)
	}

	for {
		if counter == 999 {
			break
		}
	}
}

func TestNewFixedRoutinePoolFunc(t *testing.T) {
	//clear counter
	counter = 0

	rp := NewFixedRoutinePool(1, time.Millisecond*10, 5)
	defer rp.Close()

	for i := 1; i <= 999; i++ {
		rp.ExecuteFunc(func() {
			fmt.Printf("fixed Func %d complete\n", i)
			atomic.AddInt64(&counter, 1)
		})
	}

	for {
		if counter == 999 {
			break
		}
	}
}

func TestNewCachedRoutinePool(t *testing.T) {
	//clear counter
	counter = 0

	rp := NewCachedRoutinePool(1, time.Millisecond*10, 5, 50)
	defer rp.Close()

	for i := 1; i <= 999; i++ {
		task := &TaskFixed{ID: i, Flag: "cached"}
		rp.Execute(task)
	}

	for {
		if counter == 999 {
			break
		}
	}
}
