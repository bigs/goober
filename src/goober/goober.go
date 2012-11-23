package goober

import  (
  "net/http"
  "strings"
  "io"
  "time"
  "fmt"
)

type Goober struct {
  head map[string]*routeTreeNode
  ErrorPages map[int]string
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

  g := &Goober{
    head: head,
    ErrorPages: make(map[int]string),
  }

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
      if (cur.variables[part] != nil) {
        cur = cur.variables[part]
      } else {
        cur.variables[part] = newRouteTreeNode()
        cur = cur.variables[part]
      }
    } else {
      // static
      if (cur.children[part] != nil) {
        cur = cur.children[part]
      } else {
        cur.children[part] = newRouteTreeNode()
        cur = cur.children[part]
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
    // if we've reached a terminal state, return handler
    handler = node.handler
  } else {
    // else, look for it
    var part = parts[0]

    if node.children[part] != nil {
      // check static routes first
      return walkTree(node.children[part], parts[1:], r)
    } else {
      for k, v := range node.variables {
        // check all dynamic routes, taking first match
        handler, err = walkTree(v, parts[1:], r)
        if err == 0 {
          // goofy recursive way to build up params
          r.URLParams[k] = part
          return
        }
      }

      // if we don't find any dynamic matches, there was an error
      err = -1
    }
  }

  return
}

func (g *Goober) GetHandler(r *Request) (handler Handler, err int) {
  var path = strings.TrimFunc(r.URL.Path, isSlash)
  var parts = strings.Split(path, "/")
  return walkTree(g.head[r.Method], parts, r)
}

func (g *Goober) errorHandler(w http.ResponseWriter, r *Request, code int) {
  if page, ok := g.ErrorPages[code]; ok {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(code)
    io.WriteString(w, page)
  }
}

func (g *Goober) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  var startTime = time.Now()
  defer func() {
    fmt.Printf("[%s] %s - took %s\n", r.Method, r.URL.Path, time.Since(startTime))
  }()
  // create request object
  var request = &Request{
    Request: *r,
    URLParams: make(map[string]string),
  }

  // get the handler for the request
  var f, err = g.GetHandler(request)
  if err == 0 && f != nil {
    // user response
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    f(w, request)
  } else {
    g.errorHandler(w, request, 404)
  }

}

