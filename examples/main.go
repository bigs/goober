package main

import (
  "../../goober"
  "net/http"
  "io"
  "log"
  "strings"
)

func helloWorld(w http.ResponseWriter, r *goober.Request) {
  io.WriteString(w, "Hello, world.")
}

func helloAnyone(w http.ResponseWriter, r *goober.Request) {
  io.WriteString(w, "Hello, " + r.URLParams[":foo"] + ".<br>\n")
  for k, v := range r.URL.Query() {
    io.WriteString(w, k + ": " + strings.Join(v, ", ") + "<br>\n")
  }
}

func static(w http.ResponseWriter, r *goober.Request) {
  var fileName = "./" + r.URLParams["*"]
  http.ServeFile(w, &r.Request, fileName)
}

func main() {
  var g = goober.New()
  g.Get("/hello", helloWorld)
  g.Get("/hello/:foo", helloAnyone)
  g.Get("/assets/*", static)
  g.ErrorPages[404] = "<h1>Not found.</h1>"

  err := g.ListenAndServe(":3000")

  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}

