package reflect

import (
	"errors"
	"reflect"
)

type Object struct {
	methods map[string]*Method
}

func NewObject(value interface{}) *Object {
	return newObjectFromTable(GetMethodsFromReceiver(value),
		func(in string) string { return in })
}

func NewObjectMapNames(value interface{}, mapfn func(string) string) *Object {
	return newObjectFromTable(GetMethodsFromReceiver(value), mapfn)
}

func NewObjectFromTable(table map[string]interface{}) *Object {
	return newObjectFromTable(filterMethodTable(table),
		func(in string) string { return in })
}

func newObjectFromTable(
	table map[string]interface{},
	mapfn func(string) string,
) *Object {
	obj := &Object{
		methods: mapMethodValueNames(toMethodValues(table), mapfn),
	}
	return obj
}

func (o *Object) getMethodTypes() map[string]reflect.Type {
	out := make(map[string]reflect.Type)
	for k, v := range o.methods {
		out[k] = v.value.Type()
	}
	return out
}

func (o *Object) Implements(iface *InterfaceType) bool {
	if iface == nil {
		return false
	}
	return isSubsetOfMethods(iface.methods, o.getMethodTypes())
}

func (o *Object) LookupMethod(name string) (*Method, bool) {
	method, ok := o.methods[name]
	return method, ok
}

func (o *Object) Call(name string, args ...interface{}) ([]interface{}, error) {
	method, ok := o.LookupMethod(name)
	if !ok {
		return nil, errors.New("Unknown method: " + name)
	}
	return method.Call(args...)
}

func (o *Object) AsInterface(iface *InterfaceType) (*Interface, error) {
	if !o.Implements(iface) {
		return nil, errors.New("Object does not implement interface")
	}
	return &Interface{
		impl: o,
		typ:  iface,
	}, nil
}

func (o *Object) Methods() map[string]*Method {
	return o.methods
}

type Interface struct {
	typ  *InterfaceType
	impl *Object
}

func (i *Interface) Implements(iface *InterfaceType) bool {
	if iface == nil {
		return false
	}
	return isSubsetOfMethods(iface.methods, i.typ.methods)
}

func (i *Interface) LookupMethod(name string) (*Method, bool) {
	_, ok := i.typ.methods[name]
	if !ok {
		return nil, false
	}
	return i.impl.LookupMethod(name)
}

func (i *Interface) Call(name string, args ...interface{}) ([]interface{}, error) {
	method, ok := i.LookupMethod(name)
	if !ok {
		return nil, errors.New("Unknown method: " + name)
	}
	return method.Call(args...)
}

func (i *Interface) AsInterface(iface *InterfaceType) (*Interface, error) {
	if !i.Implements(iface) {
		return nil, errors.New("Object does not implement interface")
	}
	return &Interface{
		impl: i.impl,
		typ:  iface,
	}, nil
}

func (i *Interface) Methods() map[string]*Method {
	out := make(map[string]*Method)
	for k, _ := range i.typ.methods {
		out[k] = i.impl.methods[k]
	}
	return out
}

type InterfaceType struct {
	methods map[string]reflect.Type
}

func NewInterface(obj interface{}) *InterfaceType {
	return newInterface(getMethodTypes(obj),
		func(in string) string { return in })
}

func NewInterfaceMapNames(
	obj interface{},
	mapfn func(string) string,
) *InterfaceType {
	return newInterface(getMethodTypes(obj), mapfn)
}

func NewInterfaceFromTable(table map[string]interface{}) *InterfaceType {
	return newInterface(methodTableToTypes(table),
		func(in string) string { return in })
}

func newInterface(
	table map[string]reflect.Type,
	mapfn func(string) string,
) *InterfaceType {
	return &InterfaceType{
		methods: mapMethodTypeNames(table, mapfn),
	}
}

type Method struct {
	value reflect.Value
}

func NewMethod(method interface{}) (*Method, error) {
	methodValue := reflect.ValueOf(method)
	if methodValue.Kind() != reflect.Func {
		return nil, errors.New("Argument is not a function")
	}
	return &Method{
		value: methodValue,
	}, nil
}

func (method *Method) Call(args ...interface{}) ([]interface{}, error) {
	errtype := reflect.TypeOf((*error)(nil)).Elem()
	method_type := method.value.Type()
	arg_values := interfaceSliceToValueSlice(args)
	ret_values := method.value.Call(arg_values)
	ret := valueSliceToInterfaceSlice(ret_values)
	last := method_type.NumOut() - 1
	if last > 0 && method_type.Out(last).Implements(errtype) {
		// Last parameter is of type error
		if ret[last] != nil {
			return ret[:last], ret[last].(error)
		}
		return ret[:last], nil
	}
	return ret, nil
}

