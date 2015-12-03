package serializable_meta

import (
	"encoding/json"

	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
)

type SerializableMetaInterface interface {
	GetSerializableArgumentResource() *admin.Resource
	GetSerializableArgument(SerializableMetaInterface) interface{}
	GetSerializableArgumentKind() string
	SetSerializableArgumentValue(interface{})
}

type SerializableArgument struct {
	Kind  string
	Value string `sql:"size:65532"`
}

func (serialize SerializableArgument) GetSerializableArgumentKind() string {
	return serialize.Kind
}

func (serialize *SerializableArgument) GetSerializableArgument(serializableMetaInterface SerializableMetaInterface) interface{} {
	if res := serializableMetaInterface.GetSerializableArgumentResource(); res != nil {
		value := res.NewStruct()
		json.Unmarshal([]byte(serialize.Value), value)
		return value
	}
	return nil
}

func (serialize *SerializableArgument) SetSerializableArgumentValue(value interface{}) {
	if bytes, err := json.Marshal(value); err == nil {
		serialize.Value = string(bytes)
	}
}

func (serialize *SerializableArgument) ConfigureQorResourceBeforeInitialize(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		if _, ok := res.Value.(SerializableMetaInterface); ok {
			if res.GetMeta("Kind") == nil {
				res.Meta(&admin.Meta{
					Name: "Kind",
					Type: "hidden",
					Valuer: func(value interface{}, context *qor.Context) interface{} {
						if context.GetDB().NewScope(value).PrimaryKeyZero() {
							return nil
						} else {
							return value.(SerializableMetaInterface).GetSerializableArgumentKind()
						}
					},
				})
			}

			if res.GetMeta("SerializableArgument") == nil {
				res.Meta(&admin.Meta{
					Name: "SerializableArgument",
					Type: "serialize_argument",
					Valuer: func(value interface{}, context *qor.Context) interface{} {
						if serializeArgument, ok := value.(SerializableMetaInterface); ok {
							return struct {
								Value    interface{}
								Resource *admin.Resource
							}{
								Value:    serializeArgument.GetSerializableArgument(serializeArgument),
								Resource: serializeArgument.GetSerializableArgumentResource(),
							}
						}
						return nil
					},
					Setter: func(result interface{}, metaValue *resource.MetaValue, context *qor.Context) {
						if serializeArgument, ok := result.(SerializableMetaInterface); ok {
							serializeArgumentResource := serializeArgument.GetSerializableArgumentResource()
							value := serializeArgumentResource.NewStruct()

							for _, meta := range serializeArgumentResource.GetMetas([]string{}) {
								if metaValue := metaValue.MetaValues.Get(meta.GetName()); metaValue != nil {
									if setter := meta.GetSetter(); setter != nil {
										setter(value, metaValue, context)
									}
								}
							}

							serializeArgument.SetSerializableArgumentValue(value)
						}
					},
				})
			}

			res.NewAttrs("Kind", "SerializableArgument")
			res.EditAttrs("ID", "Kind", "SerializableArgument")
		}
	}
}
