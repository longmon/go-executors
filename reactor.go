package executors

import (
	"fmt"
)

//Job 任务对象
type Job struct {
	fun  Fn
	done chan struct{}
	err  error
}

type fn func()

func newjob(fun fn) *Job {
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

//Wait 同步等待任务执行结束，返回的错误为任务执行过程产生的panic
func (j *Job) Wait(f func()) error {
	select {
	case <-j.done:
		f()
	}
	return j.err
}
