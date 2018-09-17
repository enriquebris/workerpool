## History

### v0.6

 - Pause / Resume all workers:
   - PauseAllWorkers() 
   - ResumeAllWorkers()
 - Workers will listen to higher priority channels first
 - Workers will listen to broad messages (kill all workers, ...) before get signals from any other channel:
   - KillAllWorkers()
   - KillAllWorkersAndWait()
 - Added function to kill all workers (send a broad message to all workers) and wait until it happens:
   - pool.KillAllWorkersAndWait()
 - Added code examples
   
   
### v0.5

 - Make Wait() listen to a channel (instead of use an endless for loop)
 
### v0.4

 - Sync actions over workers. A FIFO queue was created for the following actions:
   - Add new worker
   - Kill worker(s)
   - Late kill worker(s)
   - Set total workers 
   
### v0.3

 - Added function to adjust number of live workers:
   - pool.SetTotalWorkers(n)
 - Added function to kill all live workers after current jobs get processed:
   - pool.LateKillAllWorkers()
   
### v0.2

 - readme.md
 - godoc
 - code comments

### v0.1

 First stable BETA version.