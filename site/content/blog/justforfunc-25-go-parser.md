+++
title = "Understanding Go programs with go/parser"
date = "2017-12-18T11:00:02-08:00"
+++

This blog post describes the same techniques used during episode 25 of
justforfunc which you can watch right below.

<iframe width="560" height="315" src="https://www.youtube.com/embed/YRWCa84pykM" frameborder="0" allowfullscreen></iframe>

### Previously in justforfunc

In a [previous post](TODO), we used the `go/scanner` package in Go's
standard library to identify which was the most common identifier in
the standard library itself.

_Spoiler alert_: it was `v`.

In order to get a somehow more meaningful list, we limited the search
to those identifiers that were three letters long or more.
This gave us `err` and `nil`, which we've all seen before
in the (in)famous `if err != nil {`  expression.

### Variables: package vs local

What if I wanted to know which is the most common name for local variables?
What about  functions or types? In order to answer this question
scanners fall short because they lack context. We know what tokens we saw
before, but in order to know whether a `var a = 3` is declared at package,
function, or even block levels we need more context.

A package has many declarations, some of those declarations may be of
functions which in time could declare local variables, constants, or
even more functions!

But how do we find that structure from a sequence of tokens?
Well, every single programming language has
a set of rules that inform us on how to build such a tree from a sequence
of tokens. This looks something like:

```
VarDecl     = "var" ( VarSpec | "(" { VarSpec ";" } ")" ) .
VarSpec     = IdentifierList ( Type [ "=" ExpressionList ] | "=" ExpressionList ) .
```

The rules above tell us that a `VarDecl` (variable declaration) starts with a
`var` token, followed by either a `VarSpec` (variable specification) or a list
of them surrounded by parentheses and separated by semicolons. 

_Note_: Those semicolons are actually added by the Go scanner,
so you might not see them but the parser does.

If we start with a piece of Go code containing `var a = 3 `, using `go/scanner`
we could obtain the following list of tokens.

```
[VAR], [IDENT "a"], [ASSIGN], [INT "3"], [SEMICOLON]
```

The rules above help us figure out this is a `VarDecl` with only one `VarSpec`.
Then we'll parse an `IdentifierList` with a single `Identifier` `a`, no `Type`,
and an `ExpressionList` with an `Expression` which is an integer with value `3`.

Or, as a tree, it would look like the image below.

![AST from the previous tokens](/img/parser/ast.png)

The rules that allow us to go from a sequence of tokens to a tree
structure form a *language grammar*, or syntax.
The resulting tree is called an *Abstract Syntax Tree*, often simply *AST*.

## Using go/scanner

Enough theory for now, let's write some code! Let's see how we can parse
the expression `var a = 3` and obtain an AST from it.

{{<highlight go>}}
package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "", "var a = 3", parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(f)
}
{{</highlight>}}

This program compiles (yay!) but if you run you'll see the following error:

```bash
1:1: expected 'package', found 'var' (and 1 more errors)
```

Oh, yeah. In order to parse a declaration we are calling `ParseFile`, which
expects a full Go file therefore starting with `package` before any other
code (expect for comments).

If you were parsing an expression, such as `3 + 5` or other pieces of code
which you could see as a value you could pass as a parameter, then you have
`ParseExpr`, but that function will not help with a function declaration.

Let's simply add `package main` at the beginning of our code and see the AST
we obtain.

{{<highlight go>}}
package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "", "package main; var a = 3", parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(f)
}
{{</highlight>}}


And when we run it we get something a bit better ... just a bit, though.

```bash
$ go run main.go
&{<nil> 1 main [0xc420054100] scope 0xc42000e210 {
        var a
}
 [] [] []}
 ```

Let's print more detail by replacing the `Println` call by `fmt.Printf("%#v", f)`
and try again.

```bash
go run main.go
&ast.File{Doc:(*ast.CommentGroup)(nil), Package:1, Name:(*ast.Ident)(0xc42000a060), Decls:[]ast.Decl{(*ast.GenDecl)(0xc420054100)}, Scope:(*ast.Scope)(0xc42000e210), Imports:[]*ast.ImportSpec(nil), Unresolved:[]*ast.Ident(nil), Comments:[]*ast.CommentGroup(nil)}
```

