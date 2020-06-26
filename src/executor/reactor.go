package executor

import (
	"fmt"
	"log"
)

type Job struct {
	fun  Fn
	done chan struct{}
	err  error
}

type Fn func()

func newjob(fun Fn) *Job {
	return &Job{
		fun:  fun,
		done: make(chan struct{}, 1),
	}
}

func (j *Job) exec() {
	defer func() {
		j.done <- struct{}{}
		close(j.done)
	}()
	defer func() {
		if err := recover(); err != nil {
			j.err = fmt.Errorf("catch panic: %v", err)
		}
	}()

	j.fun()
}

//Wait wait for job done
func (j *Job) Wait(f func()) error {
	select {
	case <-j.done:
		log.Println("wait func")
		f()
	}
	return j.err
}
