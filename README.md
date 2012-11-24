Goober
---------------
A  micro web framework for go.

### Why? / About
-----------

Go has an irresistibly fully featured standard library that includes, but is not limited to, http serving. With this kind of functionality baked in to the standard libraries, it is possible to create powerful frameworks in very little code.

Goober is, amongst other things, one of these frameworks. Its only distinguishing factor is the way it handles route matching — through a Trie-like tree structure. This way, the route parser doesn't waste time scanning through a long list of routes that won't match.

### Example
-----------

An example can be found in examples/main.go, but here is an excerpt of all that is needed to run a simple hello world:

	package main

	import (
	  "github.com/bigs/goober"
	  "net/http"
	  "io"
	  "log"
	)
	
	func hello(w http.ResponseWriter, r *goober.Request) {
	  io.WriteString(w, "Hello, <i>" + r.URLParams[":name"] + "</i>.")
	}

	func main() {
	  var g = goober.New()
	  g.Get("/hello/:name", hello)
	  g.ErrorPages[404] = "<h1>Not found.</h1>"

	  err := g.http.ListenAndServe(":3000")
	  if err != nil {
	    log.Fatal("ListenAndServe: ", err)
	  }
	}

And that's it!

### Implementation Details
------------

There are a few implementation details worthy of note. By default, your `Content-Type` header will be set to `text/html`.
