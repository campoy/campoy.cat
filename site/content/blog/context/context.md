+++
date = "2017-04-04T14:06:35-07:00"
title = "The Context Package"
draft = true
+++

<script src="//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.4.0/languages/go.min.js"></script>

Let's talk about the [context](https://golang.org/pkg/context) package!

In this post I'd like to cover how to:

- call functions that receive `Context` values as parameters,
- define functions that receive them,
- avoid wasting resources by using the `context` package in your HTTP server.

This post is a written adaptation of an episode of my YouTube series [justforfunc](https://www.youtube.com/watch?v=LSzR0VEraWw).

<!--more-->

#### A bit of history

Even though the `context` package was added to Go's standard library with Go 1.7, the package has a long history.
It was initially created at Google, where it madurated for a couple of years before being open sourced under
the `golang.org/x/net/context` package.

While the implementation has changed dramatically during its history, the API has been surprisingly stable.

The context package has only around 500 lines of code, so I encourage you to go and read the source code
[here](https://golang.org/src/context/context.go).

#### Cancellation and propagation

The `context` package has a main purpose: cancellation of requests and propagation. Cancellation means that we can notify a server
that the request we sent doesn't need to be fullfilled anymore.

As an example, we could imagine that you asked me to go make a sandwich.

<div style="text-align:center">
<img src="https://imgs.xkcd.com/comics/sandwich.png" alt="make me a sandwich">
<p>Image from <a href="https://xkcd.com/149/">xkcd</a></p>
</div>

Assuming I accepted your request, I would start following a series of steps in order to make the sandwich: buying bread, tomatoes, etc.
At any point you might decide that you don't want that sandwich anymore. When you do that, you're cancelling your request (and being
a bit annoying, tbh).

#### Some code

Some Go code

```go
type Context interface {
	Done() <-chan struct{}
	Err() error
	Deadline() (time.Time, bool)
	Value(interface{}) interface{}
}
```

And some more

[embedmd]:# (main.go)
```go
package main

import "fmt"

func main() {
	fmt.Println("hello")
}
```
