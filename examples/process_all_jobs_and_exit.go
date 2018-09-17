// On this example:
//  - 10 workers will be running
//  - 30 jobs will be enqueued to be processed by the workers
//  - workers will exit after processing first 30 jobs

package main

import (
	"log"
	"time"

	"github.com/enriquebris/workerpool"
)

func main() {
	// total workers
	totalWorkers := 10
	// max number of tasks waiting in the channel
	maxNumberJobsInChannel := 15
	// do not show messages about the pool processing
	verbose := false

	pool := workerpool.NewPool(totalWorkers, maxNumberJobsInChannel, verbose)

	// add the worker function
	pool.SetWorkerFunc(func(data interface{}) bool {
		log.Printf("processing %v\n", data)
		// add a 1 second delay (to makes it look as it were processing the job)
		time.Sleep(time.Second)
		log.Printf("processing finished for: %v\n", data)

		// let the pool knows that the worker was able to complete the task
		return true
	})

	// start up the workers
	pool.StartWorkers()

	// add tasks in a separate goroutine
	go func() {
		for i := 0; i < 30; i++ {
			pool.AddTask(i)
		}

		// kill all workers after the current enqueued jobs get processed
		pool.LateKillAllWorkers()
	}()

	// wait while at least one worker is alive
	pool.Wait()
}
