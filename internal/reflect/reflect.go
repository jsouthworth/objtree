package reflect

import (
	"errors"
	"reflect"
	"sync"
)

var errtype = reflect.TypeOf((*error)(nil)).Elem()

type Object struct {
	properties map[string]*Property
	methods    map[string]*Method
}

func NewObject(value interface{}) *Object {
	return newObjectFromTable(
		getMethodsFromReceiver(value),
		getPropertiesFromObject(value),
		func(in string) string { return in })
}

func NewObjectMapNames(value interface{}, mapfn func(string) string) *Object {
	return newObjectFromTable(
		getMethodsFromReceiver(value),
		getPropertiesFromObject(value),
		mapfn)
}

func NewObjectFromTable(table map[string]interface{}) *Object {
	return newObjectFromTable(
		getMethodsFromTable(table),
		getPropertiesFromTable(table),
		func(in string) string { return in })
}

func newObjectFromTable(
	mtable map[string]interface{},
	ptable map[string]interface{},
	mapfn func(string) string,
) *Object {
	obj := &Object{
		methods: mapMethodValueNames(
			toMethodValues(mtable),
			mapfn),
		properties: mapPropertyValueNames(
			toPropertyValues(ptable),
			mapfn),
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

func (o *Object) getPropertyTypes() map[string]reflect.Type {
	out := make(map[string]reflect.Type)
	for k, v := range o.properties {
		out[k] = v.value.Type()
	}
	return out
}

func (o *Object) Implements(iface *InterfaceType) bool {
	if iface == nil {
		return false
	}
	return isSubsetOfMethods(iface.methods, o.getMethodTypes()) &&
		isSubsetOfProperties(iface.properties, o.getPropertyTypes())
}

func (o *Object) LookupMethod(name string) (*Method, bool) {
	method, ok := o.methods[name]
	return method, ok
}

func (o *Object) LookupProperty(name string) (*Property, bool) {
	prop, ok := o.properties[name]
	return prop, ok
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

func (o *Object) Properties() map[string]*Property {
	return o.properties
}

type Interface struct {
	typ  *InterfaceType
	impl *Object
}

func (i *Interface) Implements(iface *InterfaceType) bool {
	if iface == nil {
		return false
	}
	return isSubsetOfMethods(iface.methods, i.typ.methods) &&
		isSubsetOfProperties(iface.properties, i.typ.properties)
}

func (i *Interface) LookupMethod(name string) (*Method, bool) {
	_, ok := i.typ.methods[name]
	if !ok {
		return nil, false
	}
	return i.impl.LookupMethod(name)
}

func (i *Interface) LookupProperty(name string) (*Property, bool) {
	_, ok := i.typ.properties[name]
	if !ok {
		return nil, false
	}
	return i.impl.LookupProperty(name)
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

func (i *Interface) Properties() map[string]*Property {
	out := make(map[string]*Property)
	for k, _ := range i.typ.properties {
		out[k] = i.impl.properties[k]
	}
	return out
}

func (i *Interface) Methods() map[string]*Method {
	out := make(map[string]*Method)
	for k, _ := range i.typ.methods {
		out[k] = i.impl.methods[k]
	}
	return out
}

type InterfaceType struct {
	properties map[string]reflect.Type
	methods    map[string]reflect.Type
}

func NewInterface(obj interface{}) *InterfaceType {
	return newInterface(
		getMethodTypes(obj),
		getPropertyTypes(obj),
		func(in string) string { return in })
}

func NewInterfaceMapNames(
	obj interface{},
	mapfn func(string) string,
) *InterfaceType {
	return newInterface(
		getMethodTypes(obj),
		getPropertyTypes(obj),
		mapfn)
}

func NewInterfaceFromTable(table map[string]interface{}) *InterfaceType {
	return newInterface(
		methodTableToTypes(table),
		propertyTableToTypes(table),
		func(in string) string { return in })
}

func newInterface(
	mtable map[string]reflect.Type,
	ptable map[string]reflect.Type,
	mapfn func(string) string,
) *InterfaceType {
	return &InterfaceType{
		methods:    mapTypeNames(mtable, mapfn),
		properties: mapTypeNames(ptable, mapfn),
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
	method_type := method.value.Type()
	arg_values := interfaceSliceToValueSlice(args)
	ret_values := method.value.Call(arg_values)
	ret := valueSliceToInterfaceSlice(ret_values)
	last := method_type.NumOut() - 1
	if last >= 0 && method_type.Out(last).Implements(errtype) {
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
	method_type := method.value.Type()
	last := method_type.NumOut() - 1
	if last >= 0 && method_type.Out(last).Implements(errtype) {
		return method_type.NumOut() - 1
	}
	return method_type.NumOut()
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

type Property struct {
	value reflect.Value
	mu    sync.RWMutex
}

func NewProperty(value interface{}) *Property {
	prop := &Property{}
	rval := reflect.ValueOf(value)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}
	prop.value = rval
	return prop
}

func (p *Property) Set(value interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if reflect.TypeOf(value) != p.value.Type() {
		return errors.New("Value type does not match Property type")
	}
	p.value.Set(reflect.ValueOf(value))
	return nil
}

func (p *Property) Get() interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.value.Interface()
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

func isSubsetOfProperties(subset, set map[string]reflect.Type) bool {
	if len(subset) > len(set) {
		return false
	}
	for property_name, iface_property_type := range subset {
		property_type, exists := set[property_name]
		if !exists {
			return false
		}
		if iface_property_type != property_type {
			return false
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

func getPropertyTypes(object interface{}) map[string]reflect.Type {
	obj_type, is_iface := resolveType(object)
	if is_iface {
		return nil
	}
	out := make(map[string]reflect.Type)
	obj := reflect.ValueOf(object)

	if obj.Kind() == reflect.Ptr {
		obj = obj.Elem()
		obj_type = obj_type.Elem()
	}

	if obj.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < obj_type.NumField(); i++ {
		field := obj.Field(i)
		fieldType := obj_type.Field(i)
		out[fieldType.Name] = field.Type()
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

func mapPropertyValueNames(
	table map[string]*Property,
	mapfn func(string) string,
) map[string]*Property {
	out := make(map[string]*Property)
	for k, v := range table {
		out[mapfn(k)] = v
	}
	return out
}

func mapTypeNames(
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

func toPropertyValues(table map[string]interface{}) map[string]*Property {
	out := make(map[string]*Property)
	for k, v := range table {
		out[k] = NewProperty(v)
	}
	return out
}

func getMethodsFromTable(table map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range table {
		if reflect.ValueOf(v).Kind() != reflect.Func {
			continue
		}
		out[k] = v
	}
	return out
}

func getPropertiesFromTable(table map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range table {
		rval := reflect.ValueOf(v)
		if rval.Kind() == reflect.Func {
			continue
		}
		if rval.Kind() != reflect.Ptr {
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

func propertyTableToTypes(table map[string]interface{}) map[string]reflect.Type {
	types := make(map[string]reflect.Type)
	for name, field := range table {
		if reflect.ValueOf(field).Kind() != reflect.Ptr {
			continue
		}
		types[name] = reflect.TypeOf(field).Elem()
	}
	return types
}

func getMethodsFromReceiver(receiver interface{}) map[string]interface{} {
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

func getPropertiesFromObject(object interface{}) map[string]interface{} {
	if object == nil {
		return nil
	}
	out := make(map[string]interface{})

	rval := reflect.ValueOf(object)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}
	if rval.Kind() != reflect.Struct {
		return nil
	}

	objType := rval.Type()
	for i := 0; i < rval.NumField(); i++ {
		fieldType := objType.Field(i)
		if fieldType.PkgPath != "" {
			continue //skip private fields
		}
		field := rval.Field(i)
		out[fieldType.Name] = field.Addr().Interface()
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
