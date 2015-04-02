# gocelot
[![Build Status](https://drone.io/github.com/gfjalar/gocelot/status.png?branch=master)](https://drone.io/github.com/gfjalar/gocelot/latest)
[![Coverage Status](https://coveralls.io/repos/gfjalar/gocelot/badge.svg?branch=master)](https://coveralls.io/r/gfjalar/gocelot?branch=master)
[![GoDoc](https://godoc.org/github.com/gfjalar/gocelot?status.svg)](https://godoc.org/github.com/gfjalar/gocelot)

Gocelot is a url router for Go. It supports url parameters and stores them in
request.Form. The creation of this router was inspired by 
https://github.com/julienschmidt/httprouter and started of as its fork. However,
further along the way I decied to implement my own prefix tree based router.

### How does it work?
Gocelot was designed to be minimal and as such offers creating new router and
adding a handler for a specific path and method. The router itself implements
http.Handler interface which means it can be used straight out of the box. Also
the path/method handlers conform to http.Handler interface.

##### What paths does the router accept?
The router accepts both static paths and paths with parameters. All the paths
must start with '/' and all the parameters are identified by ':' before the
parameter name. There is no support for wildcard parameters which means that
during the url matching process ':' parameters will only match until the next
occurence of '/' or the end of the url.

##### How are parameters passed to the handler?
The parameters are passed throgh the request object. They are put inside 
http.request.Form field. This field is of type url.Values ie.
map[string][]string where the key is the parameter name and the value is an
array of parameter values. The parameter values are put in the map only if a
matching path/method were found. They are put in the reverse order.
Eg.
```
let path = "/path/:key/:key/:key/:otherKey"
if url = "/path/1/2/3/4" matches the path
	handler is called with
		request.Form = {
			"key": ["3", "2", "1"],
			"otherKey": ["4"]
		}
```

##### How are routes added to the router?
Router's routes are stored in a special prefix tree. A router holds a pointer
to the root of the prefix tree which is always '/'. When route is being added
it is matched against the nodes of the tree.

Now, if the path exactly matches the node path, then no new path is added.

If the node path is a prefix of the path, then the path is split at the length
of the node path and the remainder is matched against the node's children. If
no match is found(ie. none of the children's paths begin with the same letter
as the remainder), the remainder is added as a new child node.

If the path is a prefix of the node path, the node path is split at the length
of the path. The remainder forms a new node which becomes the only child node of
the current node while it inherits all the old children of that node.

If the path and the node path have a common prefix shorter than either of them, 
then the node path is split at that length and two new nodes are created in
the manner described above.

Note that params receive a special treatment in the process of addition. The
process ensures that if a param name is stored in the node, then it is the only
thing stored in this node.
Eg.
```
adding "/path/:key/:key/path/:key"
will create tree "/" -> "path/" -> ":key" -> "/" -> ":key" -> "/path/" -> ":key"
whereas
adding "/path/path/path/path/path"
will create tree "/" -> "path/path/path/path/path"
```

##### How are handlers stored?
Each node has a pointer to an array of method/handler pairs.

##### How are urls matched against the paths?
The router only looks for the exact matches ie. "/path" != "/path/". Once the
exact node containing the correct path was found. The router checks if a handler
for the given method exists. If so, it retrieves the handler, populates the
request with all the parameters values and calls the handler.

##### What happens if path/method isn't found?
You can specify NotFound and MethodNotAllowed handlers for the router. First
is called if there are no handlers at all for the specified path. The second
one if there are some handlers for the specified path, however not for the
specified method. If the MethodNotAllowed is not specified, the router calls
NotFound handler in both cases. By default the router uses http.NotFound as the
NotFound handler.

### Usage

To import:
```go
import "github.com/gfjalar/gocelot"
```

To create new router:
```go
router := gocelot.New()
```

To set NotFound/MethodNotAllowed handlers:
```go
router.NotFound = handler
router.MethodNotAllowed = handler
```

To add path, method, handler:
```go
router.Handle("GET", "/path", handler)
```

To add path, method, handler by handler function:
```go
router.HandleFunc("GET", "/path", handlerFunc)
```

To use router:
```go
http.ListenAndServe(":8080", router)
```

### Example:

```go
package main

import (
	"bytes"
	"net/http"

	"github.com/gfjalar/gocelot"
)

type SimpleHandler struct {
	code int
	message string
}

func (h *SimpleHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	response.WriteHeader(h.code)
	response.Write([]byte(h.message))
}

type MultiParamsHandler struct {
	code int
	message string
	paramsNames []string
}

func (h *MultiParamsHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	buffer := bytes.NewBufferString(h.message)
	for _, paramName := range h.paramsNames {
		buffer.WriteString(" ")
		buffer.WriteString(paramName)
		buffer.WriteString(" ")
		buffer.WriteString(request.Form.Get(paramName))
	}
	response.WriteHeader(h.code)
	response.Write(buffer.Bytes())
}

func UserHandlerFunc(response http.ResponseWriter, request *http.Request) {
	userId := request.Form.Get("id")
	response.WriteHeader(200)
	response.Write([]byte("/users/:id endpoint with id " + userId))
}

func main() {
	router := gocelot.New()

	router.NotFound = &SimpleHandler{404, "404 Not Found"}
	router.MethodNotAllowed = &SimpleHandler{405, "405 Method Not Allowed"}

	router.HandleFunc("GET", "/users", func(response http.ResponseWriter, request *http.Request) {
		response.WriteHeader(200)
		response.Write([]byte("/users endpoint"))
	})
	router.Handle("GET", "/users/0", &SimpleHandler{200, "/users/0 endpoint matching the specific id = 0"})
	router.HandleFunc("GET", "/users/:id", UserHandlerFunc)
	router.Handle("GET", "/users/:id/:param", &MultiParamsHandler{200, "/users/:id/:param endpoint with", []string{"id", "param"}})

	http.ListenAndServe(":8080", router)
}
```

### TODO
* ```go
func (r *Router) Merge(path string, router Router) to merge router on a specific path
```
* ```go
router.PanicHandler
```