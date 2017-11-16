+++
date = "2017-11-15T00:00:00-00:00"
title = "source{d}: why I left Google"
+++

On November 1st I left Google and in my [goodbye note](/blog/googbye/)
I sneakily said I would "try my luck in a small startup with huge potential".

The time to explain more has come, so let me tell you what I'm up to lately.
I am now the VP of Developer Relations at [source{d}](https://sourced.tech).
Maybe you've never heard about it, maybe you've heard a bit, or used some of
their awesome open source libraries such as
[go-git](https://github.com/src-d/go-git) or
[kmcuda](https://github.com/src-d/kmcuda) or
[proteus](https://github.com/src-d/proteus).

Before I tell you about what source{d} does, let me give you a bit of
context.

### Extracting information from source code

For maintainers of any large codebase, such as open source projects or large
tech companies, it is essential to be able to understand their codebases.
Decisions are made based on this information.

A year ago I wrote [an article](https://medium.com/google-cloud/analyzing-go-code-with-bigquery-485c70c3b451)
on how one could use [Bigquery](https://cloud.google.com/bigquery) to analyze
all of the Go code available on GitHub.

Later on, this kind of analysis started to become a requirement to justify
additions to Go's standard library.

For instance, the `time.Until` function was added to Go with this
[proposal](https://github.com/golang/go/issues/14595) after an analysis of
how many times we could find an equivalent piece of code on GitHub.

![screenshot from the proposal](img/hello-sourced/issue.png)
[comment](https://github.com/golang/go/issues/14595#issuecomment-235651095) on the issue

This approach is powerful, but it definitely has limitations.

First of all, it limits
the analysis we can perform to regular expressions on source code. Most questions require
a deeper understanding of the structure of source code, such as the abstract syntax tree,
or even type information.

Additionally, when I said "all of the repositories on GitHub" this was not completely
accurate, it is a partial dump of all of these repositories, and even if it was there's
many repositories that are not there: what about the unix kernel?

### So what does source{d} do?

So what does source{d} do? They ... We! provide a powerful platform to access all of this
data in an easier and more powerful way.

Rather than limiting repositories on GitHub, source{d} is able to analyze any public
repository in the world, including those that are not even on the internet by running
our open source software on your own premise.

Secondly, we consider that the input to many good analysis should be the abstract syntax
tree of a program rather than the flat suite of bytes that is the source code.
We believe this so much that we've create [bblfsh](https://doc.bblf.sh/), a project that
one day will be able to parse any programming language and generate an abstract syntax
tree in a universal format. We call this format a universal abstract syntax tree or UAST.

Finally, while we love regular expressions, we believe that Machine Learning will
revolutionize how we analyze programs. There's inifite use cases that can use ML
over source code: better autocompletion (that doesn't require a connection to a 3rd party
server), better linters, automated code reviews, and one day source code generation
from natural language specifications.

To get there we have an incredibly talented Machine Learning team building models and
publishing papers and blog posts which will eventually become composable tools that
you'll be able to use in your own tools.

### So what's coming?

As VP of Developer Relations my job is to strategize how source{d} can empower developers
all around the world to write better code by:

- writing new models that will power amazing tools,
- building tools powered by ML on code, and of course
- using those tools to analyze and improve source code.

Want to learn more?

Join our [slack channel](https://sourced.tech/), follow us on
[twitter](https://twitter.com/srcd_), or drop me a line on francesc@sourced.tech.
I'm incredibly excited about the new opportunities that ML on code provide. Together we
can build better tools, for better source code, for eventually a better world.

Francesc