Ok, that's better but I'm too lazy to read this. Let's import 
`github.com/davecgh/go-spew/spew` and use it to print an easier to read value.

{{<highlight go>}}
package main

import (
	"go/parser"
	"go/token"
	"log"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "", "package main; var a = 3", parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}
	spew.Dump(f)
}
{{</highlight>}}

Running this program shows us pretty much the same as before, but in a much more
readable format.

```bash
$ go run main.go
(*ast.File)(0xc42009c000)({
 Doc: (*ast.CommentGroup)(<nil>),
 Package: (token.Pos) 1,
 Name: (*ast.Ident)(0xc42000a120)(main),
 Decls: ([]ast.Decl) (len=1 cap=1) {
  (*ast.GenDecl)(0xc420054100)({
   Doc: (*ast.CommentGroup)(<nil>),
   TokPos: (token.Pos) 15,
   Tok: (token.Token) var,
   Lparen: (token.Pos) 0,
   Specs: ([]ast.Spec) (len=1 cap=1) {
    (*ast.ValueSpec)(0xc4200802d0)({
     Doc: (*ast.CommentGroup)(<nil>),
     Names: ([]*ast.Ident) (len=1 cap=1) {
      (*ast.Ident)(0xc42000a140)(a)
     },
     Type: (ast.Expr) <nil>,
     Values: ([]ast.Expr) (len=1 cap=1) {
      (*ast.BasicLit)(0xc42000a160)({
       ValuePos: (token.Pos) 23,
       Kind: (token.Token) INT,
       Value: (string) (len=1) "3"
      })
     },
     Comment: (*ast.CommentGroup)(<nil>)
    })
   },
   Rparen: (token.Pos) 0
  })
 },
 Scope: (*ast.Scope)(0xc42000e2b0)(scope 0xc42000e2b0 {
        var a
}
),
 Imports: ([]*ast.ImportSpec) <nil>,
 Unresolved: ([]*ast.Ident) <nil>,
 Comments: ([]*ast.CommentGroup) <nil>
})
```

I recommend spending some time actually reading this tree and seeing
how each piece matches with the original code. Feel free to ignore
`Scope`, `Obj`, and `Unresolved` as we'll talk about those later on.

## Going from AST to code

It is sometime useful to print an AST into the corresponding source code
rather than its tree form. To do so, we have the `go/printer` package
which you can easily use as a way to see what information is stored in
the AST.

{{<highlight go>}}
package main

import (
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
)

func main() {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "", "package main; var a = 3", parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}
	printer.Fprint(os.Stdout, fs, f)
}
{{</highlight>}}


Executing this program prints the source code we parsed originally. It is now a
good point to see how different values of the parsing mode affect what information
is kept in the AST. Replace `parser.AllErrors` with `parser.ImportsOnly` or other
possible values.

## Navigating the AST

So far we have a tree that contains all of the information we want, but how do we
extract the pieces of information we care about? This is where the `go/ast` package
comes in handy (other than for declaring the `ast.File` that `parser.ParseFile`
returns!).

Let's use `ast.Walk`. This function receives two arguments. The second one is an
`ast.Node` which is an interface satified by all the nodes in the AST.
The first argument is an `ast.Visitor` which is also obviously an interface.

This interface has a single method.

{{<highlight go>}}
type Visitor interface {
	Visit(node Node) (w Visitor)
}
{{</highlight>}}


Ok, so we already have a node, since the `ast.File` returned by `parser.ParseFile`
satisfies the interface, but we still need to create an `ast.Visitor`.

Let's simply write one that prints the type of the node and returns itself.

{{<highlight go>}}
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "", "package main; var a = 3", parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	var v visitor
	ast.Walk(v, f)
}

type visitor struct{}

