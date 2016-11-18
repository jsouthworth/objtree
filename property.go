package objtree

import (
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"github.com/jsouthworth/objtree/internal/reflect"
)

type Property struct {
	name string
	impl *reflect.Property
}

func (p *Property) Introspect() introspect.Property {
	return introspect.Property{
		Name:        p.name,
		Type:        dbus.SignatureOf(p.impl.Get()).String(),
		Access:      "readwrite",
		Annotations: make([]introspect.Annotation, 0),
	}
}
