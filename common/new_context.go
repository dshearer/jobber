package common

import (
	"fmt"
	"sync"
)

/*
A context can only be cancelled by a parent context's thread.

A context can only be finished by its own thread.

NOTE: A context's thread must ALWAYS call Finished, even if the context
has been cancelled.
*/
type NewContext struct {
	name           string
	mu             sync.Mutex
	parent         *NewContext
	childAmt       int
	childWaitGroup sync.WaitGroup
	children       map[string]*NewContext
	cancelledChan  chan struct{}
	finishedChan   chan struct{}
	finished       bool
}

/*
Finish this context.

This method marks this context as "finished".  It then waits for all
of its child contexts to finish.  Finally, it notifies its parent context
that it has finished.

If this context is already finished, this method does nothing.

WARNING: Must be called only by this context's own thread.
*/
func (self *NewContext) Finish() {
	if self.finished {
		return
	}

	// mark self as finished
	self.finished = true

	// wait for child contexts
	self.childWaitGroup.Wait()

	// notify parent context
	if self.parent != nil {
		self.parent.childWaitGroup.Done()
		self.parent.forgetChild(self)
	}

	// notify waiters
	close(self.finishedChan)
}

/*
Get whether this context is finished.
*/
func (self *NewContext) Finished() bool {
	return self.finished
}

/*
Cancel this context.

WARNING: Must not be called by this context's thread.

NOTE: Thread-safe
*/
func (self *NewContext) Cancel() {
	self.mu.Lock()

	if self.Cancelled() {
		self.mu.Unlock()
		return
	}

	// cancel child contexts
	for _, child := range self.children {
		child.Cancel()
	}
	self.children = nil

	// cancel self
	close(self.cancelledChan)

	self.mu.Unlock()
}

/*
Get a channel that is closed when this context is cancelled.
*/
func (self *NewContext) CancelledChan() <-chan struct{} {
	return self.cancelledChan
}

/*
Get whether this context has been cancelled.
*/
func (self *NewContext) Cancelled() bool {
	select {
	case <-self.cancelledChan:
		return true
	default:
		return false
	}
}

/*
Wait for this context to finish.

WARNING: Must be called only by this context's parent thread.
*/
func (self *NewContext) Wait() {
	<-self.finishedChan
}

func (self *NewContext) forgetChild(child *NewContext) {
	self.mu.Lock()
	if !self.Cancelled() {
		delete(self.children, child.name)
	}
	self.mu.Unlock()
}

/*
Make a child context of this context.  Returns nil if an error
occurred.

WARNING: Must be called only by this context's own thread.
*/
func (self *NewContext) MakeChild() *NewContext {
	self.mu.Lock()

	if self.Cancelled() || self.Finished() {
		self.mu.Unlock()
		return nil
	}

	// make context
	newCtx := &NewContext{
		name:          fmt.Sprintf("%v.%v", self.name, self.childAmt),
		parent:        self,
		children:      make(map[string]*NewContext),
		cancelledChan: make(chan struct{}),
		finishedChan:  make(chan struct{}),
	}

	self.childAmt++
	self.children[newCtx.name] = newCtx
	self.childWaitGroup.Add(1)

	self.mu.Unlock()
	return newCtx
}

var backgroundCtx *NewContext = &NewContext{
	name:          "0",
	children:      make(map[string]*NewContext),
	cancelledChan: make(chan struct{}),
	finishedChan:  make(chan struct{}),
}

func BackgroundContext() *NewContext {
	return backgroundCtx
}
