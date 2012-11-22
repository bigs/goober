package goober

import  (
  "net/http"
  "strings"
  "fmt"
)

type Goober struct {
  head map[string]*routeTreeNode
}

type routeTreeNode struct {
  handler Handler
  children map[string]*routeTreeNode
  dynamicChildren map[string]*routeTreeNode
}

// Special request structure

type Request struct {
  http.Request
  URLParams map[string]string
}

type Handler interface {
  ServeHTTP(http.ResponseWriter, *Request)
}

type HandlerFunc func(http.ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *Request) {
  f(w, r)
}

type RouteMap map[string]*routeTreeNode

func newRouteTreeNode() (node *routeTreeNode) {
  node = new(routeTreeNode)
  node.children = make(RouteMap)
  node.dynamicChildren = make(RouteMap)

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

    if strings.HasPrefix(part, ":") {
      // dynamic
      if (cur.dynamicChildren[part] != nil) {
        cur = cur.dynamicChildren[part]
      } else {
        var next = newRouteTreeNode()
        cur.dynamicChildren[part] = next
        cur = next
      }
    } else {
      if (cur.children[part] != nil) {
        cur = cur.children[part]
      } else {
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

func (g *Goober) Get(route string, handler Handler) {
  g.AddHandler("GET", route, handler)
}

func (g *Goober) Post(route string, handler Handler) {
  g.AddHandler("POST", route, handler)
}

func (g *Goober) Put(route string, handler Handler) {
  g.AddHandler("PUT", route, handler)
}

func (g *Goober) Delete(route string, handler Handler) {
  g.AddHandler("DELETE", route, handler)
}

func walkTree(node *routeTreeNode, parts []string) (handler Handler, err int) {
  err = 0
  handler = nil
  if len(parts) == 0 {
    handler = node.handler
  } else {
    var part = parts[0]

    fmt.Println(part)
    if node.children[part] != nil {
      return walkTree(node.children[part], parts[1:])
    } else {
      for _, v := range node.dynamicChildren {
        handler, err = walkTree(v, parts[1:])
        if err == 0 {
          return
        }
      }
    }
  }

  return
}

func (g *Goober) GetHandler(method string, path string) (handler Handler, err int) {
  path = strings.TrimFunc(path, isSlash)
  var parts = strings.Split(path, "/")
  handler, err = walkTree(g.head[method], parts)
  fmt.Printf("Error: %d\n", err)
  return
}