func (v visitor) Visit(n ast.Node) ast.Visitor {
	fmt.Printf("%T\n", n)
	return v
}
{{</highlight>}}

Running this program gives us a sequence of nodes, but we've lost the tree structure.
Also, what are all of those `nil` nodes? Well, we should read the documentation of
`ast.Walk`! Turns out the `Visitor` we return is used to visit each one of the children
of the current node, and we end up by calling the `Visitor` with a `nil` node as a
way to communicate there's no more nodes to visited.

Using that knowledge we can now print something that looks more like a tree.

{{<highlight go>}}
type visitor int

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	fmt.Printf("%s%T\n", strings.Repeat("\t", int(v)), n)
	return v + 1
}
{{</highlight>}}

The rest of the code in our program remains unchanged, and executing it prints this
output:

```go
*ast.File
        *ast.Ident
        *ast.GenDecl
                *ast.ValueSpec
                        *ast.Ident
                        *ast.BasicLit
```

### What's the most common name per kind of identifier?

Ok, so now that we understand how to parse code and visit its nodes,
we are ready to extract the information we want: what are the most
most common names for variables declared at package level vs those
declared locally inside of a function.

We wills tart with a piece of code very similar to what we had in the
previous episode, where we used `go/scanner` over a list of files
passed as command line arguments.


{{<highlight go>}}
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage:\n\t%s [files]\n", os.Args[0])
		os.Exit(1)
	}

	fs := token.NewFileSet()
	var v visitor

	for _, arg := range os.Args[1:] {
		f, err := parser.ParseFile(fs, arg, nil, parser.AllErrors)
		if err != nil {
			log.Printf("could not parse %s: %v", arg, err)
			continue
		}
		ast.Walk(v, f)
	}
}

type visitor int

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	fmt.Printf("%s%T\n", strings.Repeat("\t", int(v)), n)
	return v + 1
}
{{</highlight>}}

Running this program will print the AST of each one of the files given
as command line argument. You can try it on its own by running:

```bash
$ go build -o parser main.go  && parser main.go
# output removed for brevity
```

Let's now change our `visitor` to keep track of how many times each
identifier is used for each kind of variable declaration.

First let's start by tracking all short variable declarations, since
we know they are always local declarations.

{{<highlight go>}}
type visitor struct {
	locals map[string]int
}

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	switch d := n.(type) {
	case *ast.AssignStmt:
		for _, name := range d.Lhs {
			if ident, ok := name.(*ast.Ident); ok {
				if ident.Name == "_" {
					continue
				}
				if ident.Obj != nil && ident.Obj.Pos() == ident.Pos() {
					v.locals[ident.Name]++
				}
			}
		}
	}
	return v
}
{{</highlight>}}

For each assignment statement we're checking whether the identifier
name is `_`, which should be ignored, and whether this is the declaration
point of the identifier. In order to do so we use the `Obj` field which
keeps track of all the objects declared in a context.

If the `Obj` field is `nil` we know that the variable was declared in a
different file, therefore it's not a local variable declaration and we can
ignore it.

If we run this program on the whole standard library we'll see that the most
common identifiers are:

```
  7761 err
  6310 x
  5446 got
  4702 i
  3821 c
```

Interestingly enough, `v` doesn't appear at all! Are we missing any other ways
of declaring local variables?


### Counting parameters and range variables too

We're missing a couple node types to have a more completely analysis of local
variables. These are:

- function parameters, receivers, and named returned values.
- range statements.

Since the code to count a local variable identifier will be repeated all over
let's define a method on `visitor` instead.

{{<highlight go>}}
func (v visitor) local(n ast.Node) {
	ident, ok := n.(*ast.Ident)
	if !ok {
		return
	}
	if ident.Name == "_" || ident.Name == "" {
		return
	}
	if ident.Obj != nil && ident.Obj.Pos() == ident.Pos() {
		v.locals[ident.Name]++
	}
}
{{</highlight>}}

For parameters and returned values we will have a list of identifiers.
We also have the same for method receivers, even though they always have
only one element. Let's define an extra method for lists of identifiers.

