package main

import (
  "goober"
  "net/http"
  "io"
  "fmt"
  "log"
)

func foo(w http.ResponseWriter, r *goober.Request) {
  io.WriteString(w, "Hello, world.")
}

func yay(w http.ResponseWriter, r *goober.Request) {
  io.WriteString(w, "Hello, " + r.URLParams[":foo"])
}

func main() {
  var x = goober.New()
  x.Get("/butts", foo)
  x.Get("/butts/bugs", foo)
  x.Get("/butt/:foo", yay)

  http.Handle("/", x)
  err := http.ListenAndServe(":3000", nil)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
  fmt.Println("blah")
}

