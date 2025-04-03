package meta

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

func APIBody(body interface{}) (*anypb.Any, error) {
	m, ok := body.(proto.Message)
	if !ok {
		// body is not the type/value we want to process
		return nil, nil
	}

	// Deep copy before field redaction so we do not unintentionally remove fields
	// from the original object that were passed by reference
	m = proto.Clone(m)
	return anypb.New(ClearLogDisabledFields(m))
}

func ClearLogDisabledFields(m proto.Message) proto.Message {
	if m == nil {
		return m
	}

	pb := m.ProtoReflect()
	pb.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		// Handle nested types.
		switch t := v.Interface().(type) {
		case protoreflect.Message:
			ClearLogDisabledFields(t.Interface())
		case protoreflect.Map:
			t.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
				if _, ok := v.Interface().(protoreflect.Message); ok {
					ClearLogDisabledFields(v.Message().Interface())
				}
				return true
			})
		case protoreflect.List: // i.e. `repeated`.
			for i := 0; i < t.Len(); i++ {
				if _, ok := t.Get(i).Interface().(protoreflect.Message); ok {
					ClearLogDisabledFields(t.Get(i).Message().Interface())
				}
			}
		}
		return true
	})

	return m
}
