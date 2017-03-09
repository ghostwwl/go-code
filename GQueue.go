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

func (this *Queue) TaskDone() {
	this.AllTasksDone.L.Lock()
	unfinished := this.UnfinishedTasks - 1
	if unfinished <= 0 {
		if unfinished < 0 {
			panic("called too many times")
		}
		this.AllTasksDone.Broadcast()
		this.UnfinishedTasks = unfinished
	}
	this.AllTasksDone.L.Unlock()
}

func (this *Queue) Join() {
	this.AllTasksDone.L.Lock()
	for {
		if this.UnfinishedTasks > 0 {
			this.AllTasksDone.Wait()
		} else {
			break
		}
	}
	this.AllTasksDone.L.Unlock()
}

func (this *Queue) Qsize() int {
	this.Mutex.Lock()
	n := this.List.Len()
	this.Mutex.Unlock()
	return n
}

func (this *Queue) IsEmpty() bool {
	if this.Qsize() > 0{
		return false
	} 
	return true
}

func (this *Queue) IsFull() bool {
	this.Mutex.Lock()
	n := this.List.Len()
	r := n > 0 && uint32(n) == this.MaxSize
	this.Mutex.Unlock()
	return r
}

func (this *Queue) Put(item interface{}) {
	this.NotFull.L.Lock()
	if this.MaxSize > 0 {
		if uint32(this.List.Len()) == this.MaxSize {
			this.NotFull.Wait()
		}
	}
	this.List.PushBack(item)
	this.UnfinishedTasks++
	this.NotEmpty.Signal()
	this.NotFull.L.Unlock()
}

func (this *Queue) Get() interface{} {
	this.NotEmpty.L.Lock()
	if this.List.Len() <= 0 {
		this.NotEmpty.Wait()
	}
	item := this.List.Front()
	this.List.Remove(item)
	this.NotFull.Signal()
	this.NotEmpty.L.Unlock()

	return item.Value
}
