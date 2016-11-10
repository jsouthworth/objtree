package objtree

import (
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	ireflect "github.com/jsouthworth/objtree/internal/reflect"
	"reflect"
)

var (
	sendertype = reflect.TypeOf((*dbus.Sender)(nil)).Elem()
	errtype    = reflect.TypeOf((*error)(nil)).Elem()
)

type Method struct {
	name    string
	impl    *ireflect.Method
	sender  string
	message *dbus.Message
}

func (method *Method) Introspect() introspect.Method {
	getArguments := func(
		num func() int,
		get func(int) reflect.Type,
		typ string,
	) []introspect.Arg {
		var args []introspect.Arg
		for j := 0; j < num(); j++ {
			arg := get(j)
			if typ == "out" && j == num()-1 {
				if arg.Implements(errtype) {
					continue
				}
			}
			if typ == "in" && arg == sendertype {
				// Hide argument from introspection
				continue
			}
			iarg := introspect.Arg{
				"",
				dbus.SignatureOfType(arg).String(),
				typ,
			}
			args = append(args, iarg)
		}
		return args
	}

	intro := introspect.Method{
		Name: method.name,
		Args: make([]introspect.Arg, 0,
			method.NumArguments()+method.NumReturns()),
		Annotations: make([]introspect.Annotation, 0),
	}
	intro.Args = append(intro.Args,
		getArguments(method.NumArguments,
			method.impl.ArgumentType, "in")...)
	intro.Args = append(intro.Args,
		getArguments(method.NumReturns,
			method.impl.ReturnType, "out")...)
	return intro
}

func (method *Method) DecodeArguments(
	conn *dbus.Conn,
	sender string,
	msg *dbus.Message,
	args []interface{},
) ([]interface{}, error) {
	body := msg.Body
	pointers := make([]interface{}, method.NumArguments())
	decode := make([]interface{}, 0, len(body))

	method.sender = sender
	method.message = msg

	for i := 0; i < method.impl.NumArguments(); i++ {
		tp := method.impl.ArgumentType(i)
		val := reflect.New(tp)
		pointers[i] = val.Interface()
		if tp == sendertype {
			val.Elem().SetString(sender)
		} else {
			decode = append(decode, pointers[i])
		}
	}

	if len(decode) != len(body) {
		return nil, dbus.ErrMsgInvalidArg
	}

	if err := dbus.Store(body, decode...); err != nil {
		return nil, dbus.ErrMsgInvalidArg
	}
	// Deref the pointers created by reflect.New above
	for i, ptr := range pointers {
		pointers[i] = reflect.ValueOf(ptr).Elem().Interface()
	}
	return pointers, nil
}

func (method *Method) Call(args ...interface{}) ([]interface{}, error) {
	return method.impl.Call(args...)
}

func (method *Method) NumArguments() int {
	return method.impl.NumArguments()
}

func (method *Method) NumReturns() int {
	return method.impl.NumReturns()
}

func (method *Method) ArgumentValue(position int) interface{} {
	return method.impl.ArgumentValue(position)
}

func (method *Method) ReturnValue(position int) interface{} {
	return method.impl.ReturnValue(position)
}
