package scheduler

// Task has process which going to be run if Occurrence for given date is
// positive. Shutdown is fired when worker is terminated
type Task func()
