package ghostlib

/*****************************************
 * FileName : GQueue.go
 * Author   : ghostwwl
 * Note     : 线程/Goroutine 安全的队列呢
 *****************************************/

import (
	"container/list"
	"sync"
)

type Queue struct {
	MaxSize         uint32
	UnfinishedTasks uint32
	Mutex           *sync.Mutex
	NotEmpty        *sync.Cond
	NotFull         *sync.Cond
	AllTasksDone    *sync.Cond
	List            *list.List
}

func NewQueue() *Queue {
	obj := new(Queue)
	obj.MaxSize = 0
	obj.UnfinishedTasks = 0
	obj.Mutex = new(sync.Mutex)
	obj.NotEmpty = sync.NewCond(obj.Mutex)
	obj.NotFull = sync.NewCond(obj.Mutex)
	obj.AllTasksDone = sync.NewCond(obj.Mutex)
	obj.List = list.New()

	return obj
}

func (this *Queue) task_done() {
	this.AllTasksDone.L.Lock()
	defer this.AllTasksDone.L.Unlock()

	unfinished := this.UnfinishedTasks - 1
	if unfinished == 0 {
		this.AllTasksDone.Broadcast()
		this.UnfinishedTasks = unfinished
	}
}

func (this *Queue) join() {
	this.AllTasksDone.L.Lock()
	this.AllTasksDone.L.Unlock()
	for {
		if this.UnfinishedTasks > 0 {
			this.AllTasksDone.Wait()
		} else {
			break
		}
	}
}

func (this *Queue) Qsize() int {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	return this.List.Len()
}

func (this *Queue) IsEmpty() bool {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	return this.List.Len() <= 0
}

func (this *Queue) IsFull() bool {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	n := this.Qsize()
	return n > 0 && uint32(n) == this.MaxSize
}

func (this *Queue) Put(item interface{}) {
	this.NotFull.L.Lock()
	defer this.NotFull.L.Unlock()
	if this.MaxSize > 0 {
		if uint32(this.List.Len()) == this.MaxSize {
			this.NotFull.Wait()
		}
	}
	this.List.PushBack(item)
	this.UnfinishedTasks++
	this.NotEmpty.Signal()
}

func (this *Queue) Get() interface{} {
	this.NotEmpty.L.Lock()
	defer this.NotEmpty.L.Unlock()
	if this.List.Len() <= 0 {
		this.NotEmpty.Wait()
	}
	item := this.List.Front()
	this.List.Remove(item)
	this.NotFull.Signal()

	return item.Value
}
