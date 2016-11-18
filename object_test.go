package objtree

import (
	"bytes"
	"encoding/xml"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"reflect"
	"testing"
	"time"
)

type testIface interface {
	CallMe() string
}

type testObj struct{}

func (*testObj) CallMe() string { return "hello, world" }

func TestNewObjectAtRoot(t *testing.T) {
	const introExpected = `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
			 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd"><node></node>`
	root := newObjectFromImpl("", nil, nil, nil)
	obj := root.NewObject("/", &testObj{})
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	outs, err := root.Call(fdtIntrospectable, "Introspect")
	if err != nil {
		t.Fatal(err)
	}
	expectedNode := decodeIntrospection(introExpected)
	gotNode := decodeIntrospection(outs[0].(string))
	if !reflect.DeepEqual(expectedNode, gotNode) {
		t.Fatalf("expected:\n%s\ngot:\n%s", introExpected, outs[0].(string))
	}

}

func TestNewObjectMapAtRoot(t *testing.T) {
	const introExpected = `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
			 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd"><node></node>`
	root := newObjectFromImpl("", nil, nil, nil)
	obj := root.NewObjectMap("/", &testObj{},
		func(in string) string {
			if in == "CallMe" {
				return "call-me"
			}
			return in
		})
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	outs, err := root.Call(fdtIntrospectable, "Introspect")
	if err != nil {
		t.Fatal(err)
	}
	expectedNode := decodeIntrospection(introExpected)
	gotNode := decodeIntrospection(outs[0].(string))
	if !reflect.DeepEqual(expectedNode, gotNode) {
		t.Fatalf("expected:\n%s\ngot:\n%s", introExpected, outs[0].(string))
	}

}

func TestNewObjectFromTableAtRoot(t *testing.T) {
	const introExpected = `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
			 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd"><node></node>`
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := root.NewObjectFromTable("/", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	outs, err := root.Call(fdtIntrospectable, "Introspect")
	if err != nil {
		t.Fatal(err)
	}
	expectedNode := decodeIntrospection(introExpected)
	gotNode := decodeIntrospection(outs[0].(string))
	if !reflect.DeepEqual(expectedNode, gotNode) {
		t.Fatalf("expected:\n%s\ngot:\n%s", introExpected, outs[0].(string))
	}

}

func TestTableObjectCall(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := newObjectFromTable("foo", methods, nil, nil)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	err := obj.Implements("foo", (*testIface)(nil))
	if err != nil {
		t.Fatal(err)
	}
	iface, exists := obj.LookupInterface("foo")
	if !exists {
		t.Fatal("export failed")
	}
	method, exists := iface.LookupMethod("CallMe")
	if !exists {
		t.Fatal("export failed")
	}

	outs, err := method.Call()
	if err != nil {
		t.Fatal(err)
	}
	if outs[0].(string) != "hello, world" {
		t.Fatal("didn't get expected output")
	}
}

type testNonMatchingFunc interface {
	CallMe(string) string
}

func TestTableObjectImplementsNonMatchingFunc(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := newObjectFromTable("foo", methods, nil, nil)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	err := obj.Implements("foo", (*testNonMatchingFunc)(nil))
	if err == nil {
		t.Fatal("Should have failed")
	}
}

func TestTableObjectImplementsNonMatchingFuncWrongName(t *testing.T) {
	methods := map[string]interface{}{
		"Call": interface{}(func(string) string { return "hello, world" }),
	}
	obj := newObjectFromTable("foo", methods, nil, nil)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	err := obj.Implements("foo", (*testNonMatchingFunc)(nil))
	if err == nil {
		t.Fatal(err)
		t.Fatal("Should have failed")
	}
}

type testTooManyMethods interface {
	CallMe() string
	CallMe2() string
}

func TestTableObjectImplementsTooManyMethods(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := newObjectFromTable("foo", methods, nil, nil)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	err := obj.Implements("foo", (*testTooManyMethods)(nil))
	if err == nil {
		t.Fatal("Should have failed")
	}
}

type testMismatchedTypes interface {
	CallMe() bool
}

func TestTableObjectImplementsMismatchedTypes(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := newObjectFromTable("foo", methods, nil, nil)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	err := obj.Implements("foo", (*testMismatchedTypes)(nil))
	if err == nil {
		t.Fatal("Should have failed")
	}
}

type testTooManyOutputs interface {
	CallMe() (string, bool)
}

func TestTableObjectImplementsTooManyOutputs(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := newObjectFromTable("foo", methods, nil, nil)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	err := obj.Implements("foo", (*testTooManyOutputs)(nil))
	if err == nil {
		t.Fatal("Should have failed")
	}
}

func TestTableObjectImplementsMoreThanOneFunction(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe":  interface{}(func() string { return "hello, world" }),
		"CallMe2": interface{}(func() string { return "hello, world2" }),
	}
	obj := newObjectFromTable("foo", methods, nil, nil)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	err := obj.Implements("foo", (*testIface)(nil))
	if err != nil {
		t.Fatal(err)
	}
}

