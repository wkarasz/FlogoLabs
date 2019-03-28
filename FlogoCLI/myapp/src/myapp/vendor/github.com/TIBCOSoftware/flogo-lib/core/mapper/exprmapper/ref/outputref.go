package ref

import (
	"fmt"
	"strings"

	"github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/json/field"

	"github.com/TIBCOSoftware/flogo-lib/core/data"
)

func GetValueFromOutputScope(mapfield *field.MappingField, outputtscope data.Scope) (interface{}, error) {
	fieldName, err := GetMapToAttrName(mapfield)
	if err != nil {
		return nil, err
	}
	log.Debugf("GetValueFromOutputScope field name %s", fieldName)

	attribute, exist := outputtscope.GetAttr(fieldName)
	log.Debugf("GetValueFromOutputScope field name %s and exist %t ", fieldName, exist)

	if exist {
		switch attribute.Type() {
		case data.TypeComplexObject:
			complexObject := attribute.Value().(*data.ComplexObject)
			object := complexObject.Value
			//Convert the object to exist struct.
			//TODO return interface rather than string
			if object == nil {
				return "{}", nil
			}
			return object, nil
		default:
			return attribute.Value(), nil
		}

	}
	return nil, fmt.Errorf("Cannot found attribute %s", fieldName)
}

func GetMapToAttrName(field *field.MappingField) (string, error) {
	fields := field.Getfields()
	return getFieldName(fields[0]), nil
}

func GetMapToPathFields(mapField *field.MappingField) (*field.MappingField, error) {
	fields := mapField.Getfields()

	if len(fields) == 1 && !HasArray(fields[0]) {
		return field.NewMappingField([]string{}), nil
	} else if HasArray(fields[0]) {
		arrayIndexPart := getArrayIndexPart(fields[0])
		fields[0] = arrayIndexPart
		return field.NewMappingField(fields), nil
	} else if len(fields) > 1 {
		if strings.HasSuffix(fields[0], "]") {
			//Root element is an array
			arrayIndexPart := getArrayIndexPart(fields[0])
			fields[0] = arrayIndexPart
			return field.NewMappingField(fields), nil
		} else {
			return field.NewMappingField(mapField.Getfields()[1:]), nil
		}
	} else {
		//Only attribute name no field name
		return field.NewMappingField([]string{}), nil
	}
}

func getFieldName(fieldname string) string {
	if strings.Index(fieldname, "[") > 0 && strings.Index(fieldname, "]") > 0 {
		return fieldname[:strings.Index(fieldname, "[")]
	}
	return fieldname
}

func HasArray(fieldname string) bool {
	if strings.Index(fieldname, "[") > 0 && strings.Index(fieldname, "]") > 0 {
		return true
	}
	return false
}

//getArrayIndexPart get array part of the string. such as name[0] return [0]
func getArrayIndexPart(fieldName string) string {
	if strings.Index(fieldName, "[") >= 0 {
		return fieldName[strings.Index(fieldName, "[") : strings.Index(fieldName, "]")+1]
	}
	return ""
}
