package executors

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	JOB_CHAN_CAP         = 32
	EVT_CHAN_CAP         = 16
	MSG_CHAN_CAP         = 16
	JOB_CHAN_LEN_WARNING = 8
)

type evt_type int

const (
	evt_type_terminate evt_type = iota
	evt_type_terminated
	evt_type_running
	evt_type_free
)

type _executor struct {
	cap      int32
	min      int32
	size     int32
	idle     int32
	jobCh    chan *Job
	evtCh    chan event
	msgCh    chan event
	workers  *workerpool
	shutdown bool
}

var exector *_executor
var once = &sync.Once{}

//InitExecutorWithCapacity 初始化线程池
func InitExecutorWithCapacity(min, capacity int32) {
	once.Do(func() {
		exector = &_executor{
			cap:      capacity,
			min:      min,
			size:     min,
			idle:     min,
			jobCh:    make(chan *Job, JOB_CHAN_CAP),
			evtCh:    make(chan event, EVT_CHAN_CAP),
			msgCh:    make(chan event, MSG_CHAN_CAP),
			shutdown: false,
			workers:  makeWorkerpool(min),
		}
		var i int32 = 0
		for ; i < exector.size; i++ {
			w := newWorker(exector)
			exector.workers.addWorkerWithoutLock(w)
		}
		go exector.catchEvent()
		go exector.monitor()
	})
}

//Run 执行任务
func Run(fun Fn) (*Job, error) {
	return exector.add(fun)
}

//Shutdown 准备关闭线程池
func Shutdown() {
	if exector.shutdown {
		return
	}
	ln := exector.workers.len()
	for i := 0; i < ln; i++ {
		exector.msgCh <- event{
			evt: evt_type_terminate,
		}
	}
}

func (e *_executor) catchEvent() {
	for {
		select {
		case evt := <-e.evtCh:
			if evt.evt == evt_type_terminated {
				if evt.worker != nil {
					e.workers.remove(evt.worker)
				}
				atomic.AddInt32(&e.idle, -1)
			}

			if evt.evt == evt_type_running {
				atomic.AddInt32(&e.idle, -1)
			}

			if evt.evt == evt_type_free {
				atomic.AddInt32(&e.idle, 1)
			}
		}
	}
}

func (e *_executor) monitor() {
	go e.scale()
	go e.reduce()
	//调整线程数
	for {
		time.Sleep(time.Second * 10)
		ln := int32(e.workers.len())
		if ln < e.size {
			for i := ln; i < e.size; i++ {
				e.workers.addWorker(newWorker(e))
			}
		}
		if ln > e.size {
			for i := ln; i > e.size; i-- {
				e.msgCh <- event{
					evt: evt_type_terminate,
				}
			}
		}
	}
}

func (e *_executor) scale() {
	jobFullKeepTimes := 0
	for {
		time.Sleep(time.Second * 10)

		if len(e.jobCh) > JOB_CHAN_LEN_WARNING {
			if jobFullKeepTimes < 2 {
				jobFullKeepTimes++
				continue
			}
			add := int32(float32(e.size) * 0.5)
			newSize := e.size + add
			if newSize > e.cap {
				newSize = e.cap
			}
			atomic.AddInt32(&e.idle, add)
			e.size = newSize
		}
		jobFullKeepTimes = 0
	}
}

func (e *_executor) reduce() {
	idleKeepTimes := 0
	for {
		time.Sleep(time.Second * 10)
		if e.idle > e.min && e.idle > int32(float32(e.size)*0.7) {
			if idleKeepTimes < 2 {
				idleKeepTimes++
				continue
			}
			redu := int32(float32(e.size) * 0.3)
			nsize := e.size - redu
			if nsize < e.min {
				nsize = e.min
			}
			e.size = nsize
			atomic.AddInt32(&e.idle, -redu)
			idleKeepTimes = 0
		} else {
			idleKeepTimes = 0
		}
	}
}

func (e *_executor) add(fun Fn) (*Job, error) {
	if e.shutdown {
		return nil, fmt.Errorf("executor is going down")
	}
	job := newjob(fun)
	e.jobCh <- job
	return job, nil
}
