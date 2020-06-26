package executors

import (
	"sync"
)

//worker 工作线程
type worker struct {
	fun   func()
	jobCh <-chan *Job  //只读通道，接收主线程的任务消息
	evtCh chan<- event //只写通道，子线程发生的事件发送给主线程
	msgCh <-chan event //只读通道，主线程通知子线程的通道
}

type event struct {
	evt    evt_type
	worker *worker
}

//workerpool 线程池
//使用map实现的一个集合
type workerpool struct {
	workers map[*worker]struct{}
	mu      *sync.RWMutex
}

func newWorker(e *_executor) *worker {
	w := &worker{
		jobCh: e.jobCh,
		evtCh: e.evtCh,
		msgCh: e.msgCh,
	}

	w.fun = func() {
		defer w.close()
		for {
			select {
			case job := <-w.jobCh:
				w.evtCh <- event{evt: evt_type_running}
				job.exec()
				w.evtCh <- event{evt: evt_type_free}

			case msg := <-w.msgCh:
				if msg.evt == evt_type_terminate {
					return
				}
			}
		}
	}
	go w.fun()
	return w
}

func makeWorkerpool(size int32) *workerpool {
	p := &workerpool{
		mu:      &sync.RWMutex{},
		workers: make(map[*worker]struct{}, size),
	}
	return p
}

func (w *worker) close() {
	w.evtCh <- event{
		evt:    evt_type_terminated,
		worker: w,
	}
}

func (wp *workerpool) remove(w *worker) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	delete(wp.workers, w)
}

func (wp *workerpool) addWorker(w *worker) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.workers[w] = struct{}{}
}

func (wp *workerpool) addWorkerWithoutLock(w *worker) {
	wp.workers[w] = struct{}{}
}

func (wp *workerpool) len() int {
	return len(wp.workers)
}
