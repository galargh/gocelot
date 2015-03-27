package gocelot

import (
	"testing"
	"net/http"
	"net/http/httptest"
)

type emptyHandlerStruct struct {}
func (h *emptyHandlerStruct) ServeHTTP(response http.ResponseWriter, request *http.Request) {}
var emptyHandler *emptyHandlerStruct = &emptyHandlerStruct{}
var differentEmptyHandler *emptyHandlerStruct = &emptyHandlerStruct{}

type failHandlerStruct struct {
	t *testing.T
}
func (h *failHandlerStruct) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	h.t.Fail()
}

func TestNewCreatesRouterWithTreeSet(t *testing.T) {
	router := New()
	if router == nil || router.tree.path != "/" || router.NotFound != nil ||
		router.MethodNotAllowed != nil {
		
		t.Fail()
	}
}

func TestHandleAddsPathAndHandlerToTheTree(t *testing.T) {
	router := New()
	router.Handle("GET", "/path", emptyHandler) 
	root := router.tree
	if root.next == nil || len(root.next) != 1 || root.next[0].path != "path" ||
		root.next[0].handlers.get("GET") != emptyHandler {

		t.Fail()
	}
}

func TestHandlerFuncCreatesHandlerAddsPathAndHandlerToTheTree(t *testing.T) {
	router := New()
	router.HandleFunc("GET", "/path", emptyHandler.ServeHTTP) 
	root := router.tree
	if root.next == nil || len(root.next) != 1 || root.next[0].path != "path" ||
		root.next[0].handlers.get("GET") == nil {

		t.Fail()
	}
}



func TestServeHTTPUsesHandlerIfFound(t *testing.T) {
	failHandler := &failHandlerStruct{t}
	router := New()
	router.Handle("GET", "/", emptyHandler)
	router.MethodNotAllowed = failHandler
	router.NotFound = failHandler

	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
}

func TestServeHTTPUsesMethodNotAllowedHandlerIfOnlyPathFound(t *testing.T) {
	failHandler := &failHandlerStruct{t}
	router := New()
	router.Handle("POST", "/", failHandler)
	router.MethodNotAllowed = emptyHandler
	router.NotFound = failHandler

	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
}

func TestServeHTTPUsesNotFoundHandlerIfNoPathFound(t *testing.T) {
	failHandler := &failHandlerStruct{t}
	router := New()
	router.MethodNotAllowed = failHandler
	router.NotFound = emptyHandler

	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
}

func TestServeHTTPUsesHTTPNotFoundByDefaultIfNoPathFound(t *testing.T) {
	failHandler := &failHandlerStruct{t}
	router := New()
	router.MethodNotAllowed = failHandler

	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
}
