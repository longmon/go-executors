package executors

import (
	"fmt"
	"sync/atomic"
)

//Fn 任务闭包
type Fn func() error

//Job 任务对象
type Job struct {
	fun     Fn
	done    chan struct{}
	Err     error
	Panic   error
	waiters int32
}

func newjob(fun Fn) *Job {
	return &Job{
		fun:  fun,
		done: make(chan struct{}, 8),
	}
}

func (j *Job) exec() {
	defer func() {
		var i int32
		for i = 0; i < j.waiters; i++ {
			j.done <- struct{}{}
		}
		close(j.done)
	}()
	defer func() {
		if err := recover(); err != nil {
			j.Panic = fmt.Errorf("catch panic: %v", err)
		}
	}()

	j.Err = j.fun()
}

//Wait 同步等待任务执行结束，返回的错误为任务执行过程产生的panic
func (j *Job) Wait(f func()) error {
	atomic.AddInt32(&j.waiters, 1)
	select {
	case <-j.done:
		f()
	}
	return j.Err
}

//Notify 任务调用结束后的异步通知
//闭包参数本次任务对象
func (j *Job) Notify(f func(j *Job)) {
	atomic.AddInt32(&j.waiters, 1)
	go func() {
		select {
		case <-j.done:
			f(j)
		}
	}()
}
