package goober

import  (
  "net/http"
  "strings"
  "io"
  "time"
  "fmt"
)

// Main goober struct. Abides the handler interface.
type Goober struct {
  head map[string]*routeTreeNode
  ErrorPages map[int]string
}

// Goober handlers, for simplicity, are just functions with a given
// signature.
type Handler func(http.ResponseWriter, *Request)
type PreHandler func(http.ResponseWriter, *Request) (error)

// We use this a few places, so we can give it a type as well.
type RouteMap map[string]*routeTreeNode

// Our parse tree structure for routes
type routeTreeNode struct {
  handler Handler // Handler if a node is a terminal
  pre []PreHandler
  post []Handler
  children RouteMap // Static children
  variables RouteMap // Dynamic/variable children
}

// Chaining for pre/post handlers

func (n *routeTreeNode) AddPreFunc(f PreHandler) (*routeTreeNode) {
  n.pre = append(n.pre[:], f)
  return n
}

func (n *routeTreeNode) AddPostFunc(f Handler) (*routeTreeNode) {
  n.post = append(n.post[:], f)
  return n
}

// Augment http.Request with URLParams that will be grabbed
// from the request in the form of /:variables/
type Request struct {
  http.Request
  URLParams map[string]string
}

// A quick initializer for routeTreeNodes
func newRouteTreeNode() (node *routeTreeNode) {
  node = &routeTreeNode{
    children: make(RouteMap),
    variables: make(RouteMap),
  }

  return
}

// Initialize our Goober object
func New() (* Goober) {
  var head = make(RouteMap)
  head["GET"] = newRouteTreeNode()
  head["HEAD"] = newRouteTreeNode()
  head["POST"] = newRouteTreeNode()
  head["PUT"] = newRouteTreeNode()
  head["DELETE"] = newRouteTreeNode()

  g := &Goober{
    head: head,
    ErrorPages: make(map[int]string),
  }

  return g
}

// Simple helper to allow us to trim leading and trailing /'s
func isSlash(s rune) (bool) {
  return s == '/'
}

type BadRouteError struct {
  Route string
  Reason string
}

func (e BadRouteError) Error() string {
  return "\"" + e.Route + "\" is an invalid route because " + e.Reason + "."
}

// Adds a handler to our route tree
func (g *Goober) AddHandler(method string, route string, handler Handler) (cur *routeTreeNode){
  route = strings.TrimFunc(route, isSlash)
  var parts = strings.Split(route, "/")

  // Iterate through the bits of our path and add to the tree
  cur = g.head[method]
  for i := range parts {
    var part = parts[i]

    // No // empty paths
    if (len(part) == 0) {
      return nil
    }

    // Check for variables
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

// Wrapper functions for common types of request
func (g *Goober) Get(route string, handler Handler) (node *routeTreeNode) {
  return g.AddHandler("GET", route, handler)
}

func (g *Goober) Head(route string, handler Handler) (node *routeTreeNode) {
  return g.AddHandler("HEAD", route, handler)
}

func (g *Goober) Post(route string, handler Handler) (node *routeTreeNode) {
  return g.AddHandler("POST", route, handler)
}

func (g *Goober) Put(route string, handler Handler) (node *routeTreeNode) {
  return g.AddHandler("PUT", route, handler)
}

func (g *Goober) Delete(route string, handler Handler) (node *routeTreeNode) {
  return g.AddHandler("DELETE", route, handler)
}

type RouteNotFoundError struct {
  Route string
}

func (e RouteNotFoundError) Error() string {
  return "Route \"" + e.Route + "\" was not found."
}

func walkTree(node *routeTreeNode, parts []string, r *Request) (*routeTreeNode, error) {
  var err error = nil
  if len(parts) == 0 {
    // if we've reached a terminal state, return node
    if node.handler == nil {
      err = &RouteNotFoundError{Route: r.URL.Path}
    }
    return node, err
  } else {
    // else, look for it
    var part = parts[0]

    if child, ok := node.children["*"]; ok {
      node = child
      r.URLParams["*"] = strings.Join(parts, "/")
    } else if node.children[part] != nil {
      // check static routes first, they have priority
      return walkTree(node.children[part], parts[1:], r)
    } else {
      for k, v := range node.variables {
        // check all dynamic routes, taking first match
        node, err = walkTree(v, parts[1:], r)
        if err == nil {
          // goofy recursive way to build up params
          r.URLParams[k] = part
          return node, err
        }
      }

      // if we don't find any dynamic matches, there was an error
      err = &RouteNotFoundError{Route: r.URL.Path}
    }
  }

  return node, err
}

// Given a request, find the appropriate handler
func (g *Goober) GetHandler(r *Request) (node *routeTreeNode, err error) {
  path := strings.TrimFunc(r.URL.Path, isSlash)
  parts := strings.Split(path, "/")
  return walkTree(g.head[r.Method], parts, r)
}

// A simple function to handle error pages for us
func (g *Goober) errorHandler(w http.ResponseWriter, r *Request, code int) {
  if page, ok := g.ErrorPages[code]; ok {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(code)
    io.WriteString(w, page)
  }
}

// Borrowed from web.go
func webTime(t time.Time) string {
  ftime := t.Format(time.RFC1123)
  if strings.HasSuffix(ftime, "UTC") {
    ftime = ftime[0:len(ftime)-3] + "GMT"
  }
  return ftime
}

// Routes requests
func (g *Goober) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  var startTime = time.Now()
  defer func() {
    fmt.Printf("[%s] %s - took %s\n", r.Method, r.URL.Path, time.Since(startTime))
    r.Body.Close()
  }()

  // create augmented request object
  var request = &Request{
    Request: *r,
    URLParams: make(map[string]string),
  }

  // get the handler for the request
  node, err := g.GetHandler(request)
  if err == nil {
    // user response. pad with content-type.
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Header().Set("Server", "goober.go")
    w.Header().Set("Date", webTime(time.Now().UTC()))

    // Run prefunctions
    for _, f := range node.pre {
      if e := f(w, request); e != nil {
        // if there is an error, 404 and exit out
        fmt.Println("[ERROR] " + e.Error())
        g.errorHandler(w, request, 404)
        return
      }
    }

    // Run the handler
    node.handler(w, request)
  } else {
    fmt.Println("[ERROR] " + err.Error())
    g.errorHandler(w, request, 404)
  }

}

// shortcut to start serving a goober service
func (g *Goober) ListenAndServe(addr string) (err error)  {
  http.Handle("/", g)
  return http.ListenAndServe(addr, nil)
}