func decodeIntrospection(intro string) *introspect.Node {
	var node introspect.Node
	buf := bytes.NewBufferString(intro)
	dec := xml.NewDecoder(buf)
	dec.Decode(&node)
	return &node
}

func TestTableObjectIntrospection(t *testing.T) {
	const introExpected = `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
			 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd"><node><interface name="foo"><method name="CallMe"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Peer"><method name="GetMachineId"><arg type="s" direction="out"></arg></method><method name="Ping"></method></interface></node>`

	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := newObjectFromTable("foo", methods, nil, nil)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	err := obj.Implements("foo", (*testIface)(nil))
	if err != nil {
		t.Fatal(err)
	}
	iface, exists := obj.LookupInterface(fdtIntrospectable)
	if !exists {
		t.Fatal("Not intropsectable")
	}
	method, exists := iface.LookupMethod("Introspect")
	if !exists {
		t.Fatal("export failed")
	}

	outs, err := method.Call()
	if err != nil {
		t.Fatal(err)
	}
	expectedNode := decodeIntrospection(introExpected)
	gotNode := decodeIntrospection(outs[0].(string))
	if !reflect.DeepEqual(expectedNode, gotNode) {
		t.Fatalf("expected:\n%s\ngot:\n%s", introExpected, outs[0].(string))
	}
}

func TestTableObjectNoReturns(t *testing.T) {
	const introExpected = `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
			 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd"><node><interface name="foo"><method name="CallMe"></method></interface><interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Peer"><method name="GetMachineId"><arg type="s" direction="out"></arg></method><method name="Ping"></method></interface></node>`

	methods := map[string]interface{}{
		"CallMe": func() {},
	}
	obj := newObjectFromTable("foo", methods, nil, nil)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	err := obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	iface, exists := obj.LookupInterface(fdtIntrospectable)
	if !exists {
		t.Fatal("Not intropsectable")
	}
	method, exists := iface.LookupMethod("Introspect")
	if !exists {
		t.Fatal("export failed")
	}

	outs, err := method.Call()
	if err != nil {
		t.Fatal(err)
	}
	expectedNode := decodeIntrospection(introExpected)
	gotNode := decodeIntrospection(outs[0].(string))
	if !reflect.DeepEqual(expectedNode, gotNode) {
		t.Fatalf("expected:\n%s\ngot:\n%s", introExpected, outs[0].(string))
	}
}

func TestTableObjectBogusMethod(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": "foobar",
	}
	obj := newObjectFromTable("foo", methods, nil, nil)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	err := obj.Implements("foo", (*testIface)(nil))
	if err == nil {
		t.Fatal("Object should not implement testIface")
	}
}

func TestImplementsTable(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := newObjectFromTable("foo", methods, nil, nil)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.Implements("foo", (*testIface)(nil))
	if err != nil {
		t.Fatal(err)
	}
}

func TestImplementsTableWithTable(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := newObjectFromTable("foo", methods, nil, nil)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
}

func TestImplementsTableWithTableNoMatch(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	iface_methods := map[string]interface{}{
		"CallMe":  interface{}(func() string { return "hello, world" }),
		"CallMe2": interface{}(func() string { return "hello, world" }),
	}
	obj := newObjectFromTable("foo", methods, nil, nil)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ImplementsTable("foo", iface_methods)
	if err == nil {
		t.Fatal("Object should not have implemented the interface")
	}
}

func TestObjectPath(t *testing.T) {
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
}

func TestObjectPathCall(t *testing.T) {
	expected := "hello, world"
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return expected }),
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ImplementsTable("foo", methods)
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

func TestObjectPathBogus(t *testing.T) {
	expected := "hello, world"
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return expected }),
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	_, err = obj.Call("foo", "Bogus")
	if err == nil {
		t.Fatal("Call should have failed on Bogus method")
	}
}

