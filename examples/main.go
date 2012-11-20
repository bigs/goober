package main

import (
  "goober"
  "net/http"
  "io"
  "fmt"
)

func foo(w http.ResponseWriter, r *http.Request) {
  io.WriteString(w, "Hello, world.")
}

func main() {
  var x = goober.New()
  x.Get("/butts", http.HandlerFunc(foo))
  x.Get("/butts/bugs", http.HandlerFunc(foo))
  fmt.Println("blah")
}

