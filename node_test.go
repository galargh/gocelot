package gocelot

import (
	"testing"
	"net/http"
)

func TestNewNodeCreatesEmptyNode(t *testing.T) {
	node := newNode()
	if node == nil || node.path != "" || node.next != nil ||
		node.handlers != nil {

		t.Fail()
	}
}

func isNodeSeqCorrect(first, last *node, segments ...string) bool {
	i := 0
	node := first
	for node != last && i < len(segments) {
		if node.path != segments[i] || node.next == nil || len(node.next) != 1 {
			return false
		}
		node = node.next[0]
		i++
	}
	if i == len(segments) {
		return false
	}
	if last.path != segments[i] || last.next != nil {
		return false
	}
	return true
}

func TestNodeSeqForEmptyPathReturnsTwoEmptyNodes(t *testing.T) {
	first, last := nodeSeq("")
	if !isNodeSeqCorrect(first, last, "") {
		t.Fail()
	}
}

func TestNodeSeqForPathWithoutParamsReturnsTwoSameNodes(t *testing.T) {
	path := "/some/path/without/params"
	first, last := nodeSeq(path)
	if !isNodeSeqCorrect(first, last, path) {
		t.Fail()
	}
}

func TestNodeSeqForPathEndingWithOpenParam(t *testing.T) {
	path := "/path/:param"
	first, last := nodeSeq(path)
	if !isNodeSeqCorrect(first, last, "/path/", ":param") {
		t.Fail()
	}
}

func TestNodeSeqForPathEndingWithClosedParam(t *testing.T) {
	path := "/path/:param/"
	first, last := nodeSeq(path)
	if !isNodeSeqCorrect(first, last, "/path/", ":param", "/") {
		t.Fail()
	}

}

func TestNodeSeqForParamOnlyPath(t *testing.T) {
	path := ":param"
	first, last := nodeSeq(path)
	if !isNodeSeqCorrect(first, last, path) {
		t.Fail()
	}
}

func TestNodeSeqForPathWithMultipleParams(t *testing.T) {
	path := "path/:a/:b/:c/break/:d/:e/end"
	first, last := nodeSeq(path)
	if !isNodeSeqCorrect(first, last, "path/", ":a", "/", ":b", "/", ":c",
		"/break/", ":d", "/", ":e", "/end") {

		t.Fail()
	}
}

func TestMinReturnsMinOfTwoInts(t *testing.T) {
	if min(3, 8) != 3 || min(8, 3) != 3 {
		t.Fail()
	}
}

func TestLcpWithEmptyString(t *testing.T) {
	if lcp("", "path") != 0 || lcp("path", "") != 0 || lcp("", "") != 0 {
		t.Fail()
	}
}

func TestLcpShorterThanBothStrings(t *testing.T) {
	if lcp("path", "part") != 2 || lcp("part", "path") != 2 {
		t.Fail()
	}
}

func TestLcpOfLengthOfOneOfTheStrings(t *testing.T) {
	if lcp("path", "paths") != 4 || lcp("paths", "path") != 4 {
		t.Fail()
	}
}

func TestLcpOfEqualWords(t *testing.T) {
	if lcp("path", "path") != 4 {
		t.Fail()
	}
}

func TestIndexOfInEmptyString(t *testing.T) {
	if indexOf("", '/') != 0 {
		t.Fail()
	}
}

func TestIndexOfNonExistingRune(t *testing.T) {
	if indexOf("path", '/') != 4 {
		t.Fail()
	}
}

func TestIndexOfExistingRune(t *testing.T) {
	if indexOf("path/", '/') != 4 {
		t.Fail()
	}
}

func TestHandleAddsHandlerNodeToHandlers(t *testing.T) {
	node := newNode()
	node.handle("GET", emptyHandler)
	if node.handlers == nil || len(node.handlers.nodes) != 1 ||
		node.handlers.nodes[0].method != "GET" ||
		node.handlers.nodes[0].handler != emptyHandler {

		t.Fail()
	}
}

func TestHandleDoesNotAddHandlerForExistingMethod(t *testing.T) {
	node := newNode()
	node.handle("GET", emptyHandler)
	node.handle("GET", differentEmptyHandler)
	if node.handlers == nil || len(node.handlers.nodes) != 1 ||
		node.handlers.nodes[0].method != "GET" ||
		node.handlers.nodes[0].handler != emptyHandler {

		t.Fail()
	}
}

func TestAddParamAddsParamIfHandlerIsNotNil(t *testing.T) {
	request, _ := http.NewRequest("GET", "/path", nil)
	addParam(request, emptyHandler, "key", "value")
	if request.Form == nil || request.Form.Get("key") != "value" {
		t.Fail()
	}
}

func TestAddParamDoesNotAddParamIfHandlerIsNil(t *testing.T) {
	request, _ := http.NewRequest("GET", "/path", nil)
	addParam(request, nil, "key", "value")
	if request.Form != nil {
		t.Fail()
	}
}

func TestAddParamAddsParamToTheEndOfUrlValuesArray(t *testing.T) {
	request, _ := http.NewRequest("GET", "/path", nil)
	addParam(request, emptyHandler, "key", "value")
	addParam(request, emptyHandler, "key", "differentValue")
	if request.Form == nil || request.Form["key"] == nil ||
		len(request.Form["key"]) != 2 || request.Form["key"][0] != "value" ||
		request.Form["key"][1] != "differentValue" {
	
		t.Fail()
	}

}

func TestAddOfTheSamePathReturnsTheNode(t *testing.T) {
	node := newNode()
	node.path = "/path"
	added := node.add("/path")
	if added == nil || added != node {
		t.Fail()
	}
}