func TestObjectPathBogusInterface(t *testing.T) {
	expected := "hello, world"
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return expected }),
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	_, err = obj.Call("Bogus", "CallMe")
	if err == nil {
		t.Fatal("Call should have failed on Bogus interface")
	}
}

func TestObjectPathMultipleObjects(t *testing.T) {
	const introExpected = `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
			 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd"><node><node name="foo"><node name="bar"><interface name="foo"><method name="CallMe"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Peer"><method name="GetMachineId"><arg type="s" direction="out"></arg></method><method name="Ping"></method></interface><node name="call"><interface name="foo"><method name="CallMe"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Peer"><method name="GetMachineId"><arg type="s" direction="out"></arg></method><method name="Ping"></method></interface></node></node><node name="baz"><interface name="foo"><method name="call-me"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Peer"><method name="GetMachineId"><arg type="s" direction="out"></arg></method><method name="Ping"></method></interface></node></node></node>`

	mapper := func(in string) string {
		if in == "CallMe" {
			return "call-me"
		}
		return in
	}
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	obj2 := root.NewObject("/foo/bar", &testObj{})
	if obj2 == nil {
		t.Fatal("unexpected nil")
	}
	err = obj2.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}

	obj3 := root.NewObjectMap("/foo/baz", &testObj{}, mapper)
	if obj3 == nil {
		t.Fatal("unexpected nil")
	}
	err = obj3.ImplementsMap("foo", (*testIface)(nil), mapper)
	if err != nil {
		t.Fatal(err)
	}

	outs, err := root.Call(fdtIntrospectable, "Introspect")
	if err != nil {
		t.Fatal(err)
	}
	expectedNode := decodeIntrospection(introExpected)
	gotNode := decodeIntrospection(outs[0].(string))
	if !reflect.DeepEqual(expectedNode, gotNode) {
		t.Fatalf("expected:\n%s\ngot:\n%s", introExpected, outs[0].(string))
	}
}

func TestObjectPathIntrospect(t *testing.T) {
	const introExpected = `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
			 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd"><node><node name="foo"><node name="bar"><node name="call"><interface name="foo"><method name="CallMe"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Peer"><method name="GetMachineId"><arg type="s" direction="out"></arg></method><method name="Ping"></method></interface></node></node></node></node>`
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	outs, err := root.Call(fdtIntrospectable, "Introspect")
	if err != nil {
		t.Fatal(err)
	}
	expectedNode := decodeIntrospection(introExpected)
	gotNode := decodeIntrospection(outs[0].(string))
	if !reflect.DeepEqual(expectedNode, gotNode) {
		t.Fatalf("expected:\n%s\ngot:\n%s", introExpected, outs[0].(string))
	}

}

func TestObjectPathDeleteSingleObject(t *testing.T) {
	const introExpected = `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
			 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd"><node></node>`
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	root.DeleteObject("/foo/bar/call")
	outs, err := root.Call(fdtIntrospectable, "Introspect")
	if err != nil {
		t.Fatal(err)
	}
	expectedNode := decodeIntrospection(introExpected)
	gotNode := decodeIntrospection(outs[0].(string))
	if !reflect.DeepEqual(expectedNode, gotNode) {
		t.Fatalf("expected:\n%s\ngot:\n%s", introExpected, outs[0].(string))
	}

}

func TestObjectPathDeleteNotExists(t *testing.T) {
	const introExpected = `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
			 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd"><node><node name="foo"><node name="bar"><node name="call"><interface name="foo"><method name="CallMe"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Peer"><method name="GetMachineId"><arg type="s" direction="out"></arg></method><method name="Ping"></method></interface></node></node></node></node>`
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	root.DeleteObject("/foo/bar/call2")
	outs, err := root.Call(fdtIntrospectable, "Introspect")
	if err != nil {
		t.Fatal(err)
	}
	expectedNode := decodeIntrospection(introExpected)
	gotNode := decodeIntrospection(outs[0].(string))
	if !reflect.DeepEqual(expectedNode, gotNode) {
		t.Fatalf("expected:\n%s\ngot:\n%s", introExpected, outs[0].(string))
	}

}

