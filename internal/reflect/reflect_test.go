package reflect

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type testIface interface {
	CallMe() string
}

type testObj struct {
}

func (o *testObj) CallMe() string {
	return "hello, world"
}

func TestObjectCall(t *testing.T) {
	obj := NewObject(&testObj{})
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if !obj.Implements(NewInterface((*testIface)(nil))) {
		t.Fatal("Object does not implement interface")
	}
	method, exists := obj.LookupMethod("CallMe")
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

func TestObjectNil(t *testing.T) {
	obj := NewObject(nil)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if obj.Implements(NewInterface((*testIface)(nil))) {
		t.Fatal("Object should not implement interface")
	}
}

func TestObjectImplementsObject(t *testing.T) {
	obj := NewObject(&testObj{})
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if !obj.Implements(NewInterface(&testObj{})) {
		t.Fatal("Object does not implement interface")
	}
	method, exists := obj.LookupMethod("CallMe")
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

type testObjWithUnexportedMethod struct {
	testObj
}

func (*testObjWithUnexportedMethod) unexported() {
}

func TestObjectUnexportedMethod(t *testing.T) {
	obj := NewObject(&testObjWithUnexportedMethod{})
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	_, exists := obj.LookupMethod("unexported")
	if exists {
		t.Fatal("unexported method exists")
	}
}

type testIntfWithUnexported interface {
	testIface
	unexported()
}

func TestObjectUnexportedMethodInInterface(t *testing.T) {
	obj := NewObject(&testObjWithUnexportedMethod{})
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if obj.Implements(NewInterface((*testIntfWithUnexported)(nil))) {
		t.Fatal("Object should not implement interface")
	}
}

func TestTableObjectCall(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}

	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if !obj.Implements(NewInterface((*testIface)(nil))) {
		t.Fatal("Object does not implement interface")
	}
	method, exists := obj.LookupMethod("CallMe")
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
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if obj.Implements(NewInterface((*testNonMatchingFunc)(nil))) {
		t.Fatal("Should have failed")
	}
}

func TestTableObjectImplementsNonMatchingFuncWrongName(t *testing.T) {
	methods := map[string]interface{}{
		"Call": interface{}(func(string) string { return "hello, world" }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if obj.Implements(NewInterface((*testNonMatchingFunc)(nil))) {
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
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if obj.Implements(NewInterface((*testTooManyMethods)(nil))) {
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
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if obj.Implements(NewInterface((*testMismatchedTypes)(nil))) {
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
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if obj.Implements(NewInterface((*testTooManyOutputs)(nil))) {
		t.Fatal("Should have failed")
	}
}

func TestTableObjectImplementsMoreThanOneFunction(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe":  interface{}(func() string { return "hello, world" }),
		"CallMe2": interface{}(func() string { return "hello, world2" }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if !obj.Implements(NewInterface((*testIface)(nil))) {
		t.Fatal("Object does not implement interface")
	}
}

func TestTableObjectBogusMethod(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": "foobar",
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if obj.Implements(NewInterface((*testIface)(nil))) {
		t.Fatal("Object should not implement testIface")
	}
}

func TestImplementsTable(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	if !obj.Implements(NewInterface((*testIface)(nil))) {
		t.Fatal("Object does not implement interface")
	}
}

func TestImplementsTableWithTable(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	if !obj.Implements(NewInterfaceFromTable(methods)) {
		t.Fatal("Object does not implement interface")
	}
}

func TestImplementsTableWithTableNonFuncType(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func() string { return "hello, world" }),
		"FooBar": "foo bar",
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	if !obj.Implements(NewInterfaceFromTable(methods)) {
		t.Fatal("Object does not implement interface")
	}
}

func TestObjectImplementsNonMatchingInputTypes(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func(in string) string { return in }),
	}
	iface := map[string]interface{}{
		"CallMe": interface{}(func(in int) string { return "" }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	if obj.Implements(NewInterfaceFromTable(iface)) {
		t.Fatal("Object should not implement interface")
	}
}

func TestObjectImplementsNil(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	if obj.Implements(nil) {
		t.Fatal("Object should not implement interface")
	}
}

type testIfaceWithArgs interface {
	CallMe(in string) string
}

func TestObjectCallWithArgs(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if !obj.Implements(NewInterface((*testIfaceWithArgs)(nil))) {
		t.Fatal("Object does not implement interface")
	}
	method, exists := obj.LookupMethod("CallMe")
	if !exists {
		t.Fatal("export failed")
	}

	outs, err := method.Call("hello, world")
	if err != nil {
		t.Fatal(err)
	}
	if outs[0].(string) != "hello, world" {
		t.Fatal("didn't get expected output")
	}
}

func TestObjectCallWithArgsAndError(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": func(in string) (string, error) { return in, nil },
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if !obj.Implements(NewInterfaceFromTable(methods)) {
		t.Fatal("Object does not implement interface")
	}
	method, exists := obj.LookupMethod("CallMe")
	if !exists {
		t.Fatal("export failed")
	}

	outs, err := method.Call("hello, world")
	if err != nil {
		t.Fatal(err)
	}
	if outs[0].(string) != "hello, world" {
		t.Fatal("didn't get expected output")
	}
}

func TestObjectCallWithArgsAndErrorOutput(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": func(in string) (string, error) {
			return in, errors.New("dead")
		},
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if !obj.Implements(NewInterfaceFromTable(methods)) {
		t.Fatal("Object does not implement interface")
	}
	method, exists := obj.LookupMethod("CallMe")
	if !exists {
		t.Fatal("export failed")
	}

	_, err := method.Call("hello, world")
	if err == nil {
		t.Fatal("error should have occured")
	}
}

func TestObjectCallDirectMethod(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	outs, err := obj.Call("CallMe", "hello, world")
	if err != nil {
		t.Fatal(err)
	}
	if outs[0].(string) != "hello, world" {
		t.Fatal("didn't get expected output")
	}
}

func TestObjectCallDirectUnknownMethod(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	_, err := obj.Call("CallMe2", "hello, world")
	if err == nil {
		t.Fatal("Call should have failed, bogus method name")
	}

}

func TestObjectAsInterface(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	iface, err := obj.AsInterface(NewInterfaceFromTable(methods))
	if err != nil {
		t.Fatal(err)
	}
	if iface.impl != obj {
		t.Fatal("Interface implementation should have been object")
	}
}

func TestObjectAsInterfaceBogusMethod(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe": interface{}(func(in string) string { return in }),
	}
	iface_methods := map[string]interface{}{
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	_, err := obj.AsInterface(NewInterfaceFromTable(iface_methods))
	if err == nil {
		t.Fatal("AsInterface should have failed, object does not implement interface")
	}
}

func TestObjectHasCorrectMethods(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe":  interface{}(func(in string) string { return in }),
		"CallMe2": interface{}(func() string { return "CallMe2" }),
		"CallMe3": interface{}(func(in int) int { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}
	obj_methods := obj.Methods()
	for k, _ := range methods {
		if _, exists := obj_methods[k]; !exists {
			t.Fatal("Mismatched method", k)
		}
	}

	for k, _ := range obj_methods {
		if _, exists := methods[k]; !exists {
			t.Fatal("Mismatched method", k)
		}
	}

}

func TestInterfaceImplementsSubInterface(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe":  interface{}(func(in string) string { return in }),
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	iface_methods := map[string]interface{}{
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	iface1_type := NewInterfaceFromTable(methods)
	iface2_type := NewInterfaceFromTable(iface_methods)
	iface1, err := obj.AsInterface(iface1_type)
	if err != nil {
		t.Fatal(err)
	}
	if !iface1.Implements(iface2_type) {
		t.Fatal("Interface should implement iface2")
	}
}

func TestInterfaceImplementsNoMatch(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe":  interface{}(func(in string) string { return in }),
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	iface_methods := map[string]interface{}{
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	iface1_type := NewInterfaceFromTable(methods)
	iface2_type := NewInterfaceFromTable(iface_methods)
	iface2, err := obj.AsInterface(iface2_type)
	if err != nil {
		t.Fatal(err)
	}
	if iface2.Implements(iface1_type) {
		t.Fatal("Interface should not implement iface1")
	}
}

func TestInterfaceImplementsNil(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe":  interface{}(func(in string) string { return in }),
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	iface1_type := NewInterfaceFromTable(methods)
	iface1, err := obj.AsInterface(iface1_type)
	if err != nil {
		t.Fatal(err)
	}
	if iface1.Implements(nil) {
		t.Fatal("Interface should not implement nil")
	}
}

func TestInterfaceAsInterfaceImplements(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe":  interface{}(func(in string) string { return in }),
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	iface_methods := map[string]interface{}{
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	iface1_type := NewInterfaceFromTable(methods)
	iface2_type := NewInterfaceFromTable(iface_methods)
	iface1, err := obj.AsInterface(iface1_type)
	if err != nil {
		t.Fatal(err)
	}
	_, err = iface1.AsInterface(iface2_type)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInterfaceAsInterfaceNoMatch(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe":  interface{}(func(in string) string { return in }),
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	iface_methods := map[string]interface{}{
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	iface1_type := NewInterfaceFromTable(methods)
	iface2_type := NewInterfaceFromTable(iface_methods)
	iface2, err := obj.AsInterface(iface2_type)
	if err != nil {
		t.Fatal(err)
	}
	_, err = iface2.AsInterface(iface1_type)
	if err == nil {
		t.Fatal("Interface should not implement iface1")
	}
}

func TestInterfaceLookupMethod(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe":  interface{}(func(in string) string { return in }),
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	iface_methods := map[string]interface{}{
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	iface1_type := NewInterfaceFromTable(iface_methods)
	iface1, err := obj.AsInterface(iface1_type)
	if err != nil {
		t.Fatal(err)
	}
	_, exists := iface1.LookupMethod("CallMe2")
	if !exists {
		t.Fatal("Method should have existed")
	}
}

func TestInterfaceLookupMethodNoMatch(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe":  interface{}(func(in string) string { return in }),
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	iface_methods := map[string]interface{}{
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	iface1_type := NewInterfaceFromTable(iface_methods)
	iface1, err := obj.AsInterface(iface1_type)
	if err != nil {
		t.Fatal(err)
	}
	_, exists := iface1.LookupMethod("CallMe")
	if exists {
		t.Fatal("Method should not have existed")
	}
}

func TestInterfaceCall(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe":  interface{}(func(in string) string { return in }),
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	iface_methods := map[string]interface{}{
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	iface1_type := NewInterfaceFromTable(iface_methods)
	iface1, err := obj.AsInterface(iface1_type)
	if err != nil {
		t.Fatal(err)
	}
	outs, err := iface1.Call("CallMe2", "hello, world")
	if err != nil {
		t.Fatal(err)
	}
	if outs[0].(string) != "hello, world" {
		t.Fatal("didn't get expected output")
	}
}

func TestInterfaceCallNoMatch(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe":  interface{}(func(in string) string { return in }),
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	iface_methods := map[string]interface{}{
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	iface1_type := NewInterfaceFromTable(iface_methods)
	iface1, err := obj.AsInterface(iface1_type)
	if err != nil {
		t.Fatal(err)
	}
	_, err = iface1.Call("CallMe", "hello, world")
	if err == nil {
		t.Fatal("Method call should have failed")
	}
}

func TestInterfaceMethods(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe":  interface{}(func(in string) string { return in }),
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	iface_methods := map[string]interface{}{
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	iface1_type := NewInterfaceFromTable(iface_methods)
	iface1, err := obj.AsInterface(iface1_type)
	if err != nil {
		t.Fatal(err)
	}

	iface_methods_ret := iface1.Methods()
	for k, _ := range iface_methods {
		if _, exists := iface_methods_ret[k]; !exists {
			t.Fatal("Mismatched method", k)
		}
	}

	for k, _ := range iface_methods_ret {
		if _, exists := iface_methods[k]; !exists {
			t.Fatal("Mismatched method", k)
		}
	}
}

func TestMethodValue(t *testing.T) {
	methods := map[string]interface{}{
		"CallMe":  interface{}(func(in string) string { return in }),
		"CallMe2": interface{}(func(in string) string { return in }),
	}
	obj := NewObjectFromTable(methods)
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	method, _ := obj.LookupMethod("CallMe")
	got := fmt.Sprintf("%T", method.Value().Interface())
	expected := "func(string) string"
	if got != expected {
		t.Fatal("expected:", expected, "got:", got)
	}
}

func TestObjectMapNames(t *testing.T) {
	iface_methods := map[string]interface{}{
		"call-me": interface{}(func() string { return "hello, world" }),
	}
	obj := NewObjectMapNames(&testObj{},
		func(in string) string {
			if in == "CallMe" {
				return "call-me"
			}
			return in
		})
	if obj == nil {
		t.Fatal("unexpected nil")
	}

	if !obj.Implements(NewInterfaceFromTable(iface_methods)) {
		t.Fatal("Object does not implement interface")
	}
	method, exists := obj.LookupMethod("call-me")
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

func TestNewMethod(t *testing.T) {
	method, err := NewMethod(func(in string) string { return in })
	if err != nil {
		t.Fatal(err)
	}
	expected := "hi!"
	outs, err := method.Call(expected)
	if err != nil {
		t.Fatal(err)
	}
	if outs[0].(string) != expected {
		t.Fatal("didn't get expected output")
	}
}

func TestNewMethodNonFunc(t *testing.T) {
	_, err := NewMethod(8)
	if err == nil {
		t.Fatal("NewMethod should have failed due to non function")
	}
}

func TestMethodNumArguments(t *testing.T) {
	method, err := NewMethod(func(in string) string { return in })
	if err != nil {
		t.Fatal(err)
	}

	expected := 1
	got := method.NumArguments()
	if got != expected {
		t.Fatal("NumArguments failed expected:", expected, "got:", got)
	}
}

func TestMethodNumReturns(t *testing.T) {
	method, err := NewMethod(func(in string) string { return in })
	if err != nil {
		t.Fatal(err)
	}

	expected := 1
	got := method.NumReturns()
	if got != expected {
		t.Fatal("NumReturns failed expected:", expected, "got:", got)
	}
}

func TestMethodArgumentValue(t *testing.T) {
	method, err := NewMethod(func(in string) string { return in })
	if err != nil {
		t.Fatal(err)
	}

	_ = method.ArgumentValue(0).(string)
}

func TestMethodArgumentValueOutOfRange(t *testing.T) {
	method, err := NewMethod(func(in string) string { return in })
	if err != nil {
		t.Fatal(err)
	}

	if method.ArgumentValue(1) != nil {
		t.Fatal("ArgumentValue out of range should have been nil")
	}
}

func TestMethodArgumentType(t *testing.T) {
	method, err := NewMethod(func(in string) string { return in })
	if err != nil {
		t.Fatal(err)
	}

	if method.ArgumentType(0) != reflect.TypeOf("foo") {
		t.Fatal("ArgumentType should have been string")
	}
}

func TestMethodArgumentTypeOutOfRange(t *testing.T) {
	method, err := NewMethod(func(in string) string { return in })
	if err != nil {
		t.Fatal(err)
	}

	if method.ArgumentType(1) != nil {
		t.Fatal("ArgumentType out of range should have been nil")
	}
}

func TestMethodReturnValue(t *testing.T) {
	method, err := NewMethod(func(in string) string { return in })
	if err != nil {
		t.Fatal(err)
	}

	_ = method.ReturnValue(0).(string)
}

func TestMethodReturnValueOutOfRange(t *testing.T) {
	method, err := NewMethod(func(in string) string { return in })
	if err != nil {
		t.Fatal(err)
	}

	if method.ReturnValue(1) != nil {
		t.Fatal("ArgumentValue out of range should have been nil")
	}
}

func TestMethodReturnType(t *testing.T) {
	method, err := NewMethod(func(in string) string { return in })
	if err != nil {
		t.Fatal(err)
	}

	if method.ReturnType(0) != reflect.TypeOf("foo") {
		t.Fatal("ArgumentType should have been string")
	}
}

func TestMethodReturnTypeOutOfRange(t *testing.T) {
	method, err := NewMethod(func(in string) string { return in })
	if err != nil {
		t.Fatal(err)
	}

	if method.ReturnType(1) != nil {
		t.Fatal("ArgumentType out of range should have been nil")
	}
}

func TestNewInterfaceMapNames(t *testing.T) {
	mapfn := func(in string) string {
		if in == "CallMe" {
			return "call-me"
		}
		return in
	}
	iface_type := NewInterfaceMapNames((*testIface)(nil), mapfn)
	object := NewObjectMapNames(&testObj{}, mapfn)
	if !object.Implements(iface_type) {
		t.Fatal("should have implemented interface")
	}
}
