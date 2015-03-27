package gocelot

import (
	"testing"
)

func TestNewHandlerArrayCreatesEmptyHandlerArray(t *testing.T) {
	array := newHandlerArray()
	if array == nil || array.nodes != nil {
		t.Fail()
	}
}

func TestAddOfNewMethodAddsItToTheEndOfTheArray(t *testing.T) {
	array := newHandlerArray()
	array.add("GET", emptyHandler)
	if array.nodes == nil || len(array.nodes) != 1 || array.nodes[0].method != "GET" || array.nodes[0].handler != emptyHandler {
		t.Fail()
	}
}

func TestAddOfExistingMethodDoesNothing(t *testing.T) {
	array := newHandlerArray()
	array.add("GET", emptyHandler)
	array.add("GET", differentEmptyHandler)
	if array.nodes == nil || len(array.nodes) != 1 || array.nodes[0].method != "GET" || array.nodes[0].handler != emptyHandler {
		t.Fail()
	}
}

func TestGetOnAnEmptyHandlerArrayReturnsNil(t *testing.T) {
	array := newHandlerArray()
	handler := array.get("GET")
	if handler != nil {
		t.Fail()
	}
}

func TestGetOfNonExistingMethodReturnsNil(t *testing.T) {
	array := newHandlerArray()
	array.add("GET", emptyHandler)
	handler := array.get("POST")
	if handler != nil {
		t.Fail()
	}
}

func TestGetOfExistingMethodReturnsCorrespondingHandler(t *testing.T) {
	array := newHandlerArray()
	array.add("GET", emptyHandler)
	handler := array.get("GET")
	if handler != emptyHandler {
		t.Fail()
	}
}
