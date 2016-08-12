# Summary:

The context package offers some great features that any go programmer can take advantage of to build great apps. However, it can be a little tricky when you first pick it up to understand the ins and outs of how to use it and when to use it.

This talk was one simple goal, to show how the to use the context package in your programs. We'll walk through the context package and how we can use it to make our go programs simpler and easier to manage.

We'll look at examples and real code using context to demonstrate the everyday use cases that we may encounter and how to support them.

# Notes:

Context package developed at google to satisy the need for request scoped contextual data and to provide a standardized way to handle cancellation signals and deadlines.

Provides a way to facilitate across API boundaries to goroutines involved in handling a request

Launched as golang.org/x/net/context

## Context Type

* deadline
* cancellation signal
* request-scoped values
* safe to use concurrently

- Incoming requests to a server should create a context
- Outgoing calls to a server should accept a context

- The chain of function calls between must propagate the context - optionally replacing with a modified copy

copies can be created with:

        WithDeadline
        WIthTimeout
        WithCancel
        WithValue

Programs that use context should follow these rules:

1. Don't store a context in a struct type, pass it as a function arg
        * The context should be the first param, preferrably named ctx

2. Do not pass a nil context, pass context.TODO

3. Use context values for request-scoped data that transits processes and API's not for passing optional params

4. The same context can be passed to multiple goroutines, contexts are safe for simultaneous use

Context Interface:

        // A Context carries a deadline, cancelation signal, and request-scoped values
        // across API boundaries. Its methods are safe for simultaneous use by multiple
        // goroutines.
        type Context interface {
            // Done returns a channel that is closed when this Context is canceled
            // or times out.
            Done() <-chan struct{}

            // Err indicates why this context was canceled, after the Done channel
            // is closed.
            Err() error

            // Deadline returns the time when this Context will be canceled, if any.
            Deadline() (deadline time.Time, ok bool)

            // Value returns the value associated with key or nil if none.
            Value(key interface{}) interface{}
        }


## Parts of the context interface:

        Done()

The Done method returns a channel that acts as a cancellation signal to functions running on behalf of Context
- when the channel is closed the functions should end execution and return

        Err()

The Err method returns an error indicating why the Context was cancelled

A Context does not have a Cancel function, the function receiving the cancel signal is usually not the one sending the signal.
Specifically if a parent operation starts sub-operations the sub-operations should not be able to cancel the parent

The WithCancel function provides a way to cancel a new Context value

        Deadline()

The Deadline function allows functions to determine if tbey should start work or not, if too little time remains before expiration of the deadline they should not, in addition the Deadline allows the function to set a timeout on I/O operations.

        Value()

The Value function allows the context to carry request scoped data.

## Derived Contexts

The context package provides functions that allow new contexts to be derived from existing ones.
These form a tree and if a Context is cancelled, all derived contexts are cancelled.

Background

        // Background returns an empty Context. It is never canceled, has no deadline,
        // and has no values. Background is typically used in main, init, and tests,
        // and as the top-level Context for incoming requests.
        func Background() Context

WithCancel and WithTimeout return contexts that can be cancelled sooner than the parent

The Context associated with an incoming request is typically canceled when the request handler returns.
WithCancel is also useful for canceling redundant requests when using multiple replicas.

        // WithCancel returns a copy of parent whose Done channel is closed as soon as
        // parent.Done is closed or cancel is called.
        func WithCancel(parent Context) (ctx Context, cancel CancelFunc)

        // A CancelFunc cancels a Context.
        type CancelFunc func()

WithTimeout is useful for setting a deadline on requests to backend servers

        // WithTimeout returns a copy of parent whose Done channel is closed as soon as
        // parent.Done is closed, cancel is called, or timeout elapses. The new
        // Context's Deadline is the sooner of now+timeout and the parent's deadline, if
        // any. If the timer is still running, the cancel function releases its
        // resources.
        func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)

WithValue

        // WithValue returns a copy of parent whose Value method returns val for key.
        func WithValue(parent Context, key interface{}, val interface{}) Context


## Context values/data

Context's value handling is completely type unsafe and can't be checked at compile time

It's essentially a `map[interface{}]interface{}`

If you can avoid using context to pass data around then you should.

However request scoped data does make sense to pass around via context, generally data that can only exist during the life of a request.

For example data extracted from headers or cookies, userID's tied to auth information, etc.

## Keys

* unexported key types
* see recent article on this
* avoiding key collisions is important as more libraries, middleware, etc us context to set request data

# Example

















