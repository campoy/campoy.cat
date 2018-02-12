+++
title = "Analyzing the performance of Go functions with benchmarks"
date = "2018-02-12T00:12:00+01:00"
+++

# Analyzing the performance of Go functions with benchmarks

This blog post is complementary to [episode 27 of justforfunc](https://www.youtube.com/watch?v=t9bEg2A4jsw)
which you can watch right below.

<iframe width="560" height="315" src="https://www.youtube.com/embed/TODO" frameborder="0" allowfullscreen></iframe>

In the [previous blog post](TODO) I discussed two different ways of merging
n channels in Go, but we did not discuss which ones was faster.
In the meanwhile a third way of merging channels was proposed in a YouTube comment.

This blog post will show that third way and compare all of them from a performance point
of view, analyzed using benchmarks.

# A third way of merging channels

Two episodes ago we discussed how two channels could be merged, using a single goroutine
and nil channels. A [justforfunc](https://justforfunc.com) viewer proposed a way to use
this function and recursion to provide a way to merge n channels.

The solution is quite smart. If we have:

- one channel, we return that channel.
- two channels or more, we merge a half of the channels and then merge them using the result of those using the efficient function.

What about no channels? We will return an already closed channel.

Written in code it would look like this:

{{<highlight go>}}
func mergeRec(chans ...<-chan int) <-chan int {
	switch len(chans) {
	case 0:
		c := make(chan int)
		close(c)
		return c
	case 1:
		return chans[0]
	default:
		m := len(chans) / 2
		return mergeTwo(
			mergeRec(chans[:m]...),
			mergeRec(chans[m:]...))
	}
}
{{</highlight>}}

This is a nice solution that avoids using reflection and also reduces the number of goroutines needed.
On the other hand, it uses more channels than before!

So, which one is fastest? Time for benchmarks!

## Writing a benchmark

Testing and benchmarks are integrated very tightly with the Go tooling, and writing a benchmark is
as simple as writing a function with a name starting with `Benchmark` and the appropriate signature.

The only parameter a benchmark receives is of type `testing.B` and it defines the number of times
we should perform the operation we're benchmarking. This is done for statistical significance, by
default benchmarks are given a number of iterations that will make the function run for a whole second.

{{<highlight go>}}
func BenchmarkFoo(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // perform the operation we're analyzing
    }
}
{{</highlight>}}

Ok, so now we know pretty much everything we need to know to write a benchmark for one of our
functions.

### Benchmarking all our functions

Ok, so we're ready to write benchmarks for each function!

One for `merge`:

{{<highlight go>}}
func BenchmarkMerge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := merge(asChan(0, 1, 2, 3, 4, 5, 6, 7, 8, 9))
		for range c {
		}
	}
}
{{</highlight>}}

One for `mergeReflect`:

{{<highlight go>}}
func BenchmarkMergeReflect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := mergeReflect(asChan(0, 1, 2, 3, 4, 5, 6, 7, 8, 9))
		for range c {
		}
	}
}
{{</highlight>}}

And finally. one for `mergeRec`:

{{<highlight go>}}
func BenchmarkMergeRec(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := mergeRec(asChan(0, 1, 2, 3, 4, 5, 6, 7, 8, 9))
		for range c {
		}
	}
}
{{</highlight>}}

And now we can finally run them!

{{<highlight bash>}}
➜ go test -bench=.
goos: darwin
goarch: amd64
pkg: github.com/campoy/justforfunc/27-merging-chans
BenchmarkMerge-4                  200000              7074 ns/op
BenchmarkMergeReflect-4           100000             11904 ns/op
BenchmarkMergeRec-4               500000              2475 ns/op
PASS
ok      github.com/campoy/justforfunc/27-merging-chans  4.077s
{{</highlight>}}

Great, so using recursion is the fastest possible option. Let's believe it for now!

### Using subbenchmarks and b.Run

Unfortunately, in order to benchmark three functions we're repeating lots of code.
Wouldn't it be nice if we could write our benchmark once?

We can write a function that iterates over our three functions:

