package objtree

import (
	"github.com/godbus/dbus"
	"testing"
	"time"
)

func TestNewSessionBusManager(t *testing.T) {
	expected := "hello, world"
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	bus, err := NewSessionBusManager("com.github.jsouthworth.objtree.Test")
	if err != nil {
		t.Fatal(err)
	}
	obj := bus.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err = obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	outs, err := obj.Call("foo", "CallMe")
	if err != nil {
		t.Fatal(err)
	}
	if outs[0].(string) != expected {
		t.Fatal("got:", outs[0].(string), "expected:", expected)
	}
}

func TestNewSessionBusManagerBogusName(t *testing.T) {
	_, err := NewSessionBusManager("foo")
	if err == nil {
		t.Fatal("name was invalid this should have passed")
	}
}

func TestNewAnonymousSessionBusManager(t *testing.T) {
	expected := "hello, world"
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	bus, err := NewAnonymousSessionBusManager()
	if err != nil {
		t.Fatal(err)
	}
	err = bus.RequestName("com.github.jsouthworth.objtree.Test")
	if err != nil {
		t.Fatal(err)
	}
	obj := bus.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err = obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	outs, err := obj.Call("foo", "CallMe")
	if err != nil {
		t.Fatal(err)
	}
	if outs[0].(string) != expected {
		t.Fatal("got:", outs[0].(string), "expected:", expected)
	}
}

func TestNewSystemBusManager(t *testing.T) {
	t.Skip("System policies can prevent this test from passing")
	expected := "hello, world"
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	bus, err := NewSystemBusManager("com.github.jsouthworth.objtree.Test")
	if err != nil {
		t.Fatal(err)
	}
	obj := bus.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err = obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	outs, err := obj.Call("foo", "CallMe")
	if err != nil {
		t.Fatal(err)
	}
	if outs[0].(string) != expected {
		t.Fatal("got:", outs[0].(string), "expected:", expected)
	}
}

func TestNewAnonymousSystemBusManager(t *testing.T) {
	t.Skip("System policies can prevent this test from passing")
	expected := "hello, world"
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	bus, err := NewAnonymousSystemBusManager()
	if err != nil {
		t.Fatal(err)
	}
	err = bus.RequestName("com.github.jsouthworth.objtree.Test")
	if err != nil {
		t.Fatal(err)
	}
	obj := bus.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err = obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	outs, err := obj.Call("foo", "CallMe")
	if err != nil {
		t.Fatal(err)
	}
	if outs[0].(string) != expected {
		t.Fatal("got:", outs[0].(string), "expected:", expected)
	}
}

func TestBusManagerCall(t *testing.T) {
	expected := "hello, world"
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	bus, err := NewSessionBusManager("com.github.jsouthworth.objtree.Test")
	if err != nil {
		t.Fatal(err)
	}
	obj := bus.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err = obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	outs, err := bus.Call("/foo/bar/call", "foo", "CallMe")
	if err != nil {
		t.Fatal(err)
	}
	if outs[0].(string) != expected {
		t.Fatal("got:", outs[0].(string), "expected:", expected)
	}
}

func TestBusManagerCallNonExistent(t *testing.T) {
	bus, err := NewSessionBusManager("com.github.jsouthworth.objtree.Test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = bus.Call("/foo/baz/call", "foo", "CallMe")
	if err == nil {
		t.Fatal("Expected error did not occur")
	}
}

func TestBusManagerLookupObject(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	bus, err := NewSessionBusManager("com.github.jsouthworth.objtree.Test")
	if err != nil {
		t.Fatal(err)
	}
	obj := bus.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err = obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := bus.LookupObject("/foo/bar/call")
	if !ok {
		t.Fatal("expected to find object")
	}

}

func TestBusManagerLookupRootObject(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	bus, err := NewSessionBusManager("com.github.jsouthworth.objtree.Test")
	if err != nil {
		t.Fatal(err)
	}
	obj := bus.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err = obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := bus.LookupObject("/")
	if !ok {
		t.Fatal("expected to find object")
	}
}

func TestBusManagerDeleteReceiver(t *testing.T) {
	ch := make(chan string)
	root, err := NewSessionBusManager("com.github.jsouthworth.objtree.Test")
	if err != nil {
		t.Fatal(err)
	}
	methods := map[string]interface{}{
		"CallMe": func(ins ...interface{}) {
			ch <- ins[0].(string)
		},
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err = obj.ReceivesTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	obj2 := root.NewObjectFromTable("/foo/baz/call", methods)
	if obj2 == nil {
		t.Fatal("unexpected nil")
	}
	err = obj2.ReceivesTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	root.DeleteObject("/foo/baz/call")
	expected := "hello, world"
	sig := &dbus.Signal{
		Body: []interface{}{expected},
	}
	root.DeliverSignal("foo", "CallMe", sig)
	got := <-ch
	if got != expected {
		t.Fatal("expected:", expected, "got:", got)
	}
	select {
	case <-ch:
		t.Fatal("expected timeout to occur")
	case <-time.After(time.Second):
	}
}

func TestBusManagerDeleteAllReceivers(t *testing.T) {
	ch := make(chan string)
	root, err := NewSessionBusManager("com.github.jsouthworth.objtree.Test")
	if err != nil {
		t.Fatal(err)
	}
	methods := map[string]interface{}{
		"CallMe": func(ins ...interface{}) {
			ch <- ins[0].(string)
		},
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err = obj.ReceivesTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	obj2 := root.NewObjectFromTable("/foo/baz/call", methods)
	if obj2 == nil {
		t.Fatal("unexpected nil")
	}
	err = obj2.ReceivesTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	root.DeleteObject("/foo/baz/call")
	root.DeleteObject("/foo/bar/call")
	expected := "hello, world"
	sig := &dbus.Signal{
		Body: []interface{}{expected},
	}
	root.DeliverSignal("foo", "CallMe", sig)

	select {
	case <-ch:
		t.Fatal("expected timeout to occur")
	case <-time.After(time.Second):
	}
	select {
	case <-ch:
		t.Fatal("expected timeout to occur")
	case <-time.After(time.Second):
	}
}
