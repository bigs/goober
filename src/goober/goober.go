package goober

import  (
  "net/http"
  "strings"
  "fmt"
)

type Goober struct {
  head map[string]*routeTreeNode
}

type Handler func(http.ResponseWriter, *Request)

type RouteMap map[string]*routeTreeNode

type routeTreeNode struct {
  handler Handler
  children RouteMap
  variables RouteMap
}

// Special request structure

type Request struct {
  http.Request
  URLParams map[string]string
}

func newRouteTreeNode() (node *routeTreeNode) {
  node = &routeTreeNode{
    children: make(RouteMap),
    variables: make(RouteMap),
  }

  return
}

// TODO: create my own interface.
//
// the program should register a default handler but all user-added
// handlers should be of our own interface using goober.Request, allowing
// for a sexier route params hash

func New() (* Goober) {
  var head = make(RouteMap)
  head = head
  head["GET"] = newRouteTreeNode()
  head["POST"] = newRouteTreeNode()
  head["PUT"] = newRouteTreeNode()
  head["DELETE"] = newRouteTreeNode()

  g := &Goober{head: head}

  return g
}

func isSlash(s rune) (bool) {
  return s == '/'
}

func (g *Goober) AddHandler(method string, route string, handler Handler) (err int){
  err = 0
  route = strings.TrimFunc(route, isSlash)
  var parts = strings.Split(route, "/")

  var cur = g.head[method]
  for i := range parts {
    var part = parts[i]

    if (len(part) == 0) {
      err = 1
      return
    }

    fmt.Println("for part: \"" + part + "\"")

    if strings.HasPrefix(part, ":") {
      // dynamic
      if (cur.variables[part] != nil) {
        fmt.Println("added to variables")
        cur = cur.variables[part]
      } else {
        fmt.Println("new variables")
        var next = newRouteTreeNode()
        cur.variables[part] = next
        cur = next
      }
    } else {
      if (cur.children[part] != nil) {
        fmt.Println("added to children")
        cur = cur.children[part]
      } else {
        fmt.Println("new children")
        var next = newRouteTreeNode()
        cur.children[part] = next
        cur = next
      }
    }
  }
  // add handler
  cur.handler = handler
  return
}

func (g *Goober) Get(route string, handler Handler) (int) {
  return g.AddHandler("GET", route, handler)
}

func (g *Goober) Post(route string, handler Handler) (int) {
  return g.AddHandler("POST", route, handler)
}

func (g *Goober) Put(route string, handler Handler) (int) {
  return g.AddHandler("PUT", route, handler)
}

func (g *Goober) Delete(route string, handler Handler) (int) {
  return g.AddHandler("DELETE", route, handler)
}

func walkTree(node *routeTreeNode, parts []string, r *Request) (handler Handler, err int) {
  err = 0
  handler = nil
  if len(parts) == 0 {
    fmt.Println("arrived at handler")
    handler = node.handler
  } else {
    var part = parts[0]

    fmt.Println("testing: " + part)
    if node.children[part] != nil {
      fmt.Println("Going static.")
      return walkTree(node.children[part], parts[1:], r)
    } else {
      fmt.Println(node.variables)
      for k, v := range node.variables {
        handler, err = walkTree(v, parts[1:], r)
        if err == 0 {
          r.URLParams[k] = part
          fmt.Println("YAY: " + k)
          return
        }
      }
      err = -1
    }
  }

  return
}

func (g *Goober) GetHandler(r *Request) (handler Handler, err int) {
  var path = strings.TrimFunc(r.URL.Path, isSlash)
  var parts = strings.Split(path, "/")
  handler, err = walkTree(g.head[r.Method], parts, r)
  fmt.Printf("Error: %d\n", err)
  return
}

func (g *Goober) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  var request = &Request{
    Request: *r,
    URLParams: make(map[string]string),
  }

  var f, err = g.GetHandler(request)
  if err == 0 && f != nil {
    f(w, request)
  }
}

