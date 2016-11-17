package objtree

import (
	"encoding/xml"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"github.com/jsouthworth/objtree/internal/reflect"
	"sort"
	"strings"
)

type Object struct {
	name       string
	impl       *reflect.Object
	interfaces multiWriterValue
	listeners  multiWriterValue
	objects    multiWriterValue
	bus        *BusManager
	parent     *Object
}

func newObjectFromTable(
	name string,
	table map[string]interface{},
	parent *Object,
	bus *BusManager,
) *Object {
	return newObjectFromImpl(name, reflect.NewObjectFromTable(table),
		parent, bus)
}

func newObjectFromImpl(
	name string,
	impl *reflect.Object,
	parent *Object,
	bus *BusManager,
) *Object {
	obj := &Object{
		name:   name,
		impl:   impl,
		bus:    bus,
		parent: parent,
	}
	obj.interfaces.value.Store(make(map[string]*Interface))
	obj.listeners.value.Store(make(map[string]*Interface))
	obj.objects.value.Store(make(map[string]*Object))
	obj.addInterface(fdtIntrospectable, newIntrospection(obj))
	obj.addInterface(fdtPeer, newPeer(obj))
	return obj
}

func (o *Object) removeListeners() {
	o.listeners.Update(func(value interface{}) interface{} {
		for dbusIfaceName, intf := range value.(map[string]*Interface) {
			for sigName, _ := range intf.impl.Methods() {
				if o.bus == nil {
					continue
				}
				o.bus.state.RemoveMatchSignal(o.bus.conn,
					dbusIfaceName, sigName)

			}
		}
		return make(map[string]*Interface)
	})
}

func (o *Object) getObjects() map[string]*Object {
	return o.objects.Load().(map[string]*Object)
}

func (o *Object) getInterfaces() map[string]*Interface {
	return o.interfaces.Load().(map[string]*Interface)
}

func (o *Object) getListeners() map[string]*Interface {
	return o.listeners.Load().(map[string]*Interface)
}

func (o *Object) newObject(path []string, impl *reflect.Object) *Object {
	name := path[0]
	switch len(path) {
	case 1:
		obj := newObjectFromImpl(name, impl, o, o.bus)
		o.addObject(name, obj)
		return obj
	default:
		obj, ok := o.LookupObject(name)
		if !ok {
			//placeholder object for introspection
			obj = newObjectFromImpl(name, nil, o, o.bus)
			o.addObject(name, obj)
		}
		return obj.newObject(path[1:], impl)
	}
}

func pathToStringSlice(path dbus.ObjectPath) []string {
	ps := strings.Split(string(path), "/")
	if ps[0] == "" {
		ps = ps[1:]
	}
	return ps
}

func (o *Object) NewObject(path dbus.ObjectPath, val interface{}) *Object {
	if string(path) == "/" {
		return o
	}
	return o.newObject(pathToStringSlice(path),
		reflect.NewObject(val))
}

func (o *Object) NewObjectFromTable(
	path dbus.ObjectPath,
	table map[string]interface{},
) *Object {
	if string(path) == "/" {
		return o
	}
	return o.newObject(pathToStringSlice(path),
		reflect.NewObjectFromTable(table))
}

func (o *Object) NewObjectMap(
	path dbus.ObjectPath,
	val interface{},
	mapfn func(string) string,
) *Object {
	if string(path) == "/" {
		return o
	}
	return o.newObject(pathToStringSlice(path),
		reflect.NewObjectMapNames(val, mapfn))
}

func (o *Object) hasActions() bool {
	return o.impl != nil
}

func (o *Object) hasChildren() bool {
	return len(o.getObjects()) > 0
}

