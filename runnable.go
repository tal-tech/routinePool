package routinepool

//Runable is the interface for tasks that can be executed by the routinepool

type Runable interface {
	Run()
}
