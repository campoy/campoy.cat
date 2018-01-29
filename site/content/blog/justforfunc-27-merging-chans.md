+++
title = "Two ways of merging N channels"
date = "2018-01-29T00:12:00+01:00"
+++

This blog post is complementary to [episode 27 of justforfunc](https://www.youtube.com/watch?v=B64hIRjNvLc)
which you can watch right below.

<iframe width="560" height="315" src="https://www.youtube.com/embed/B64hIRjNvLc" frameborder="0" allowfullscreen></iframe>

Two weeks ago I explained how nil channels were useful for some important concurrency patterns,
specifically when [merging two channels](https://medium.com/justforfunc/why-are-there-nil-channels-in-go-9877cc0b2308).
As a result, because this is the internet, many replied telling me there's better ways to merge
two channels than that. And guess what, I agree!

The algorithm I showed is useful for only two channels. Today we're going to explore how
to handle `n` channels where `n` is not known at compilation time.
This means we can't simply use `select` over all the channels, so what are the options?

## First way: using N goroutines

The first approach is to create a new goroutine per channel. Each one of these goroutines
will range over a channel and send each one of those elements to the output channel.

![schema for goroutines](img/mergingchans/goroutines.png)

Or in code:

{{<highlight go>}}
func merge(cs ...<-chan int) <-chan int {
    out := make(chan int)
    
    for _, c := range cs {
        go func() {
            for v := range c {
                out <- v
            }
        }()
    }

    return out
}
{{</highlight>}}

This code is wrong in a couple ways. Can you see how? Let's first see one of the issues.

In order to run it I define a new `asChan` function, which is explained in the
[previous justforfunc episode](https://medium.com/justforfunc/why-are-there-nil-channels-in-go-9877cc0b2308),
and use it from the `main` function as follows.

{{<highlight go>}}
func main() {
	a := asChan(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	b := asChan(10, 11, 12, 13, 14, 15, 16, 17, 18, 19)
	c := asChan(20, 21, 22, 23, 24, 25, 26, 27, 28, 29)
	for v := range merge(a, b, c) {
		fmt.Println(v)
	}
}
{{</highlight>}}

Ok, so what's the problem when we run it when the data race detector enabled, aka `-race`?

{{<highlight bash>}}
$ go run -race main.go
==================
WARNING: DATA RACE
Read at 0x00c420090018 by goroutine 9:
  main.merge.func1()
      /Users/francesc/src/github.com/campoy/campoy.cat/site/content/blog/main.go:35 +0x3f

Previous write at 0x00c420090018 by main goroutine:
  main.merge()
      /Users/francesc/src/github.com/campoy/campoy.cat/site/content/blog/main.go:33 +0xf0
  main.main()
      /Users/francesc/src/github.com/campoy/campoy.cat/site/content/blog/main.go:13 +0x282

Goroutine 9 (running) created at:
  main.merge()
      /Users/francesc/src/github.com/campoy/campoy.cat/site/content/blog/main.go:34 +0x9b
  main.main()
      /Users/francesc/src/github.com/campoy/campoy.cat/site/content/blog/main.go:13 +0x282
==================
{{</highlight>}}

### Fixing a data race

Indeed, we're accessing the `c` variable defined in the `for` loop from many goroutines.
This is dangerous, but luckily there's a common idiom that solves the problem.
Simply pass the variable as a parameter, that way it will be duplicated and each goroutine
will end up with a different variable to itself.

{{<highlight go>}}
func merge(cs ...<-chan int) <-chan int {
    out := make(chan int)
    
    for _, c := range cs {
        go func(c <-chan int) {
            for v := range c {
                out <- v
            }
        }(c)
    }

    return out
}
{{</highlight>}}

Now that the data race has been removed, let's run the program again.

{{<highlight bash>}}
$ go run main.go
20
10
0
...
9
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan receive]:
main.main()
        /Users/francesc/src/github.com/campoy/campoy.cat/site/content/blog/main.go:13 +0x24a
exit status 2
{{</highlight>}}

Ok, this is better and it looks pretty good until the end when everything crashes.
What's going on? Well, the `main` goroutine is blocked waiting for anyone to close the
channel on which it's ranging. Unfortunately we forgot to close the channel at all.

### Closing the output channel

When should we close the channel? Well, once all goroutines have finished sending values into it.
When exactly is that? We don't really know, but we can simply "wait" for them to be done.

And how do you wait for a group of things? A `WaitGroup`, defined in the `sync` package.

A `WaitGroup` has 3 main methods:
- `Add` adds a given number of things to be done.
- `Done` indicates one of those things is done.
- `Wait` blocks until the number of things to be done goes down to zero.

Let's fix our program and close the output channel when it's appropriate.

{{<highlight go>}}
func merge(cs ...<-chan int) <-chan int {
	out := make(chan int)

	var wg sync.WaitGroup
	wg.Add(len(cs))

	for _, c := range cs {
		go func(c <-chan int) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
{{</highlight>}}

Note that `Wait` is called in a different goroutine.

## Second way: using reflect.Select

A second way of solving the same problem is using `reflect.Select`
which provides a `Select` operation on a slice of `SelectCase`s.

In order to call it we're going to create a slice of `SelectCase`
that contains an element per channel.

{{<highlight go>}}
    var cases []reflect.SelectCase
    for _, c := range chans {
        cases = append(cases, reflect.SelectCase{
            Dir:  reflect.SelectRecv,
            Chan: reflect.ValueOf(c),
        })
    }
{{</highlight>}}

Then we can call `reflect.Select` and send the value we receive
to the `out` channel. Unless channel we received from was closed,
in which case we should remove the case from the slice. This is
very similar to what we did setting the channel to `nil` in the
previous episode.

{{<highlight go>}}
    i, v, ok := reflect.Select(cases)
    if !ok {
        cases = append(cases[:i], cases[i+1:]...)
        continue
    }
    out <- v.Interface().(int)
{{</highlight>}}

This should be done for as long there's at least a non closed channel.
In our case this is equivalent to the slice not being empty.

So the full code is as follows.

{{<highlight go>}}
func merge(cs ...<-chan int) <-chan int {
	out := make(chan int)

	var wg sync.WaitGroup
	wg.Add(len(cs))

	for _, c := range cs {
		go func(c <-chan int) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
{{</highlight>}}

Running this code will also work.

### Which one is fastest?

The first solution uses a linear number of goroutines (one per channel)
so that could somehow get slow at some point.

But the second solution uses reflection and we all know that's pretty slow,
right?

So which one is fastest? I sincerely don't know! And that's the question
I'll solve in the following episode when we'll talk about benchmarks.

## Thanks

If you enjoyed this episode make sure you share it and subscribe to
[justforfunc](http://justforfunc.com)!
Also, consider sponsoring the series on [patreon](https://patreon.com/justforfunc).