func (o *Object) rmChildObject(name string) {
	o.objects.Update(func(value interface{}) interface{} {
		objects := make(map[string]*Object)
		for child, obj := range o.getObjects() {
			objects[child] = obj
		}
		if obj, ok := objects[name]; ok {
			obj.removeListeners()
			// if there are children replace with placeholder
			if obj.hasChildren() {
				object := newObjectFromImpl(name, nil, o, o.bus)
				object.objects = obj.objects
				objects[name] = object
			} else {
				delete(objects, name)
			}
		}
		return objects
	})
	if !o.hasActions() && o.parent != nil {
		o.parent.rmChildObject(o.name)
	}
}

func (o *Object) DeleteObject(path dbus.ObjectPath) {
	if string(path) == "/" {
		return
	}
	o.delObject(pathToStringSlice(path))
}

func (o *Object) delObject(path []string) {
	name := path[0]
	switch len(path) {
	case 1:
		if _, ok := o.LookupObject(name); ok {
			o.rmChildObject(name)
		}
	default:
		if child, ok := o.LookupObject(name); ok {
			child.delObject(path[1:])
		}
	}
}

func (o *Object) lookupObjectPath(path []string) (*Object, bool) {
	switch len(path) {
	case 1:
		return o.LookupObject(path[0])
	default:
		obj, ok := o.LookupObject(path[0])
		if !ok {
			return nil, false
		}
		return obj.lookupObjectPath(path[1:])
	}
}

func (o *Object) LookupObject(name string) (*Object, bool) {
	obj, ok := o.getObjects()[name]
	return obj, ok
}

func (o *Object) LookupInterface(name string) (dbus.Interface, bool) {
	iface, ok := o.getInterfaces()[name]
	return iface, ok
}

func (o *Object) addInterface(name string, iface *Interface) {
	o.interfaces.Update(func(value interface{}) interface{} {
		interfaces := make(map[string]*Interface)
		for name, intf := range value.(map[string]*Interface) {
			interfaces[name] = intf
		}
		interfaces[name] = iface
		return interfaces
	})
}

func (o *Object) addListener(name string, iface *Interface) {
	o.listeners.Update(func(value interface{}) interface{} {
		listeners := make(map[string]*Interface)
		for name, intf := range value.(map[string]*Interface) {
			listeners[name] = intf
		}
		listeners[name] = iface
		if o.bus != nil {
			for method_name, _ := range iface.impl.Methods() {
				o.bus.state.AddMatchSignal(o.bus.conn,
					name, method_name)
			}
		}
		return listeners
	})
}

func (o *Object) addObject(name string, object *Object) {
	o.objects.Update(func(value interface{}) interface{} {
		objects := make(map[string]*Object)
		for name, obj := range value.(map[string]*Object) {
			objects[name] = obj
		}
		if obj, ok := objects[name]; ok {
			//there may be child objects of the object that is being
			//replaced; keep them
			object.objects = obj.objects
		}
		objects[name] = object
		return objects
	})
}

func (o *Object) Implements(name string, obj interface{}) error {
	return o.ImplementsMap(name, obj,
		func(in string) string {
			return in
		})
}

func (o *Object) ImplementsMap(
	name string,
	obj interface{},
	mapfn func(string) string,
) error {
	iface, err := o.impl.AsInterface(
		reflect.NewInterfaceMapNames(obj, mapfn))
	if err != nil {
		return err
	}
	return o.implementsIface(name, iface)
}

func (o *Object) ImplementsTable(
	name string,
	table map[string]interface{},
) error {
	iface, err := o.impl.AsInterface(
		reflect.NewInterfaceFromTable(table))
	if err != nil {
		return err
	}
	return o.implementsIface(name, iface)

}
func (o *Object) implementsIface(
	name string,
	iface *reflect.Interface,
) error {
	intf := &Interface{
		name: name,
		impl: iface,
	}

	o.addInterface(name, intf)
	return nil
}

// Call for each D-Bus interface to receive signals from
func (o *Object) Receives(
	dbusIfaceName string,
	obj interface{},
	mapfn func(string) string,
) error {
	iface, err := o.impl.AsInterface(
		reflect.NewInterfaceMapNames(obj, mapfn))
	if err != nil {
		return err
	}
	return o.receivesIface(dbusIfaceName, iface)
}

