package main

import (
  "goober"
  "net/http"
  "io"
  "fmt"
)

func foo(w http.ResponseWriter, r *goober.Request) {
  io.WriteString(w, "Hello, world.")
}

func main() {
  var x = goober.New()
  x.Get("/butts", goober.HandlerFunc(foo))
  x.Get("/butts/bugs", goober.HandlerFunc(foo))
  x.Get("/butt/:foo", goober.HandlerFunc(foo))
  x.GetHandler("GET", "/butts/bugs")
  x.GetHandler("GET", "/butt/bugs")
  fmt.Println("blah")
}

