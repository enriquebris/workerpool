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
		// add a 1 second delay (only for the example)
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
	}()

	// wait while at least one worker is active
	pool.Wait()
}