func (method *Method) Value() reflect.Value {
	return method.value
}

func (method *Method) NumArguments() int {
	return method.value.Type().NumIn()
}

func (method *Method) NumReturns() int {
	return method.value.Type().NumOut()
}

func (method *Method) ArgumentValue(position int) interface{} {
	if position >= method.NumArguments() {
		return nil
	}
	return reflect.Zero(method.value.Type().In(position)).Interface()
}

func (method *Method) ArgumentType(position int) reflect.Type {
	if position >= method.NumArguments() {
		return nil
	}
	return method.value.Type().In(position)
}

func (method *Method) ReturnValue(position int) interface{} {
	if position >= method.NumReturns() {
		return nil
	}
	return reflect.Zero(method.value.Type().Out(position)).Interface()
}

func (method *Method) ReturnType(position int) reflect.Type {
	if position >= method.NumReturns() {
		return nil
	}
	return method.value.Type().Out(position)
}

func isSubsetOfMethods(subset, set map[string]reflect.Type) bool {
	if len(subset) > len(set) {
		return false
	}
	for method_name, iface_method_type := range subset {
		method_type, exists := set[method_name]
		if !exists {
			return false
		}
		if iface_method_type.NumIn() != method_type.NumIn() {
			return false
		}
		if iface_method_type.NumOut() != method_type.NumOut() {
			return false
		}
		for j := 0; j < iface_method_type.NumIn(); j++ {
			if iface_method_type.In(j) != method_type.In(j) {
				return false
			}
		}
		for j := 0; j < iface_method_type.NumOut(); j++ {
			if iface_method_type.Out(j) != method_type.Out(j) {
				return false
			}
		}
	}
	return true
}

func getMethodTypes(object interface{}) map[string]reflect.Type {
	obj_type, is_iface := resolveType(object)
	out := make(map[string]reflect.Type)
	if is_iface {
		for i := 0; i < obj_type.NumMethod(); i++ {
			methodType := obj_type.Method(i)
			out[methodType.Name] = methodType.Type
		}
		return out
	}
	obj := reflect.ValueOf(object)
	for i := 0; i < obj_type.NumMethod(); i++ {
		method := obj.Method(i)
		methodType := obj_type.Method(i)
		out[methodType.Name] = method.Type()
	}
	return out
}

func resolveType(obj interface{}) (reflect.Type, bool) {
	obj_typ := reflect.TypeOf(obj)
	if obj_typ.Kind() == reflect.Ptr {
		typ := obj_typ.Elem()
		if typ.Kind() == reflect.Interface {
			return typ, true
		}
	}
	return obj_typ, false
}

func mapMethodValueNames(
	table map[string]*Method,
	mapfn func(string) string,
) map[string]*Method {
	out := make(map[string]*Method)
	for k, v := range table {
		out[mapfn(k)] = v
	}
	return out
}

func mapMethodTypeNames(
	table map[string]reflect.Type,
	mapfn func(string) string,
) map[string]reflect.Type {
	out := make(map[string]reflect.Type)
	for k, v := range table {
		out[mapfn(k)] = v
	}
	return out
}

func toMethodValues(table map[string]interface{}) map[string]*Method {
	out := make(map[string]*Method)
	for k, v := range table {
		out[k] = &Method{reflect.ValueOf(v)}
	}
	return out
}

func filterMethodTable(table map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range table {
		if reflect.ValueOf(v).Kind() != reflect.Func {
			continue
		}
		out[k] = v
	}
	return out
}

func methodTableToTypes(table map[string]interface{}) map[string]reflect.Type {
	types := make(map[string]reflect.Type)
	for name, method := range table {
		if reflect.ValueOf(method).Kind() != reflect.Func {
			continue
		}
		types[name] = reflect.TypeOf(method)
	}
	return types
}

func GetMethodsFromReceiver(receiver interface{}) map[string]interface{} {
	if receiver == nil {
		return nil
	}
	out := make(map[string]interface{})
	rval := reflect.ValueOf(receiver)
	recvType := reflect.TypeOf(receiver)
	for i := 0; i < rval.NumMethod(); i++ {
		method := rval.Method(i)
		methodType := recvType.Method(i)
		if methodType.PkgPath != "" {
			continue //skip private methods
		}
		out[methodType.Name] = method.Interface()
	}
	return out
}

func interfaceSliceToValueSlice(args []interface{}) []reflect.Value {
	out := make([]reflect.Value, len(args))
	for i, v := range args {
		out[i] = reflect.ValueOf(v)
	}
	return out
}

func valueSliceToInterfaceSlice(vals []reflect.Value) []interface{} {
	out := make([]interface{}, len(vals))
	for i, v := range vals {
		out[i] = v.Interface()
	}
	return out
}
