package common

import (
	"context"
	"fmt"
	"sync"
	"time"
)

/*
An extension of context.Context that adds the ability to wait for
child goroutines to finish.

Let P be a goroutine that spawns another goroutine C, and assume that
P creates a context and passes it to C.  As with context.Context, P
may cancel the context, informing C that it should soon finish up.
In addition, P may "wait" on the context, which causes P to block
until C has finished.  For this to work, C must in all circumstances
tell the context that it has finished.

In summary, these are the valid operations on a BetterContext:

ROLE			VALID OPERATIONS
------------------------------
Parent		Cancel, WaitForFinish
Child		Finish, WaitForChildren

If C spawns additional goroutines with their own subcontexts, C
naturally should not be considered finished until all of those
goroutines are too.  So, when C calls Finish on the original context,
it will block until all the subcontexts are finished.  Thus P is able
to wait on not only C but also all of C's spawned goroutines.

Example of a channel-merge function:

    func merge(ctx BetterContext, cs ...<-chan int) <-chan int {
        out := make(chan int)

        // spawn goroutine for each input channel
        output := func(ctx BetterContext, c <-chan int) {
            defer ctx.Finish()
            for {
                // read
                var n int
                select {
                case n, ok = <-c:
                    if !ok {
                        // c was closed
                        return
                    }
                case <-ctx.Done():
                    return
                }

                // write
                select {
                case out <- n:
                case <-ctx.Done():
                    return
                }
            }
        }
        for _, c := range cs {
            subctx, _ := MakeChildContext(ctx)
            output(subctx, c)
        }

        go func() {
            // wait for them to finish
            ctx.Finish()
            close(out)
        }()

        return out
    }
*/
type BetterContext interface {
	context.Context

	/*
	   Finish this context.

	   This method marks this context as "finished".  It then waits
	   for all of its child contexts to finish.  Finally, it notifies
	   its parent context that it has finished.

	   If this context is already finished, this method does nothing.
	*/
	Finish()

	/*
		Wait for the child contexts of this context to finish.
	*/
	WaitForChildren()
}

type BetterContextCtl struct {
	Cancel        context.CancelFunc
	WaitForFinish func()
}

type betterContextImpl struct {
	mu             sync.Mutex
	name           string
	parent         context.Context
	childWaitGroup sync.WaitGroup
	childAmt       int
	children       map[string]*betterContextImpl
	cancelledChan  chan struct{}
	finishedChan   chan struct{}
	finished       bool
	cancelledErr   error
	deadline       *time.Time
	deadlineTimer  *time.Timer
}

/*
Make a child context of this context.  Returns nil if this context is
already finished.
*/
func newBetterContextImpl(parent context.Context) *betterContextImpl {
	// make child context
	newCtx := &betterContextImpl{
		parent:        parent,
		children:      make(map[string]*betterContextImpl),
		cancelledChan: make(chan struct{}),
		finishedChan:  make(chan struct{}),
	}

	switch p := parent.(type) {
	case *betterContextImpl:
		p.mu.Lock()
		defer p.mu.Unlock()

		if p.cancelledErr != nil || p.finished {
			return nil
		}

		// set child's name
		newCtx.name = fmt.Sprintf("%v.%v", p.name, p.childAmt)

		// adjust children in parent
		p.childAmt++
		p.children[newCtx.name] = newCtx
		p.childWaitGroup.Add(1)
		fmt.Printf("ctx %v: Adding child: %v\n", p.name, newCtx.name)

		return newCtx

	default:
		// set child's name
		newCtx.name = fmt.Sprintf("gen-parent.0")

		// propagate parent's cancellation
		if p.Done() != nil {
			go func() {
				select {
				case <-p.Done():
					newCtx.cancel(p.Err())
				case <-newCtx.Done():
				}
			}()
		}

		return newCtx
	}
}

func (self *betterContextImpl) cancel(reason error) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.cancelledErr != nil {
		return
	}

	// kill deadline timer
	if self.deadlineTimer != nil {
		self.deadlineTimer.Stop()
	}

	// cancel child contexts
	for _, child := range self.children {
		child.cancel(reason)
	}
	self.children = nil

	// cancel self
	self.cancelledErr = reason
	close(self.cancelledChan)
}

func (self *betterContextImpl) makeControl() BetterContextCtl {
	return BetterContextCtl{
		Cancel:        func() { self.cancel(context.Canceled) },
		WaitForFinish: func() { <-self.finishedChan },
	}
}

func (self *betterContextImpl) Deadline() (deadline time.Time, ok bool) {
	if self.deadline == nil {
		if self.parent == nil {
			return time.Time{}, false
		} else {
			return self.parent.Deadline()
		}
	} else {
		return *self.deadline, true
	}
}

func (self *betterContextImpl) Done() <-chan struct{} {
	return self.cancelledChan
}

func (self *betterContextImpl) Err() error {
	return self.cancelledErr
}

func (self *betterContextImpl) Value(key interface{}) interface{} {
	// TODO
	return nil
}

func (self *betterContextImpl) WaitForChildren() {
	fmt.Printf("Waiting for %v children\n", len(self.children))
	self.childWaitGroup.Wait()
}

/*
Finish this context.

This method marks this context as "finished".  It then waits for all
of its child contexts to finish.  Finally, it notifies its parent
context that it has finished.

If this context is already finished, this method does nothing.

WARNING: This function must be called only by the goroutine that owns
this context.
*/
func (self *betterContextImpl) Finish() {
	if self.finished {
		return
	}

	fmt.Printf("ctx %v: finishing\n", self.name)

	// mark self as finished
	self.finished = true

	// wait for child contexts
	self.WaitForChildren()

	if self.parent != nil {
		if p, ok := self.parent.(*betterContextImpl); ok {
			// notify parent context
			p.mu.Lock()
			fmt.Printf("ctx %v: Removing child %v\n", p.name, self.name)
			p.childWaitGroup.Done()
			delete(p.children, self.name)
			p.mu.Unlock()
		}
	}

	// notify waiters
	close(self.finishedChan)

	fmt.Printf("ctx %v: finished\n", self.name)
}

/*
child will be nil if parent is already cancelled or finished.
*/
func MakeChildContext(parent context.Context) (child BetterContext,
	ctl BetterContextCtl) {
	c := newBetterContextImpl(parent)
	return c, c.makeControl()
}

/*
child will be nil if parent is already cancelled or finished.
*/
func ContextWithDeadline(parent context.Context,
	deadline time.Time) (child BetterContext, ctl BetterContextCtl) {

	// make child context
	c := newBetterContextImpl(parent)
	if c == nil {
		return nil, BetterContextCtl{}
	}

	// check parent's deadline
	if cur, ok := parent.Deadline(); ok && cur.Before(deadline) {
		// parent's deadline is sooner
		return c, c.makeControl()
	}

	// check given deadline
	d := deadline.Sub(time.Now())
	if d <= 0 {
		// deadline has already passed
		c.cancel(context.DeadlineExceeded)
		return c, c.makeControl()
	}

	// set child's deadline
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancelledErr == nil {
		c.deadlineTimer = time.AfterFunc(d, func() {
			c.cancel(context.DeadlineExceeded)
		})
	}
	return c, c.makeControl()
}
