package objtree

import (
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"github.com/jsouthworth/objtree/internal/reflect"
	"sort"
)

type Interface struct {
	name string
	impl *reflect.Interface
}

func (intf *Interface) lookupMethod(name string) (*Method, bool) {
	method, ok := intf.impl.LookupMethod(name)
	if !ok {
		return nil, false
	}
	// Methods have two mutable fields that are caller specific
	// Make a new method with the immutable fields from the stored
	// method.
	new_method := &Method{
		impl: method,
		name: name,
	}
	return new_method, ok
}

func (intf *Interface) LookupMethod(name string) (dbus.Method, bool) {
	method, ok := intf.lookupMethod(name)
	return method, ok
}

func (intf *Interface) Introspect() introspect.Interface {
	getMethods := func() []introspect.Method {
		methods := intf.impl.Methods()
		out := make([]introspect.Method, 0, len(methods))
		for name, _ := range methods {
			method, _ := intf.lookupMethod(name)
			out = append(out, method.Introspect())
		}
		sort.Sort(methodsByName(out))
		return out
	}

	return introspect.Interface{
		Name:    intf.name,
		Methods: getMethods(),
	}
}

type methodsByName []introspect.Method

func (a methodsByName) Len() int           { return len(a) }
func (a methodsByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a methodsByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