func TestAddOfDifferentLongerParamReturnsNil(t *testing.T) {
	node := newNode()
	node.path = ":param"
	added := node.add(":params")
	if added != nil {
		t.Fail()
	}
}

func TestAddOfDifferentShorterParamReturnsNil(t *testing.T) {
	node := newNode()
	node.path = ":param"
	added := node.add(":par")
	if added != nil {
		t.Fail()
	}
}

func TestAddOfSamePathReturnsTheSameLastNode(t *testing.T) {
	node, last := nodeSeq("/path/:param/end")
	added := node.add("/path/:param/end")
	if added == nil || added != last {
		t.Fail()
	}
}

func TestAddOfShorterPathReturnsTheNodeWithChangedPath(t *testing.T) {
	node := newNode()
	node.path = "/path/end"
	added := node.add("/path")
	if added == nil || added != node || added.path != "/path" {
		t.Fail()
	}
}

func TestAddOfDifferentPathReturnsTheEndOfNewNodeSeq(t *testing.T) {
	node := newNode()
	node.path = "/path/one"
	added := node.add("/path/two/:param/end")
	if added == nil || added == node || added.path != "/end" ||
		node.path != "/path/" {
		
		t.Fail()
	}
}

func TestGetOfOpenParamReturnsHandlerAndTrue(t *testing.T) {
	node := newNode()
	node.path = ":key"
	request, _ := http.NewRequest("GET", "value", nil)
	node.handle(request.Method, emptyHandler)
	handler, pathFound := node.get(request.URL.Path, request.Method, request)
	if handler != emptyHandler || !pathFound || request.Form == nil ||
		request.Form.Get("key") != "value" {
		
		t.Fail()
	}
}

func TestGetOfClosedParamReturnsHandlerAndTrue(t *testing.T) {
	node, last := nodeSeq(":key/")
	request, _ := http.NewRequest("GET", "value/", nil)
	last.handle(request.Method, emptyHandler)
	handler, pathFound := node.get(request.URL.Path, request.Method, request)
	if handler != emptyHandler || !pathFound || request.Form == nil ||
		request.Form.Get("key") != "value" {
		
		t.Fail()
	}
}

func TestGetOfNonMatchingParamReturnsNilAndFalse(t *testing.T) {
	node := newNode()
	node.path = ":key"
	request, _ := http.NewRequest("GET", "value/", nil)
	node.handle(request.Method, emptyHandler)
	handler, pathFound := node.get(request.URL.Path, request.Method, request)
	if handler != nil || pathFound || request.Form != nil {
		t.Fail()
	}
}

func TestGetOfMatchingPathReturnsHandlerAndTrue(t *testing.T) {
	node := newNode()
	node.path = "/"
	request, _ := http.NewRequest("GET", "/", nil)
	node.handle(request.Method, emptyHandler)
	handler, pathFound := node.get(request.URL.Path, request.Method, request)
	if handler != emptyHandler || !pathFound || request.Form != nil {
		t.Fail()
	}
}

func TestGetOfNonMatchingPathReturnsNilAndFalse(t *testing.T) {
	node := newNode()
	node.path = "/path/"
	request, _ := http.NewRequest("GET", "/path", nil)
	node.handle(request.Method, emptyHandler)
	handler, pathFound := node.get(request.URL.Path, request.Method, request)
	if handler != nil || pathFound {
		t.Fail()
	}
}

func TestGetOfLongerMatchingPathReturnsHandlerAndTrue(t *testing.T) {
	node := newNode()
	node.path = "/"
	nextNode := newNode()
	nextNode.path = "path"
	node.next = append(node.next, nextNode)
	request, _ := http.NewRequest("GET", "/path", nil)
	nextNode.handle(request.Method, emptyHandler)
	handler, pathFound := node.get(request.URL.Path, request.Method, request)
	if handler != emptyHandler || !pathFound || request.Form != nil {
		t.Fail()
	}
}

func TestGetOfMatchingPathWithoutHandlersReturnsNilAndFalse(t *testing.T) {
	node := newNode()
	node.path = "/"
	request, _ := http.NewRequest("GET", "/", nil)
	handler, pathFound := node.get(request.URL.Path, request.Method, request)
	if handler != nil || pathFound || request.Form != nil {
		t.Fail()
	}
}


func TestGetOfMatchingParamWithoutHandlersReturnsNilAndFalse(t *testing.T) {
	node := newNode()
	node.path = ":key"
	request, _ := http.NewRequest("GET", "value", nil)
	handler, pathFound := node.get(request.URL.Path, request.Method, request)
	if handler != nil || pathFound || request.Form != nil {
		t.Fail()
	}
}

func TestGetOfMatchingPathWithDifferentHandlerReturnsNilAndTrue(t *testing.T) {
	node := newNode()
	node.path = "/"
	request, _ := http.NewRequest("GET", "/", nil)
	node.handle("POST", emptyHandler)
	handler, pathFound := node.get(request.URL.Path, request.Method, request)
	if handler != nil || !pathFound || request.Form != nil {
		t.Fail()
	}
}

func TestGetOfMatchingParamWithDifferentHandlerReturnsNilAndTrue(t *testing.T) {
	node := newNode()
	node.path = ":key"
	request, _ := http.NewRequest("GET", "value", nil)
	node.handle("POST", emptyHandler)
	handler, pathFound := node.get(request.URL.Path, request.Method, request)
	if handler != nil || !pathFound || request.Form != nil {
		t.Fail()
	}
}
