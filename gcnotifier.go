// Package gcnotifier provides a way to receive notifications after every time
// garbage collection (GC) runs. This can be useful to instruct your code to
// free additional memory resources that you may be using.
//
// To minimize the load on the GC the code that runs after receiving the
// notification should try to avoid allocations as much as possible, or at the
// very least make sure that the amount of new memory allocated is significantly
// smaller than the amount of memory that has been "freed" by your code.
package gcnotifier

import "runtime"

type sentinel struct {
	gcCh chan struct{}
}

// AfterGC returns a channel that will receive a notification after every GC
// run. If a notification is not consumed before another GC runs only one of the
// two notifications is sent. To stop the notifications you can safely close the
// channel.
//
// A common use case for this is when you have a custom pool of objects: instead
// of setting a maximum size you can leave it unbounded and then drop all or
// some of them after every GC run (e.g. sync.Pool drops all objects during GC).
func AfterGC() chan struct{} {
	s := &sentinel{gcCh: make(chan struct{})}
	runtime.SetFinalizer(s, finalizer)
	return s.gcCh
}

func finalizer(obj interface{}) {
	defer recover() // writing to a closed channel will panic
	s := obj.(*sentinel)
	select {
	case s.gcCh <- struct{}{}:
	default:
	}
	// we get here only if the channel was not closed
	runtime.SetFinalizer(s, finalizer)
}