func TestObjectPathDeleteMultipleObjects(t *testing.T) {
	const introExpected = `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
			 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd"><node><node name="foo"><node name="bar"><interface name="foo"><method name="CallMe"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Peer"><method name="GetMachineId"><arg type="s" direction="out"></arg></method><method name="Ping"></method></interface></node></node></node>`

	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	obj2 := root.NewObjectFromTable("/foo/bar", methods)
	if obj2 == nil {
		t.Fatal("unexpected nil")
	}
	err = obj2.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	root.DeleteObject("/foo/bar/call")
	outs, err := root.Call(fdtIntrospectable, "Introspect")
	if err != nil {
		t.Fatal(err)
	}
	expectedNode := decodeIntrospection(introExpected)
	gotNode := decodeIntrospection(outs[0].(string))
	if !reflect.DeepEqual(expectedNode, gotNode) {
		t.Fatalf("expected:\n%s\ngot:\n%s", introExpected, outs[0].(string))
	}
}

func TestObjectPathDeleteMultipleObjectsMiddle(t *testing.T) {
	const introExpected = `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
			 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd"><node><node name="foo"><node name="bar"><node name="call"><interface name="foo"><method name="CallMe"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Peer"><method name="GetMachineId"><arg type="s" direction="out"></arg></method><method name="Ping"></method></interface></node></node></node></node>`

	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	obj2 := root.NewObjectFromTable("/foo/bar", methods)
	if obj2 == nil {
		t.Fatal("unexpected nil")
	}
	err = obj2.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	root.DeleteObject("/foo/bar")
	outs, err := root.Call(fdtIntrospectable, "Introspect")
	if err != nil {
		t.Fatal(err)
	}
	expectedNode := decodeIntrospection(introExpected)
	gotNode := decodeIntrospection(outs[0].(string))
	if !reflect.DeepEqual(expectedNode, gotNode) {
		t.Fatalf("expected:\n%s\ngot:\n%s", introExpected, outs[0].(string))
	}
}

func TestObjectPathDeleteMultipleObjectsNonAction(t *testing.T) {
	const introExpected = `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
			 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd"><node><node name="foo"><node name="bar"><interface name="foo"><method name="CallMe"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Peer"><method name="GetMachineId"><arg type="s" direction="out"></arg></method><method name="Ping"></method></interface><node name="call"><interface name="foo"><method name="CallMe"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Peer"><method name="GetMachineId"><arg type="s" direction="out"></arg></method><method name="Ping"></method></interface></node></node></node></node>`

	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	obj2 := root.NewObjectFromTable("/foo/bar", methods)
	if obj2 == nil {
		t.Fatal("unexpected nil")
	}
	err = obj2.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	root.DeleteObject("/foo")
	outs, err := root.Call(fdtIntrospectable, "Introspect")
	if err != nil {
		t.Fatal(err)
	}
	expectedNode := decodeIntrospection(introExpected)
	gotNode := decodeIntrospection(outs[0].(string))
	if !reflect.DeepEqual(expectedNode, gotNode) {
		t.Fatalf("expected:\n%s\ngot:\n%s", introExpected, outs[0].(string))
	}
}

func TestObjectPathDeleteMultipleObjectsRoot(t *testing.T) {
	const introExpected = `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
			 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd"><node><node name="foo"><node name="bar"><interface name="foo"><method name="CallMe"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Peer"><method name="GetMachineId"><arg type="s" direction="out"></arg></method><method name="Ping"></method></interface><node name="call"><interface name="foo"><method name="CallMe"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Peer"><method name="GetMachineId"><arg type="s" direction="out"></arg></method><method name="Ping"></method></interface></node></node></node></node>`

	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	obj2 := root.NewObjectFromTable("/foo/bar", methods)
	if obj2 == nil {
		t.Fatal("unexpected nil")
	}
	err = obj2.ImplementsTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	root.DeleteObject("/")
	outs, err := root.Call(fdtIntrospectable, "Introspect")
	if err != nil {
		t.Fatal(err)
	}
	expectedNode := decodeIntrospection(introExpected)
	gotNode := decodeIntrospection(outs[0].(string))
	if !reflect.DeepEqual(expectedNode, gotNode) {
		t.Fatalf("expected:\n%s\ngot:\n%s", introExpected, outs[0].(string))
	}
}

func TestObjectReceives(t *testing.T) {
	ch := make(chan string)
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": func(ins ...interface{}) {
			ch <- ins[0].(string)
		},
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ReceivesTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	expected := "hello, world"
	sig := &dbus.Signal{
		Body: []interface{}{expected},
	}
	root.DeliverSignal("foo", "CallMe", sig)
	got := <-ch
	if got != expected {
		t.Fatal("expected:", expected, "got:", got)
	}
}

