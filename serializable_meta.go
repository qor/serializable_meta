package serializable_meta

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

// SerializableMetaInterface is a interface defined methods need for a serializable model
type SerializableMetaInterface interface {
	GetSerializableArgumentResource() *admin.Resource
	GetSerializableArgument(SerializableMetaInterface) interface{}
	GetSerializableArgumentKind() string
	SetSerializableArgumentKind(name string)
	SetSerializableArgumentValue(interface{})
}

// SerializableMeta default struct that implemented SerializableMetaInterface
type SerializableMeta struct {
	Kind  string
	Value serializableArgument `sql:"size:65532"`
}

type serializableArgument struct {
	SerializedValue string
	OriginalValue   interface{}
}

func (sa *serializableArgument) Scan(data interface{}) (err error) {
	switch values := data.(type) {
	case []byte:
		sa.SerializedValue = string(values)
	case string:
		sa.SerializedValue = values
	default:
		err = errors.New("unsupported driver -> Scan pair for MediaLibrary")
	}
	return
}

func (sa serializableArgument) Value() (driver.Value, error) {
	if sa.OriginalValue != nil {
		result, err := json.Marshal(sa.OriginalValue)
		return string(result), err
	}
	return sa.SerializedValue, nil
}

// GetSerializableArgumentKind get serializable argument kind
func (serialize SerializableMeta) GetSerializableArgumentKind() string {
	return serialize.Kind
}

// SetSerializableArgumentKind set serializable argument kind
func (serialize *SerializableMeta) SetSerializableArgumentKind(name string) {
	serialize.Kind = name
}

// GetSerializableArgument get serializable argument
func (serialize *SerializableMeta) GetSerializableArgument(serializableMetaInterface SerializableMetaInterface) interface{} {
	if serialize.Value.OriginalValue != nil {
		return serialize.Value.OriginalValue
	}

	if res := serializableMetaInterface.GetSerializableArgumentResource(); res != nil {
		value := res.NewStruct()
		json.Unmarshal([]byte(serialize.Value.SerializedValue), value)
		return value
	}
	return nil
}

// SetSerializableArgumentValue set serializable argument value
func (serialize *SerializableMeta) SetSerializableArgumentValue(value interface{}) {
	serialize.Value.OriginalValue = value
}

// ConfigureQorResourceBeforeInitialize configure qor resource for qor admin
func (serialize *SerializableMeta) ConfigureQorResourceBeforeInitialize(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		res.GetAdmin().RegisterViewPath("github.com/qor/serializable_meta/views")

		if _, ok := res.Value.(SerializableMetaInterface); ok {
			if res.GetMeta("Kind") == nil {
				res.Meta(&admin.Meta{
					Name: "Kind",
					Type: "hidden",
					Valuer: func(value interface{}, context *qor.Context) interface{} {
						if context.GetDB().NewScope(value).PrimaryKeyZero() {
							return nil
						}
						return value.(SerializableMetaInterface).GetSerializableArgumentKind()
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
							if serializeArgumentResource := serializeArgument.GetSerializableArgumentResource(); serializeArgumentResource != nil {
								var setMeta func(metaors []resource.Metaor, metaValues []*resource.MetaValue)
								value := serializeArgumentResource.NewStruct()

								setMeta = func(metaors []resource.Metaor, metaValues []*resource.MetaValue) {
									for _, meta := range metaors {
										for _, metaValue := range metaValues {
											if meta.GetName() == metaValue.Name {
												if metaResource, ok := meta.GetResource().(*admin.Resource); ok && metaResource != nil {
													setMeta(metaResource.GetMetas([]string{}), metaValue.MetaValues.Values)
													break
												}

												if setter := meta.GetSetter(); setter != nil {
													setter(value, metaValue, context)
													break
												}
											}
										}
									}
								}

								setMeta(serializeArgumentResource.GetMetas([]string{}), metaValue.MetaValues.Values)
								serializeArgument.SetSerializableArgumentValue(value)
							}
						}
					},
				})
			}

			res.NewAttrs("Kind", "SerializableMeta")
			res.EditAttrs("ID", "Kind", "SerializableMeta")
		}
	}
}