{{<highlight go>}}
func BenchmarkMerge(b *testing.B) {
	merges := []func(...<-chan int) <-chan int{
		merge,
		mergeReflect,
		mergeRec,
	}

	for _, merge := range merges {
		for i := 0; i < b.N; i++ {
			c := merge(asChan(0, 1, 2, 3, 4, 5, 6, 7, 8, 9))
			for range c {
			}
		}
	}
}
{{</highlight>}}

Unfortunately, that simply gives a single and pretty meaningless measure, mixing all performances.

{{<highlight bash>}}
➜ go test -bench=.
goos: darwin
goarch: amd64
pkg: github.com/campoy/justforfunc/27-merging-chans
BenchmarkMerge-4           50000             22440 ns/op
PASS
ok      github.com/campoy/justforfunc/27-merging-chans  1.386s
{{</highlight>}}

Luckily, we can easily create subbenchmarks by calling `testing.B.Run`, similarly as subtests
with `testing.T.Run`. In order to do so, we need a name, which is actually a pretty good idea
from a documentation point of view, anyway.

{{<highlight go>}}
func BenchmarkMerge(b *testing.B) {
	merges := []struct {
		name string
		fun  func(...<-chan int) <-chan int
	}{
		{"goroutines", merge},
		{"reflect", mergeReflect},
		{"recursion", mergeRec},
	}

	for _, merge := range merges {
		b.Run(merge.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				c := merge.fun(asChan(0, 1, 2, 3, 4, 5, 6, 7, 8, 9))
				for range c {
				}
			}
		})
	}
}
{{</highlight>}}

When we run these benchmarks we'll now see three different results as we wished.

{{<highlight bash>}}
➜ go test -bench=.
goos: darwin
goarch: amd64
pkg: github.com/campoy/justforfunc/27-merging-chans
BenchmarkMerge/goroutines-4               200000              6145 ns/op
BenchmarkMerge/reflect-4                  200000             10962 ns/op
BenchmarkMerge/recursion-4                500000              2368 ns/op
PASS
ok      github.com/campoy/justforfunc/27-merging-chans  4.829s
{{</highlight>}}

Sweet! With way less code we're still able to benchmark all of our functions.

#### Benchmarking multiple values of N

So far we've seen that for a single channel the recursive algorithm is the fastest,
but what about other values of `N`?

We could probably just change the value of `N` many times and see how it goes.
But we could also just use a for loop!

In the code below we iterate over the powers of 2 from 1 to 1024, and use a combination
of the function we're benchmarking and the value of `N` as the benchmark name.

{{<highlight go>}}
func BenchmarkMerge(b *testing.B) {
	merges := []struct {
		name string
		fun  func(...<-chan int) <-chan int
	}{
		{"goroutines", merge},
		{"reflect", mergeReflect},
		{"recursion", mergeRec},
	}

	for _, merge := range merges {
		for k := 0.; k <= 10; k++ {
			n := int(math.Pow(2, k))
			b.Run(fmt.Sprintf("%s/%d", merge.name, n), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					chans := make([]<-chan int, n)
					for j := range chans {
						chans[j] = asChan(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
					}
					c := merge.fun(chans...)
					for range c {
					}
				}
			})
		}
	}
}
{{</highlight>}}

Let's run the benchmarks!