func (o *Object) ReceivesTable(
	dbusIfaceName string,
	table map[string]interface{},
) error {
	iface, err := o.impl.AsInterface(
		reflect.NewInterfaceFromTable(table))
	if err != nil {
		return err
	}
	return o.receivesIface(dbusIfaceName, iface)
}

func (o *Object) receivesIface(
	dbusIfaceName string,
	iface *reflect.Interface,
) error {
	intf := &Interface{
		name: dbusIfaceName,
		impl: iface,
	}

	o.addListener(dbusIfaceName, intf)
	return nil
}

// Deliver the signal to this object's listeners and all child objects
func (o *Object) DeliverSignal(iface, member string, signal *dbus.Signal) {
	defer func() {
		objects := o.getObjects()
		for _, obj := range objects {
			obj.DeliverSignal(iface, member, signal)
		}
	}()

	listeners := o.getListeners()
	intf, ok := listeners[iface]
	if !ok {
		return
	}
	method, ok := intf.LookupMethod(member)
	if !ok {
		return
	}
	go func() {
		method.Call(signal.Body...)
	}()
}

func (o *Object) Call(
	ifaceName, method string,
	args ...interface{},
) ([]interface{}, error) {
	iface, exists := o.LookupInterface(ifaceName)
	if !exists {
		return nil, dbus.ErrMsgUnknownInterface
	}

	m, exists := iface.LookupMethod(method)
	if !exists {
		return nil, dbus.ErrMsgUnknownMethod
	}

	return m.Call(args...)
}

func (o *Object) Introspect() introspect.Node {
	getChildren := func() []introspect.Node {
		children := o.getObjects()
		out := make([]introspect.Node, 0, len(children))
		for _, child := range children {
			intro := child.Introspect()
			out = append(out, intro)
		}
		sort.Sort(nodesByName(out))
		return out
	}

	getInterfaces := func() []introspect.Interface {
		if !o.hasActions() {
			return nil
		}
		ifaces := o.getInterfaces()
		out := make([]introspect.Interface, 0, len(ifaces))
		for _, iface := range ifaces {
			intro := iface.Introspect()
			out = append(out, intro)
		}
		sort.Sort(interfacesByName(out))
		return out
	}

	node := introspect.Node{
		Name:       o.name,
		Interfaces: getInterfaces(),
		Children:   getChildren(),
	}
	return node
}

func newIntrospection(o *Object) *Interface {
	intro := func() string {
		n := o.Introspect()
		n.Name = "" // Make it work with busctl.
		//Busctl doesn't treat the optional
		//name attribute of the root node correctly.
		b, _ := xml.Marshal(n)
		declaration := strings.TrimSpace(
			introspect.IntrospectDeclarationString)
		return declaration + string(b)
	}

	methods := map[string]interface{}{
		"Introspect": intro,
	}
	impl, _ := reflect.NewObjectFromTable(methods).
		AsInterface(reflect.NewInterfaceFromTable(methods))
	return &Interface{
		name: fdtIntrospectable,
		impl: impl,
	}
}

func newPeer(o *Object) *Interface {
	// These are actually implemented by godbus.
	// This gives proper introspection for the interface.
	getMachineId := func() string {
		return ""
	}
	ping := func() {}
	methods := map[string]interface{}{
		"Ping":         ping,
		"GetMachineId": getMachineId,
	}
	impl, _ := reflect.NewObjectFromTable(methods).
		AsInterface(reflect.NewInterfaceFromTable(methods))
	return &Interface{
		name: fdtPeer,
		impl: impl,
	}
}

type interfacesByName []introspect.Interface

func (a interfacesByName) Len() int           { return len(a) }
func (a interfacesByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a interfacesByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

type nodesByName []introspect.Node

func (a nodesByName) Len() int           { return len(a) }
func (a nodesByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a nodesByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
