+++
title = ""
date = "2017-12-11T11:39:02-08:00"
+++

In a previous post we used the `go/scanner` package in Go's standard
library to analyze what was the most common identifier in the standard
library itself. Spoiler alert: it was `v`.

In order to get a (somehow) more meaningful list, we limited the search
to those identifiers that were three letters long or more. This gave us
the expected `err` and `nil` identifiers, which we've all used before
in the famous `if err != nil {`  expression.

But, what if my question was slightly different? What if I wanted to know
what's the most common name for local variables vs. the most common name
for say package functions or types? In order to answer this question
scanners fall short because they lack context. We know what tokens we saw
before, but in order to know whether a `var a = 3` is declared at package,
function, or even block levels we need context, and that context has a
tree shape.

A package has many declarations, some of those declarations may be of
functions which in time could declare local variables, constants, or
even more functions!

![a tree made of source code]()

How do we obtain this tree? Well, every single programming language has
a set of rules that inform us on how to build such a tree from a sequence
of tokens. This looks something like:

```
VarDecl     = "var" ( VarSpec | "(" { VarSpec ";" } ")" ) .
VarSpec     = IdentifierList ( Type [ "=" ExpressionList ] | "=" ExpressionList ) .
```

The rule above tells us that a `VarDecl` (variable declaration) starts with a
`var` token, followed by either a `VarSpec` (variable specification) or a list
of them surrounded by parentheses and separated by semicolons. Those semicolons
are actually added by the Go scanner, so you might not see them but the parser
does.

So given the list of tokens corresponding to `var a = 3 `, which would look like:

```
[VAR], [IDENT "a"], [ASSIGN], [INT "3"], [SEMICOLON]
```

The rules above helps us determine that this is `VarDecl` with only one `VarSpec`,
whose `IdentifierList` contains a single `Identifier` with name `a`, no `Type`,
and an `ExpressionList` with a single `Expression` which is an integer with value `3`.

Those rules allowing us to go from a sequence of tokens to a tree structure form
the language grammar, or syntax, and the resulting tree is called an Abstract Syntax
Tree, often called AST.

## Using go/scanner

Enough theory for now, let's write some code! Let's see how we can parse the expression
`var a = 3` and obtain an AST from it.

{{< highlight go >}}
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
{{</ highlight >}}

This program compiles (yay!) but if you run you'll see the following error:

```bash
2017/12/10 14:06:14 1:1: expected 'package', found 'var' (and 1 more errors)
exit status 1
```

Oh, yeah. In order to parse a declaration we are calling `ParseFile`, which
expects a full Go file therefore starting with `package` before any other
code (expect for comments).

If you were parsing an expression, such as `3 + 5` or other pieces of code
which you could see as a value you could pass as a parameter, then you have
`ParseExpr`, but that function will not help with a function declaration.

Let's simply add `package main` at the beginning of our code and see the AST
we obtain.

{{< highlight go >}}
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
{{</ highlight >}}

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

{{< highlight go >}}
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
{{</ highlight >}}

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

{{ < highlight go >}}
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
{{ </ highlight >}}

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

{{ < highlight go >}}
type Visitor interface {
	Visit(node Node) (w Visitor)
}
{{ </ highlight >}}

Ok, so we already have a node, since the `ast.File` returned by `parser.ParseFile`
satisfies the interface, but we still need to create an `ast.Visitor`.

Let's simply write one that prints the type of the node and returns itself.

{{ < highlight go >}}
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
{{ </ highlight >}}

Running this program gives us a sequence of nodes, but we've lost the tree structure.
Also, what are all of those `nil` nodes? Well, we should read the documentation of
`ast.Walk`! Turns out the `Visitor` we return is used to visit each one of the children
of the current node, and we end up by calling the `Visitor` with a `nil` node as a
way to communicate there's no more nodes to visited.

Using that knowledge we can now print something that looks more like a tree.

{{ < highlight go >}}
type visitor int

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	fmt.Printf("%s%T\n", strings.Repeat("\t", int(v)), n)
	return v + 1
}
{{ </ highlight >}}

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

## What's the most common name per kind of identifier?

Ok, so now that we understand how to parse code and visit its nodes,
we are ready to extract the information we want: what are the most
most common names for variables declared at package level vs those
declared locally inside of a function.

We wills tart with a piece of code very similar to what we had in the
previous episode, where we used `go/scanner` over a list of files
passed as command line arguments.


{{ < highlight go >}}
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
{{ </ highlight >}}


Running this program will print the AST of each one of the files given
as command line argument. You can try it on its own by running:

{{ < highlight bash >}}
$ go run main.go -- main.go
# output removed for brevity
{{ </ highlight >}}

Let's now change our `visitor` to keep track of how many times each
identifier is used for each kind of variable declaration.