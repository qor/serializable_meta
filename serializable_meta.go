package serializable_meta

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"os"
	"path"
	"strings"

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

type SerializableMeta struct {
	Kind  string
	Value SerializableArgument `sql:"size:65532"`
}

type SerializableArgument struct {
	SerializdValue string
	originalValue  interface{}
}

func (sa *SerializableArgument) Scan(data interface{}) (err error) {
	switch values := data.(type) {
	case []byte:
		sa.SerializdValue = string(values)
	case string:
		sa.SerializdValue = values
	default:
		err = errors.New("unsupported driver -> Scan pair for MediaLibrary")
	}
	return
}

func (sa SerializableArgument) Value() (driver.Value, error) {
	if sa.originalValue != nil {
		result, err := json.Marshal(sa.originalValue)
		return string(result), err
	}
	return sa.SerializdValue, nil
}

func (serialize SerializableMeta) GetSerializableArgumentKind() string {
	return serialize.Kind
}

func (serialize *SerializableMeta) GetSerializableArgument(serializableMetaInterface SerializableMetaInterface) interface{} {
	if serialize.Value.originalValue != nil {
		return serialize.Value.originalValue
	}

	if res := serializableMetaInterface.GetSerializableArgumentResource(); res != nil {
		value := res.NewStruct()
		json.Unmarshal([]byte(serialize.Value.SerializdValue), value)
		return value
	}
	return nil
}

func (serialize *SerializableMeta) SetSerializableArgumentValue(value interface{}) {
	serialize.Value.originalValue = value
}

func (serialize *SerializableMeta) ConfigureQorResourceBeforeInitialize(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
			admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/serializable_meta/views"))
		}

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

			if res.GetMeta("SerializableMeta") == nil {
				res.Meta(&admin.Meta{
					Name: "SerializableMeta",
					Type: "serializable_meta",
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
								for _, metaValue := range metaValue.MetaValues.Values {
									if meta.GetName() == metaValue.Name {
										if setter := meta.GetSetter(); setter != nil {
											setter(value, metaValue, context)
										}
									}
								}
							}

							serializeArgument.SetSerializableArgumentValue(value)
						}
					},
				})
			}

			res.NewAttrs("Kind", "SerializableMeta")
			res.EditAttrs("ID", "Kind", "SerializableMeta")
		}
	}
}