type recvIface interface {
	CallMe(ins ...interface{})
}

func TestObjectReceivesIface(t *testing.T) {
	ch := make(chan string)
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": func(ins ...interface{}) {
			ch <- ins[0].(string)
		},
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.Receives("foo", (*recvIface)(nil),
		func(in string) string {
			return in
		})
	if err != nil {
		t.Fatal(err)
	}
	expected := "hello, world"
	sig := &dbus.Signal{
		Body: []interface{}{expected},
	}
	root.DeliverSignal("foo", "CallMe", sig)
	got := <-ch
	if got != expected {
		t.Fatal("expected:", expected, "got:", got)
	}
}

type recvIfaceNotImplemented interface {
	CallMe(string)
}

func TestObjectReceivesIfaceNotImplemented(t *testing.T) {
	ch := make(chan string)
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": func(ins ...interface{}) {
			ch <- ins[0].(string)
		},
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.Receives("foo", (*recvIfaceNotImplemented)(nil),
		func(in string) string {
			return in
		})
	if err == nil {
		t.Fatal("Object should not implement interface")
	}
}

func TestObjectDeliverUnhandledSignal(t *testing.T) {
	ch := make(chan string)
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": func(ins ...interface{}) {
			ch <- ins[0].(string)
		},
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ReceivesTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	expected := "hello, world"
	sig := &dbus.Signal{
		Body: []interface{}{expected},
	}
	root.DeliverSignal("foo", "CallMe2", sig)
	select {
	case <-ch:
		t.Fatal("expected timeout to occur")
	case <-time.After(time.Second):
	}
}

func TestObjectReceivesInvalid(t *testing.T) {
	ch := make(chan string)
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": func(ins ...interface{}) {
			ch <- ins[0].(string)
		},
	}
	iface_methods := map[string]interface{}{
		"CallMe2": func(ins ...interface{}) {
		},
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ReceivesTable("foo", iface_methods)
	if err == nil {
		t.Fatal("Expected object to not implement the interface")
	}
}

func TestMultipleObjectReceives(t *testing.T) {
	ch := make(chan string)
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": func(ins ...interface{}) {
			ch <- ins[0].(string)
		},
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ReceivesTable("foo", methods)
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
	expected := "hello, world"
	sig := &dbus.Signal{
		Body: []interface{}{expected},
	}
	root.DeliverSignal("foo", "CallMe", sig)
	got := <-ch
	got2 := <-ch
	if got != expected {
		t.Fatal("expected:", expected, "got:", got)
	}
	if got2 != expected {
		t.Fatal("expected:", expected, "got:", got2)
	}
}

func TestMultipleInterfaceReceives(t *testing.T) {
	ch := make(chan string)
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": func(ins ...interface{}) {
			ch <- ins[0].(string)
		},
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ReceivesTable("foo", methods)
	if err != nil {
		t.Fatal(err)
	}
	err = obj.ReceivesTable("bar", methods)
	if err != nil {
		t.Fatal(err)
	}
	expected := "hello, world"
	sig := &dbus.Signal{
		Body: []interface{}{expected},
	}
	root.DeliverSignal("foo", "CallMe", sig)
	root.DeliverSignal("bar", "CallMe", sig)
	got := <-ch
	got2 := <-ch
	if got != expected {
		t.Fatal("expected:", expected, "got:", got)
	}
	if got2 != expected {
		t.Fatal("expected:", expected, "got:", got2)
	}
}

