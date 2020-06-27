package executors

import (
	"fmt"
	"reflect"
)

//Job 任务对象
type Job struct {
	fun  interface{}
	done chan struct{}
	Err  error
	Result interface{}
}


var fn_without_ret func()
var fn_with_err func() error
var fn_with_2ret func() (interface{}, error)

var fn_without_ret_ref = reflect.TypeOf(fn_without_ret)
var fn_with_err_ref = reflect.TypeOf(fn_with_err)
var fn_with_2ret_ref = reflect.TypeOf(fn_with_2ret)


func newjob(fun interface{}) *Job {
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
			j.Err = fmt.Errorf("catch panic: %v", err)
		}
	}()

	funReflect := reflect.TypeOf(j.fun)
	switch {
	case funReflect.ConvertibleTo(fn_without_ret_ref):
		j.fun.(func())()
	case funReflect.ConvertibleTo(fn_with_err_ref):
		j.Err = j.fun.(func() error)()
	case funReflect.ConvertibleTo(fn_with_2ret_ref):
		j.Result, j.Err = j.fun.(func() (interface{}, error))()
	default:
		j.Err = fmt.Errorf("Job closure must be type of `func()`, `func() error` or `func() (interface {}, error)`")
	}
}

//Wait 同步等待任务执行结束，返回的错误为任务执行过程产生的panic
func (j *Job) Wait(f func()) error {
	select {
	case <-j.done:
		f()
	}
	return j.Err
}
