package objtree

import (
	"github.com/godbus/dbus"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	fdtDBusName       = "org.freedesktop.DBus"
	fdtAddMatch       = fdtDBusName + ".AddMatch"
	fdtRemoveMatch    = fdtDBusName + ".RemoveMatch"
	fdtIntrospectable = fdtDBusName + ".Introspectable"
)

// Acts as a root to the object tree
type BusManager struct {
	*Object
	conn  *dbus.Conn
	state *mgrState
}

func NewAnonymousBusManager(
	busfn func(dbus.Handler, dbus.SignalHandler) (*dbus.Conn, error),
) (*BusManager, error) {
	state := &mgrState{sigref: make(map[string]uint64)}
	handler := &BusManager{
		Object: newObjectFromImpl("", nil, nil, nil),
		state:  state,
	}
	handler.bus = handler
	conn, err := busfn(handler, handler)
	if err != nil {
		return nil, err
	}
	err = conn.Auth(nil)
	if err != nil {
		conn.Close()
		return nil, err
	}
	err = conn.Hello()
	if err != nil {
		conn.Close()
		return nil, err
	}
	handler.conn = conn
	return handler, nil
}

func NewBusManager(
	busfn func(dbus.Handler, dbus.SignalHandler) (*dbus.Conn, error),
	name string,
) (*BusManager, error) {

	handler, err := NewAnonymousBusManager(busfn)
	if err != nil {
		return nil, err
	}

	err = handler.RequestName(name)
	if err != nil {
		handler.conn.Close()
		return nil, err
	}

	return handler, nil
}

func NewSessionBusManager(name string) (*BusManager, error) {
	return NewBusManager(dbus.SessionBusPrivateHandler, name)
}

func NewAnonymousSessionBusManager() (*BusManager, error) {
	return NewAnonymousBusManager(dbus.SessionBusPrivateHandler)
}

func NewSystemBusManager(name string) (*BusManager, error) {
	return NewBusManager(dbus.SystemBusPrivateHandler, name)
}

func NewAnonymousSystemBusManager() (*BusManager, error) {
	return NewAnonymousBusManager(dbus.SystemBusPrivateHandler)
}

func (mgr *BusManager) RequestName(name string) error {
	_, err := mgr.conn.RequestName(name, 0)
	if err != nil {
		return err
	}
	return nil
}

func (mgr *BusManager) LookupObject(path dbus.ObjectPath) (dbus.ServerObject, bool) {
	if string(path) == "/" {
		return mgr, true
	}

	ps := strings.Split(string(path), "/")
	if ps[0] == "" {
		ps = ps[1:]
	}
	return mgr.lookupObjectPath(ps)
}

func (mgr *BusManager) Call(
	path dbus.ObjectPath,
	ifaceName string,
	method string,
	args ...interface{},
) ([]interface{}, error) {
	object, ok := mgr.LookupObject(path)
	if !ok {
		return nil, dbus.ErrMsgNoObject
	}
	return object.(*Object).Call(ifaceName, method, args...)
}

func (mgr *BusManager) DeliverSignal(iface, member string, signal *dbus.Signal) {
	objects := mgr.objects.Load().(map[string]*Object)
	for _, obj := range objects {
		obj.DeliverSignal(iface, member, signal)
	}
}

type multiWriterValue struct {
	value   atomic.Value
	writelk sync.Mutex
}

func (value *multiWriterValue) Load() interface{} {
	return value.value.Load()
}

func (value *multiWriterValue) Update(fn func(interface{}) interface{}) {
	value.writelk.Lock()
	value.value.Store(fn(value.value.Load()))
	value.writelk.Unlock()
}

type mgrState struct {
	mu     sync.Mutex
	sigref map[string]uint64
}

func mkSignalKey(iface, member string) string {
	return iface + "." + member
}

func (s *mgrState) AddMatchSignal(conn *dbus.Conn, iface, member string) {
	// Only register for signal if not already registered
	s.mu.Lock()
	defer s.mu.Unlock()
	key := mkSignalKey(iface, member)
	if s.sigref[key] == 0 {
		conn.BusObject().Call(fdtAddMatch, 0,
			"type='signal',interface='"+iface+"',member='"+member+"'")
	}
	s.sigref[key] = s.sigref[key] + 1
}

func (s *mgrState) RemoveMatchSignal(conn *dbus.Conn, iface, member string) {
	// Only deregister if this is the last request
	s.mu.Lock()
	defer s.mu.Unlock()
	key := mkSignalKey(iface, member)
	if s.sigref[key] == 0 {
		return
	}
	s.sigref[key] = s.sigref[key] - 1
	if s.sigref[key] == 0 {
		conn.BusObject().Call(fdtRemoveMatch, 0,
			"type='signal',interface='"+iface+"',member='"+member+"'")
	}
}
