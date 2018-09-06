package workerpool

import (
	"log"

	"github.com/pkg/errors"
)

type PoolFunc func(interface{}) bool

type Pool struct {
	// function to be executed by the workers
	fn PoolFunc
	// initial number of workers (at initialization moment)
	initialWorkers int
	// total live workers
	totalWorkers int
	// tells whether the initialWorkers were started
	workersStarted bool
	// tells the workers: do not accept / process new jobs
	doNotProcess bool

	// total needed job successes to finish WaitUntilNSuccesses(...)
	totalWaitUntilNSuccesses int
	// channel to send the "done" to WaitUntilNSuccesses(...)
	channelWaitUntilNSuccesses chan bool

	// how many workers succeeded
	fnSuccessCounter int
	// how many workers failed
	fnFailCounter int
	// channel to send jobs, workers listen to this channel
	jobsChan chan interface{}
	// channel to keep track of how many workers are up
	totalWorkersChan chan int
	// channel to keep track of succeeded / failed jobs
	fnSuccessChan chan bool

	// log steps
	verbose bool
}

func NewPool(initialWorkers int, maxJobsInChannel int, verbose bool) *Pool {
	ret := &Pool{}

	ret.initialize(initialWorkers, maxJobsInChannel, verbose)

	return ret
}

func (st *Pool) initialize(initialWorkers int, maxJobsInChannel int, verbose bool) {
	st.jobsChan = make(chan interface{}, maxJobsInChannel)
	st.totalWorkersChan = make(chan int, 100)
	// the package will cause deadlock if st.fnSuccessChan is full
	st.fnSuccessChan = make(chan bool, maxJobsInChannel)

	// the workers were not started at this point
	st.workersStarted = false

	st.initialWorkers = initialWorkers

	st.verbose = verbose

	// goroutine that controls the active workers successes / fails
	go st.fnSuccessListener()

	// goroutine that controls the active workers counter
	go st.workerListener()

	// channel to send the "done" to WaitUntilNSuccesses(...)
	st.channelWaitUntilNSuccesses = make(chan bool)
}

// workerListener listens to the workers up/down && keep track of the up workers (st.totalWorkers)
func (st *Pool) workerListener() {

	for waitN := range st.totalWorkersChan {

		st.totalWorkers += waitN

		// add a new worker
		if waitN > 0 {

			// execute the worker function
			go st.workerFunc(st.totalWorkers)
		}

		// check whether all workers were started
		if st.totalWorkers == st.initialWorkers {
			// the workers were started
			st.workersStarted = true
		}
	}
}

// fnSuccessListener listens to the workers successes & fails
func (st *Pool) fnSuccessListener() {
	for fnSuccess := range st.fnSuccessChan {

		if fnSuccess {
			st.fnSuccessCounter++
			if st.verbose {
				log.Printf("fnSuccessCounter: %v workers: %v\n", st.fnSuccessCounter, st.totalWorkers)
			}

			if st.totalWaitUntilNSuccesses > 0 && st.fnSuccessCounter >= st.totalWaitUntilNSuccesses {
				st.channelWaitUntilNSuccesses <- true
			}
		} else {
			st.fnFailCounter++
		}
	}
}

// Wait waits until all workers are not up and running
func (st *Pool) Wait() {
	for {
		if st.workersStarted && st.totalWorkers == 0 {
			return
		}
	}
}

func (st *Pool) WaitUntilNSuccesses(n int) {
	st.totalWaitUntilNSuccesses = n

	for range st.channelWaitUntilNSuccesses {
		break
	}

	// tell workers: do not accept / process new jobs && no new jobs can be accepted
	st.doNotProcess = true

	if st.verbose {
		log.Printf("WaitUntilNSuccesses: %v . kill all workers: %v\n", st.fnSuccessCounter, st.totalWorkers)
	}

	// kill all active workers
	st.KillAllWorkers()

	// wait until all workers were stopped
	st.waitUntilNWorkers(0)

	// tell workers: you can accept / process new jobs && start accepting new jobs
	st.doNotProcess = false
}

// waitUntilNWorkers waits until ONLY n workers are up and running
func (st *Pool) waitUntilNWorkers(total int) {

	for st.totalWorkers != total {
	}

	return
}

// SetWorkerFunc sets the worker's function
func (st *Pool) SetWorkerFunc(fn PoolFunc) {
	st.fn = fn
}

// StartWorkers start workers
func (st *Pool) StartWorkers() {

	for i := 0; i < st.initialWorkers; i++ {
		st.startWorker()
	}
}

// startWorker starts a worker in a separate goroutine
func (st *Pool) startWorker() {
	// increment the active workers counter by 1 ==> start a new worker (st.workerListener())
	st.totalWorkersChan <- 1
}

// AddWorker adds a new worker to the pool
func (st *Pool) AddWorker() {
	st.startWorker()
}

// workerFunc keeps listening to st.jobsChan and executing st.fn(...)
func (st *Pool) workerFunc(n int) {

	// TODO ::: Add non-blocking channel operations ==> https://gobyexample.com/non-blocking-channel-operations

	for taskData := range st.jobsChan {

		// kill worker
		if taskData == nil {

			// decrement the active workers counter by 1
			st.totalWorkersChan <- -1

			if st.verbose {
				log.Printf("[down] worker %v is going to be down", n)
			}

			// kill worker
			return
		}

		if st.doNotProcess {
			// TODO ::: re-enqueue in a different queue/channel/struct
			// re-enqueue the job / task
			st.AddTask(taskData)

		} else {
			// execute the job
			fnSuccess := st.fn(taskData)

			// avoid to cause deadlock
			if !st.doNotProcess {
				// keep track of the job's result
				st.fnSuccessChan <- fnSuccess
			} else {
				// TODO ::: save the job result ...
			}
		}

	}
}

// AddTask adds a new task / job
func (st *Pool) AddTask(data interface{}) error {
	if !st.doNotProcess {
		st.jobsChan <- data
		return nil
	}

	return errors.New("No new jobs are accepted at this moment")
}

// KillWorker kills an inactive / idle worker
func (st *Pool) KillWorker() {
	st.jobsChan <- nil
}

// KillAllWorkers kills all workers.
// If a worker is processing a job, it will not be killed, the pool will wait until it is idle.
func (st *Pool) KillAllWorkers() {
	// get the current "totalWorkers"
	total := st.totalWorkers

	// kill all workers
	for i := 0; i < total; i++ {
		st.KillWorker()
	}
}

// GetTotalWorkers returns the number of active workers.
func (st *Pool) GetTotalWorkers() int {
	return st.totalWorkers
}