{{<highlight bash>}}
➜ go test -bench=.
goos: darwin
goarch: amd64
pkg: github.com/campoy/justforfunc/27-merging-chans
BenchmarkMerge/goroutines/1-4             200000              6919 ns/op
BenchmarkMerge/goroutines/2-4             100000             13212 ns/op
BenchmarkMerge/goroutines/4-4              50000             25469 ns/op
BenchmarkMerge/goroutines/8-4              30000             50819 ns/op
BenchmarkMerge/goroutines/16-4             20000             88566 ns/op
BenchmarkMerge/goroutines/32-4             10000            162391 ns/op
BenchmarkMerge/goroutines/64-4              5000            299955 ns/op
BenchmarkMerge/goroutines/128-4             3000            574043 ns/op
BenchmarkMerge/goroutines/256-4             1000           1129372 ns/op
BenchmarkMerge/goroutines/512-4             1000           2251411 ns/op
BenchmarkMerge/goroutines/1024-4             300           4760560 ns/op
BenchmarkMerge/reflect/1-4                200000             10868 ns/op
BenchmarkMerge/reflect/2-4                100000             22335 ns/op
BenchmarkMerge/reflect/4-4                 30000             54882 ns/op
BenchmarkMerge/reflect/8-4                 10000            148218 ns/op
BenchmarkMerge/reflect/16-4                 3000            543921 ns/op
BenchmarkMerge/reflect/32-4                 1000           1694021 ns/op
BenchmarkMerge/reflect/64-4                  200           6102920 ns/op
BenchmarkMerge/reflect/128-4                 100          22648976 ns/op
BenchmarkMerge/reflect/256-4                  20          90204929 ns/op
BenchmarkMerge/reflect/512-4                   3         383579039 ns/op
BenchmarkMerge/reflect/1024-4                  1        1676544681 ns/op
BenchmarkMerge/recursion/1-4              500000              2658 ns/op
BenchmarkMerge/recursion/2-4              100000             14707 ns/op
BenchmarkMerge/recursion/4-4               30000             44520 ns/op
BenchmarkMerge/recursion/8-4               10000            114676 ns/op
BenchmarkMerge/recursion/16-4               5000            261880 ns/op
BenchmarkMerge/recursion/32-4               3000            560284 ns/op
BenchmarkMerge/recursion/64-4               2000           1117642 ns/op
BenchmarkMerge/recursion/128-4              1000           2242910 ns/op
BenchmarkMerge/recursion/256-4               300           4784719 ns/op
BenchmarkMerge/recursion/512-4               100          10044186 ns/op
BenchmarkMerge/recursion/1024-4              100          20599475 ns/op
PASS
ok      github.com/campoy/justforfunc/27-merging-chans  61.533s
{{</highlight>}}

We can see that even though the recursive algorithm is the fastest for a single channel,
very quick the solution with many goroutines outperforms the others.

We can also this by plotting these numbers with some quick Python.
Stop acting shocking I somtimes write in Python too!

{{<highlight python>}}
import matplotlib.pyplot as plt

ns = [2**x for x in range(11)];
data = {
  'goroutines': [6919, 13212, 25469, 50819, 88566, 162391, 299955, 574043, 1129372, 2251411, 4760560],
  'reflection': [10868, 22335, 54882, 148218, 543921, 1694021, 6102920, 22648976, 90204929, 383579039, 1676544681],
  'recursion': [2658, 14707, 44520, 114676, 261880, 560284, 1117642, 2242910, 4784719, 10044186, 20599475],
}

for (label, values) in data.items():
    plt.plot(ns, values, label=label)
plt.legend()
{{</highlight>}}

![benchmark graphics - linear scale](img/benchmarks/linear.png)

It's hard to see the difference, so let's use a logarithmic scale for the Y axis (simply add `plt.yscale('log')`).

![benchmark graphics - logarithmic scale](img/benchmarks/log.png)

### Managing the benchmark timer

There's an obvious issue with the way we're measuring performance.
We're also counting the time it takes to create the `N` channels for our
tests, which adds cost that is not relevant to our algorithms.

We can avoid this by using `b.StopTimer()` and `b.StartTimer()`.

{{<highlight go>}}
    for i := 0; i < b.N; i++ {
        b.StopTimer()
        chans := make([]<-chan int, n)
        for j := range chans {
            chans[j] = asChan(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
        }
        b.StartTimer()
        c := merge.fun(chans...)
        for range c {
        }
    }
{{</highlight>}}

The results are not that different to our previous benchmark, but we can see
the timing difference increases with the value of `N`.

## Can we make them faster?

We finally have a good view of how each function performs for different sizes of `N`.
This is great, but doesn't really help us making our functions faster.

How could we know more? Well, benchmarks take us only so far, but the next tool we
we can use is profiling! Wanna know more? Next episode coming up in two weeks.

## Thanks

If you enjoyed this episode make sure you share it and subscribe to
[justforfunc](http://justforfunc.com)!
Also, consider sponsoring the series on [patreon](https://patreon.com/justforfunc).