{{<highlight go>}}
func (v visitor) localList(fs []*ast.Field) {
	for _, f := range fs {
		for _, name := range f.Names {
			v.local(name)
		}
	}
}
{{</highlight>}}

Then we can use that method for all the node types that might declare local
variables.

{{<highlight go>}}
	case *ast.AssignStmt:
		if d.Tok != token.DEFINE {
			return v
		}
		for _, name := range d.Lhs {
			v.local(name)
		}
	case *ast.RangeStmt:
		v.local(d.Key)
		v.local(d.Value)
	case *ast.FuncDecl:
		v.localList(d.Recv.List)
		v.localList(d.Type.Params.List)
		if d.Type.Results != nil {
			v.localList(d.Type.Results.List)
		}
{{</highlight>}}

Great, let's run this and see which one is the most common local variable name for now!

```bash
$ ./parser ~/go/src/**/*.go
most common local variable names
 12264 err
  9395 t
  9163 x
  7442 i
  6127 c
```

### Handling var declarations

Let's move into handling the `var` declarations. These are more interesting
because they could be local or global, and the only way to know is to check
whether they're at the `ast.File` level.

To do so we're going to create a new `visitor` per file which will keep track
of the declarations that are global, so we can count the identifiers correctly.

To do so we'll add a `pkgDecls` field of type `map[*ast.GenDecl]bool` in our
visitor, and it will be initialized by a `newVisitor` function.
We'll also add a `globals` field tracking how many times an identifier has been
declared.

{{<highlight go>}}
type visitor struct {
	pkgDecls map[*ast.GenDecl]bool
	globals  map[string]int
	locals   map[string]int
}

func newVisitor(f *ast.File) visitor {
	decls := make(map[*ast.GenDecl]bool)
	for _, decl := range f.Decls {
		if d, ok := decl.(*ast.GenDecl); ok {
			decls[d] = true
		}
	}

	return visitor{
		decls,
		make(map[string]int),
		make(map[string]int),
	}
}
{{</highlight>}}

Our main program will need to create a new `visitor` per file
and keep track of the total results.

{{<highlight go>}}
	locals, globals := make(map[string]int), make(map[string]int)

	for _, arg := range os.Args[1:] {
		f, err := parser.ParseFile(fs, arg, nil, parser.AllErrors)
		if err != nil {
			log.Printf("could not parse %s: %v", arg, err)
			continue
		}

		v := newVisitor(f)
		ast.Walk(v, f)
		for k, v := range v.locals {
			locals[k] += v
		}
		for k, v := range v.globals {
			globals[k] += v
		}
	}
{{</highlight>}}

Ok, we just have one more piece of the puzzle to complete. We need to track the
`*ast.GenDecl` nodes and find all the variable declarations in them.

{{<highlight go>}}
	case *ast.GenDecl:
		if d.Tok != token.VAR {
			return v
		}
		for _, spec := range d.Specs {
			if value, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range value.Names {
					if name.Name == "_" {
						continue
					}
					if v.pkgDecls[d] {
						v.globals[name.Name]++
					} else {
						v.locals[name.Name]++
					}
				}
			}
		}
{{</highlight>}}

For each declaration we will only count those that start with a `token.VAR`,
therefore ignoring constants, types, and other kind of identifiers.
THen, for each value declared, we'll check whether it's a global or local
declaration, and count them accordingly and ignoring `_`.

The whole program is available [here](TODO).

Executing the program will give us this result:

```
$ ./parser ~/go/src/**/*.go
most common local variable names
 12565 err
  9876 x
  9464 t
  7554 i
  6226 b
most common global variable names
    29 errors
    28 signals
    23 failed
    15 tests
    12 debug
```

So there you go, the most common local variable name is `err`, while
the most common package variable name is `errors`.

Which one is the most common constant name? How would you find it?

## Thanks

If you enjoyed this episode make sure you share it and subscribe to
[justforfunc](http://justforfunc.com)!
Also, consider sponsoring the series on [patreon](https://patreon.com/justforfunc).