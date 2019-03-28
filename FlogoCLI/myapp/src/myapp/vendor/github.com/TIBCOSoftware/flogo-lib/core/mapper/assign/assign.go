package assign

import (
	"errors"
	"fmt"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/json"
	"github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/json/field"
	"github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/ref"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"strings"
)

var log = logger.GetLogger("assign-mapper")

func MapAssign(mapping *data.MappingDef, inputScope, outputScope data.Scope, resolver data.Resolver) error {
	mappingValue, err := GetMappingValue(mapping.Value, inputScope, resolver)
	if err != nil {
		return err
	}
	err = SetValueToOutputScope(mapping.MapTo, outputScope, mappingValue)
	if err != nil {
		err = fmt.Errorf("Set value %+v to output [%s] error - %s", mappingValue, mapping.MapTo, err.Error())
		log.Error(err)
		return err
	}
	log.Debugf("Set value %+v to %s Done", mappingValue, mapping.MapTo)
	return nil
}

func GetMappingValue(mappingV interface{}, inputScope data.Scope, resolver data.Resolver) (interface{}, error) {
	mappingValue, ok := mappingV.(string)
	if !ok {
		return mappingV, nil
	}
	if !isMappingRef(mappingValue) {
		log.Debug("Mapping value is literal set directly to field")
		log.Debugf("Mapping ref %s and value %+v", mappingValue, mappingValue)
		return mappingValue, nil
	} else {
		mappingref := ref.NewMappingRef(mappingValue)
		mappingValue, err := mappingref.GetValue(inputScope, resolver)
		if err != nil {
			return nil, fmt.Errorf("Get value from ref [%s] error - %s", mappingref.GetRef(), err.Error())

		}
		log.Debugf("Mapping ref %s and value %+v", mappingValue, mappingValue)
		return mappingValue, nil
	}
	return nil, nil
}

func isMappingRef(mappingref string) bool {
	if mappingref == "" || !strings.HasPrefix(mappingref, "$") {
		return false
	}
	return true
}

func SetValueToOutputScope(mapTo string, outputScope data.Scope, value interface{}) error {
	mapField, err := field.ParseMappingField(mapTo)
	if err != nil {
		return err
	}

	actRootField, err := ref.GetMapToAttrName(mapField)
	if err != nil {
		return err
	}

	fields := mapField.Getfields()
	if len(fields) == 1 && !ref.HasArray(fields[0]) {
		//No complex mapping exist
		return SetAttribute(actRootField, value, outputScope)
	} else if ref.HasArray(fields[0]) || len(fields) > 1 {
		//Complex mapping
		return settValue(mapField, actRootField, outputScope, value)
	} else {
		return fmt.Errorf("No field name found for mapTo [%s]", mapTo)
	}

}

func settValue(mapField *field.MappingField, fieldName string, outputScope data.Scope, value interface{}) error {
	existValue, err := ref.GetValueFromOutputScope(mapField, outputScope)
	if err != nil {
		return err
	}
	newValue, err2 := SetPathValue(existValue, mapField, value)
	if err2 != nil {
		return err2
	}

	return SetAttribute(fieldName, newValue, outputScope)
}

func SetPathValue(value interface{}, mapField *field.MappingField, attrvalue interface{}) (interface{}, error) {
	pathfields, err := ref.GetMapToPathFields(mapField)
	if err != nil {
		return nil, err
	}

	log.Debugf("Set value %+v to fields %s", value, pathfields)
	value, err = json.SetFieldValue(attrvalue, value, pathfields)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func SetAttribute(fieldName string, value interface{}, outputScope data.Scope) error {
	//Set Attribute value back to attribute
	attribute, exist := outputScope.GetAttr(fieldName)
	if exist {
		switch attribute.Type() {
		case data.TypeComplexObject:
			complexObject := attribute.Value().(*data.ComplexObject)
			newComplexObject := &data.ComplexObject{Metadata: complexObject.Metadata, Value: value}
			outputScope.SetAttrValue(fieldName, newComplexObject)
		default:
			outputScope.SetAttrValue(fieldName, value)
		}

	} else {
		return errors.New("Cannot found attribute " + fieldName + " at output scope")
	}
	return nil
}
