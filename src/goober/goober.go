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
  handler http.Handler
  children map[string]*routeTreeNode
  dynamicChildren map[string]*routeTreeNode
}

func newRouteTreeNode() (node *routeTreeNode) {
  node = new(routeTreeNode)
  node.children = make(map[string]*routeTreeNode)
  node.dynamicChildren = make(map[string]*routeTreeNode)

  return
}

func New() (* Goober) {
  var head = make(map[string]*routeTreeNode)
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

func (g *Goober) AddHandler(method string, route string, handler http.Handler) (err int){
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

func (g *Goober) Get(route string, handler http.Handler) {
  g.AddHandler("GET", route, handler)
}

func (g *Goober) Post(route string, handler http.Handler) {
  g.AddHandler("POST", route, handler)
}

func (g *Goober) Put(route string, handler http.Handler) {
  g.AddHandler("PUT", route, handler)
}

func (g *Goober) Delete(route string, handler http.Handler) {
  g.AddHandler("DELETE", route, handler)
}

func walkTree(node map[string]*routeTreeNode, parts *[]string) (handler http.Handler, err int) {
  if len(parts) == 0 {

  } else {

  }
}

func (g *Goober) GetHandlers(method string, path string) {
  path = strings.TrimFunc(path, isSlash)
  var parts = strings.Split(path, "/")
  for i := range parts {
    fmt.Println(parts[i])
  }
}

