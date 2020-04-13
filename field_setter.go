package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	pgs "github.com/lyft/protoc-gen-star"

	"github.com/mikekonan/protoc-gen-setter/setter"
)

type fieldSetter struct {
	*pgs.ModuleBase
}

func NewFieldSetter() *fieldSetter { return &fieldSetter{ModuleBase: &pgs.ModuleBase{}} }

func (fieldsSetter *fieldSetter) Name() string { return "fieldSetter" }

func (fieldsSetter *fieldSetter) Execute(targets map[string]pgs.File, _ map[string]pgs.Package) []pgs.Artifact {
	for _, target := range targets {
		buf := &bytes.Buffer{}

		hasInclude := false

		pkgName, base := getPackageName(target)

		setterFile := &setterFile{
			Package: pkgName,
			Name:    target.Name().String(),
			All:     target.Descriptor().Options.ProtoReflect().Get(setter.E_AllMessages.TypeDescriptor()).Bool(),
		}

		for _, msg := range target.AllMessages() {
			setterMsg := &setterMessage{
				Name: msg.Name().String(),
				All:  msg.Descriptor().Options.ProtoReflect().Get(setter.E_AllFields.TypeDescriptor()).Bool(),
			}

			setterFile.Messages = append(setterFile.Messages, setterMsg)

			for _, field := range msg.Fields() {
				include := setterFile.All

				if field.Descriptor().Options.ProtoReflect().Get(setter.E_Exclude.TypeDescriptor()).Bool() {
					include = false
				} else if field.Descriptor().Options.ProtoReflect().Get(setter.E_Include.TypeDescriptor()).Bool() {
					include = true
				}

				if include {
					setterField := &setterField{
						Name:    field.Name().UpperCamelCase().String(),
						VarName: field.Name().LowerCamelCase().String(),
						VarType: fieldsSetter.goType(field),
					}

					hasInclude = true
					setterMsg.Fields = append(setterMsg.Fields, setterField)
				}
			}
		}

		if hasInclude {
			if err := setterFile.into(buf); err != nil {
				panic(err)
			}
		}

		fieldsSetter.AddGeneratorFile(
			target.InputPath().SetBase(base).SetExt(".setter.pb.go").String(),
			buf.String(),
		)
	}

	return fieldsSetter.Artifacts()
}

func (fieldsSetter *fieldSetter) goType(field pgs.Field) (typ string) {
	detectType := func(fieldDescriptor descriptor.FieldDescriptorProto_Type) string {
		switch fieldDescriptor {
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
			return "float64"
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			return "float32"
		case descriptor.FieldDescriptorProto_TYPE_INT64:
			return "int64"
		case descriptor.FieldDescriptorProto_TYPE_UINT64:
			return "uint64"
		case descriptor.FieldDescriptorProto_TYPE_INT32:
			return "int32"
		case descriptor.FieldDescriptorProto_TYPE_UINT32:
			return "uint32"
		case descriptor.FieldDescriptorProto_TYPE_FIXED64:
			return "uint64"
		case descriptor.FieldDescriptorProto_TYPE_FIXED32:
			return "uint32"
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			return "bool"
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			return "string"
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			return "[]byte"
		case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
			return "int32"
		case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
			return "int64"
		case descriptor.FieldDescriptorProto_TYPE_SINT32:
			return "int32"
		case descriptor.FieldDescriptorProto_TYPE_SINT64:
			return "int64"
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			typeName := strings.SplitN(field.Descriptor().GetTypeName(), ".", -1)
			return "*" + typeName[len(typeName)-1]
		default:
			panic(fmt.Errorf("unknown type for %s", field.Descriptor().String()))
		}
	}

	if field.Type().IsMap() {
		var key, val string

		if embeddedKey := field.Type().Key().Embed(); embeddedKey == nil {
			key = detectType(field.Type().Key().ProtoType().Proto())
		} else {
			key = "*" + embeddedKey.Name().String()
		}

		if embeddedVal := field.Type().Element().Embed(); embeddedVal == nil {
			val = detectType(field.Type().Element().ProtoType().Proto())
		} else {
			val = "*" + embeddedVal.Name().String()
		}

		return fmt.Sprintf("map[%s]%s", key, val)
	}

	if field.Type().IsRepeated() {
		typ += "[]"
	}

	typ += detectType(field.Descriptor().GetType())

	return
}

func getPackageName(target pgs.File) (string, string) {
	goPackage := target.Descriptor().GetOptions().GetGoPackage()

	if goPackage == "" {
		return target.Package().ProtoName().String(), target.File().Name().String()
	}

	if index := strings.Index(goPackage, ";"); index > 0 && index+1 < len(goPackage) {
		return goPackage[index+1:], goPackage[:index] + "/" + target.File().Name().String()
	}

	if index := strings.LastIndex(goPackage, "/"); index > 0 {
		return goPackage[index+1:], goPackage + "/" + target.File().Name().String()
	}

	return goPackage, target.File().Name().String()
}
