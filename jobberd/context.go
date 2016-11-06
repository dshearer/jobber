package main

import (
	"fmt"
	"github.com/dshearer/jobber/Godeps/_workspace/src/golang.org/x/net/context"
	"sync"
	"time"
)

type JobberCancelFunc func() <-chan struct{}
type JobberWaitFunc func()

type JobberContext struct {
	Name           string
	impl           context.Context
	finishedChan   chan struct{}
	finished       bool
	childAmt       int
	childWaitGroup sync.WaitGroup
	parent         *JobberContext
	mutex          sync.RWMutex
}

type JobberCtl struct {
	Cancel JobberCancelFunc
	Wait   JobberWaitFunc
}

func (ctx *JobberContext) Deadline() (deadline time.Time, ok bool) {
	return ctx.impl.Deadline()
}

func (ctx *JobberContext) Done() <-chan struct{} { // must be called only from subthread
	return ctx.impl.Done()
}

func (ctx *JobberContext) Err() error { // thread-safe
	return ctx.impl.Err()
}

func (ctx *JobberContext) Value(key interface{}) interface{} {
	return ctx.impl.Value(key)
}

func (ctx *JobberContext) Finish() { // must be called only from subthread
	ctx.mutex.Lock()

	if !ctx.finished {
		ctx.finished = true
		ctx.mutex.Unlock()

		/* c.childWaitGroup will never again be incremented. */

		// wait for children to finish
		ctx.childWaitGroup.Wait()

		// announce that we've finished
		close(ctx.finishedChan)
		if ctx.parent != nil {
			ctx.parent.childWaitGroup.Done()
		}
	} else {
		ctx.mutex.Unlock()
	}
}

var background *JobberContext = &JobberContext{
	Name:         "0",
	impl:         context.Background(),
	finishedChan: make(chan struct{}),
	finished:     false,
	parent:       nil,
}

func BackgroundJobberContext() *JobberContext {
	return background
}

func NewJobberContext(parent *JobberContext) (*JobberContext, JobberCtl) {
	parent.mutex.RLock()
	defer parent.mutex.RUnlock()

	// make context
	newCtxName := fmt.Sprintf("%v.%v", parent.Name, parent.childAmt)
	newImpl, newImplCancel := context.WithCancel(parent.impl)
	var newCtx *JobberContext = &JobberContext{
		Name:         newCtxName,
		impl:         newImpl,
		finishedChan: make(chan struct{}),
		finished:     false,
		parent:       parent,
	}

	// make control struct
	var newCtl JobberCtl

	// make control funcs
	newCtl.Cancel = func() <-chan struct{} {
		newImplCancel()
		return newCtx.finishedChan
	}
	newCtl.Wait = func() {
		<-newCtx.finishedChan
	}

	// update parent
	parent.childAmt++
	if parent.finished {
		newCtx.finished = true
		close(newCtx.finishedChan)
	} else {
		parent.childWaitGroup.Add(1)
	}

	return newCtx, newCtl
}