func TestObjectDeleteReceiver(t *testing.T) {
	ch := make(chan string)
	root := newObjectFromImpl("", nil, nil, nil)
	methods := map[string]interface{}{
		"CallMe": func(ins ...interface{}) {
			ch <- ins[0].(string)
		},
	}
	obj := root.NewObjectFromTable("/foo/bar/call", methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	err := obj.ReceivesTable("foo", methods)
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

func TestPropertyInterfaceGet(t *testing.T) {
	root := newObjectFromImpl("", nil, nil, nil)
	props := map[string]interface{}{
		"Prop1": new(int),
	}
	obj := root.NewObjectFromTable("/foo/bar/props", props)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	err := obj.ImplementsTable("foo.Props", props)
	if err != nil {
		t.Fatal(err)
	}
	iface, exists := obj.LookupInterface(fdtProperties)
	if !exists {
		t.Fatal("Not Property")
	}
	method, exists := iface.LookupMethod("Get")
	if !exists {
		t.Fatal("export failed")
	}

	outs, err := method.Call("foo.Props", "Prop1")
	if err != nil {
		t.Fatal(err)
	}
	if outs[0].(int) != 0 {
		t.Fatal("expected", 0, "got", outs[0].(int))
	}
}

func TestPropertyInterfaceGetAll(t *testing.T) {
	root := newObjectFromImpl("", nil, nil, nil)
	prop1 := 10
	prop2 := "foo bar"
	props := map[string]interface{}{
		"Prop1": &prop1,
		"Prop2": &prop2,
	}
	obj := root.NewObjectFromTable("/foo/bar/props", props)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	err := obj.ImplementsTable("foo.Props", props)
	if err != nil {
		t.Fatal(err)
	}
	iface, exists := obj.LookupInterface(fdtProperties)
	if !exists {
		t.Fatal("Not Property")
	}
	method, exists := iface.LookupMethod("GetAll")
	if !exists {
		t.Fatal("export failed")
	}

	outs, err := method.Call("foo.Props")
	if err != nil {
		t.Fatal(err)
	}
	got := outs[0].(map[string]interface{})
	got1 := got["Prop1"].(int)
	got2 := got["Prop2"].(string)
	if got1 != prop1 {
		t.Fatal("expected", prop1, "got", got1)
	}
	if got2 != prop2 {
		t.Fatal("expected", prop2, "got", got2)
	}
}

func TestPropertyInterfaceSet(t *testing.T) {
	root := newObjectFromImpl("", nil, nil, nil)
	prop1 := 10
	props := map[string]interface{}{
		"Prop1": &prop1,
	}
	obj := root.NewObjectFromTable("/foo/bar/props", props)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	err := obj.ImplementsTable("foo.Props", props)
	if err != nil {
		t.Fatal(err)
	}
	iface, exists := obj.LookupInterface(fdtProperties)
	if !exists {
		t.Fatal("Not Property")
	}
	method, exists := iface.LookupMethod("Set")
	if !exists {
		t.Fatal("export failed")
	}

	expected := 20
	outs, err := method.Call("foo.Props", "Prop1", expected)
	if err != nil {
		t.Fatal(err)
	}

	method, exists = iface.LookupMethod("Get")
	if !exists {
		t.Fatal("export failed")
	}

	outs, err = method.Call("foo.Props", "Prop1")
	if err != nil {
		t.Fatal(err)
	}
	got := outs[0].(int)
	if got != expected {
		t.Fatal("expected", expected, "got", got)
	}

}

func TestPropertyIntrospect(t *testing.T) {
	introExpected := `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
			 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd"><node><node name="foo"><node name="bar"><node name="props"><interface name="foo.Props"><property name="Prop1" type="i" access="readwrite"></property></interface><interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg type="s" direction="out"></arg></method></interface><interface name="org.freedesktop.DBus.Peer"><method name="GetMachineId"><arg type="s" direction="out"></arg></method><method name="Ping"></method></interface><interface name="org.freedesktop.DBus.Properties"><method name="Get"><arg type="s" direction="in"></arg><arg type="s" direction="in"></arg><arg type="v" direction="out"></arg></method><method name="GetAll"><arg type="s" direction="in"></arg><arg type="a{sv}" direction="out"></arg></method><method name="Set"><arg type="s" direction="in"></arg><arg type="s" direction="in"></arg><arg type="v" direction="in"></arg></method></interface></node></node></node></node>`
	root := newObjectFromImpl("", nil, nil, nil)
	prop1 := int32(10)
	props := map[string]interface{}{
		"Prop1": &prop1,
	}
	obj := root.NewObjectFromTable("/foo/bar/props", props)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	err := obj.ImplementsTable("foo.Props", props)
	if err != nil {
		t.Fatal(err)
	}
	outs, err := root.Call(fdtIntrospectable, "Introspect")
	if err != nil {
		t.Fatal(err)
	}

	expectedNode := decodeIntrospection(introExpected)
	gotNode := decodeIntrospection(outs[0].(string))
	if !reflect.DeepEqual(expectedNode, gotNode) {
		t.Fatalf("expected:\n%s\ngot:\n%s", introExpected, outs[0].(string))
	}
}
