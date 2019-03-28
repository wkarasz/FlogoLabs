package exprmapper

import (
	"encoding/json"
	"fmt"
	"github.com/TIBCOSoftware/flogo-lib/core/mapper/assign"
	"github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/json/field"
	"runtime/debug"
	"strings"

	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/expression"
	flogojson "github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/json"
	"github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/ref"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

var arraylog = logger.GetLogger("array-mapping")

const (
	PRIMITIVE = "primitive"
	FOREACH   = "foreach"
	NEWARRAY  = "NEWARRAY"
)

type ArrayMapping struct {
	From   interface{}     `json:"from"`
	To     string          `json:"to"`
	Type   string          `json:"type"`
	Fields []*ArrayMapping `json:"fields,omitempty"`
}

func (a *ArrayMapping) Validate() error {
	//Validate root from/to field
	if a.From == nil {
		return fmt.Errorf("The array mapping validation failed for the mapping [%s]. Ensure valid array is mapped in the mapper. ", a.To)
	}

	if a.To == "" || len(a.To) <= 0 {
		return fmt.Errorf("The array mapping validation failed for the mapping [%s]. Ensure valid array is mapped in the mapper. ", a.From)
	}

	if a.Type == FOREACH {
		//Validate root from/to field
		if a.From == NEWARRAY {
			//Make sure no array ref fields exist
			for _, field := range a.Fields {
				if field.Type == FOREACH {
					return field.Validate()
				}
				stringVal, ok := field.From.(string)
				if ok && ref.IsArrayMapping(stringVal) {
					return fmt.Errorf("The array mapping validation failed, due to invalid new array mapping [%s]", stringVal)
				}

			}
		} else {
			for _, field := range a.Fields {
				if field.Type == FOREACH {
					return field.Validate()
				}
			}
		}

	}

	return nil
}

func (a *ArrayMapping) DoArrayMapping(inputScope, outputScope data.Scope, resolver data.Resolver) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%+v", r)
			logger.Debugf("StackTrace: %s", debug.Stack())
		}
	}()

	//First level must be foreach
	switch a.Type {
	case FOREACH:
		//First Level
		var fromValue interface{}
		var err error

		stringVal, _ := a.From.(string)
		if strings.EqualFold(stringVal, NEWARRAY) {
			log.Debugf("Init a new array for field", a.To)
			fromValue = make([]interface{}, 1)
		} else {
			fromValue, err = GetExpresssionValue(stringVal, inputScope, resolver)
		}

		//Check if fields is empty for primitive array mapping
		if a.Fields == nil || len(a.Fields) <= 0 {
			//Set value directlly to MapTo field
			return assign.SetValueToOutputScope(a.To, outputScope, fromValue)
		}

		//Loop array
		fromArrayvalues, ok := fromValue.([]interface{})
		if !ok {
			//Try to convert to array.
			fromArrayvalues, err = data.CoerceToArray(fromValue)
			if err != nil {
				return fmt.Errorf("Failed to get array value from [%s], due to error- [%s] value not an array", a.From, a.From)
			}
		}

		toRef := ref.NewMappingRef(a.To)
		toMapField, err := field.ParseMappingField(toRef.GetRef())
		if err != nil {
			return err
		}
		toValue, err := ref.GetValueFromOutputScope(toMapField, outputScope)
		if err != nil {
			return err
		}
		toValue = toInterface(toValue)
		objArray := make([]interface{}, len(fromArrayvalues))
		for i, _ := range objArray {
			objArray[i] = make(map[string]interface{})
		}

		mappingField, err := ref.GetMapToPathFields(toMapField)
		if err != nil {
			return fmt.Errorf("Get fields from mapping string error, due to [%s]", err.Error())
		}
		if mappingField != nil && len(mappingField.Getfields()) > 0 {
			vv, err := flogojson.SetFieldValue(objArray, toValue, mappingField)
			if err != nil {
				return err
			}
			log.Debugf("Set Value return as %+v", vv)
		} else {
			toValue = objArray
		}

		if err != nil {
			return err
		}

		for i, arrayV := range fromArrayvalues {
			err = a.iterator(arrayV, objArray[i], a.Fields, inputScope, outputScope, resolver)
			if err != nil {
				log.Error(err)
				return err
			}
		}

		//Get Value from fields
		toFieldName, err := ref.GetMapToAttrName(toMapField)
		if err != nil {
			return err
		}

		if len(mappingField.Getfields()) > 0 {
			return assign.SetAttribute(toFieldName, toValue, outputScope)
		}
		return assign.SetAttribute(toFieldName, getFieldValue(toValue, toFieldName), outputScope)
	}
	return nil
}

func (a *ArrayMapping) mappingDef() *data.MappingDef {
	return &data.MappingDef{MapTo: a.To, Value: a.From, Type: data.MtExpression}
}

func (a *ArrayMapping) iterator(fromValue, value interface{}, fields []*ArrayMapping, inputScope, outputScope data.Scope, resolver data.Resolver) error {
	for _, arrayField := range fields {
		switch arrayField.Type {
		//Backward compatibility
		case PRIMITIVE, "expression":
			fValue, err := getArrayExpresssionValue(fromValue, arrayField.From, inputScope, resolver)
			if err != nil {
				return err
			}
			log.Debugf("Array mapping from %s 's value %+v", arrayField.From, fValue)
			tomapField, err := field.ParseMappingField(ref.GetFieldNameFromArrayRef(arrayField.To))
			if err != nil {
				return err
			}
			err = arrayField.DoMap(fValue, value, tomapField, inputScope, outputScope, resolver)
			if err != nil {
				return err
			}
		case "assign", "literal":
			fValue, err := getArrayValue(fromValue, arrayField.From, inputScope, resolver)
			if err != nil {
				return err
			}
			log.Debugf("Array mapping from %s 's value %+v", arrayField.From, fValue)
			tomapField, err := field.ParseMappingField(ref.GetFieldNameFromArrayRef(arrayField.To))
			if err != nil {
				return err
			}
			err = arrayField.DoMap(fValue, value, tomapField, inputScope, outputScope, resolver)
			if err != nil {
				return err
			}
		case "object":
			//TODO support object mapping
		case FOREACH:
			var fromArrayvalues []interface{}
			if strings.EqualFold(arrayField.From.(string), NEWARRAY) {
				log.Debugf("Init a new array for field", arrayField.To)
				fromArrayvalues = make([]interface{}, 1)
			} else {
				fValue, err := getArrayExpresssionValue(fromValue, arrayField.From, inputScope, resolver)
				if err != nil {
					return err
				}
				var ok bool
				fromArrayvalues, ok = fValue.([]interface{})
				if !ok {
					return fmt.Errorf("Failed to get array value from [%s], due to error- value not an array", fValue)
				}
			}

			toValue := toInterface(value)
			objArray := make([]interface{}, len(fromArrayvalues))
			for i, _ := range objArray {
				objArray[i] = make(map[string]interface{})
			}

			tomapField, err := field.ParseMappingField(ref.GetFieldNameFromArrayRef(arrayField.To))
			if err != nil {
				return err
			}
			_, err = flogojson.SetFieldValue(objArray, toValue, tomapField)
			if err != nil {
				return err
			}
			//Check if fields is empty for primitive array mapping
			if arrayField.Fields == nil || len(arrayField.Fields) <= 0 {
				for f, v := range fromArrayvalues {
					objArray[f] = v
				}
				continue
			}

			for i, arrayV := range fromArrayvalues {
				err = a.iterator(arrayV, objArray[i], arrayField.Fields, inputScope, outputScope, resolver)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil

}

func (a *ArrayMapping) DoMap(fromValue, value interface{}, tomapField *field.MappingField, inputScope, outputScope data.Scope, resolver data.Resolver) error {
	switch a.Type {
	case PRIMITIVE, "assign", "literal", "expression", "object":
		_, err := flogojson.SetFieldValue(fromValue, value, tomapField)
		if err != nil {
			return err
		}
	case FOREACH:
		fmt.Println("============")
		fValue, err := getArrayExpresssionValue(fromValue, a.From, inputScope, resolver)
		if err != nil {
			return err
		}
		tValue, err := getArrayExpresssionValue(value, a.To, inputScope, resolver)
		if err != nil {
			return err
		}
		err = a.iterator(fValue, tValue, a.Fields, inputScope, outputScope, resolver)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *ArrayMapping) RemovePrefixForMapTo() {
	if a == nil {
		return
	}

	a.To = RemovePrefixInput(a.To)

	if a.Type == FOREACH {
		//Validate root from/to field
		if a.From == NEWARRAY {
			//Make sure no array ref fields exist
			for _, field := range a.Fields {
				if field.Type == FOREACH {
					field.RemovePrefixForMapTo()
				} else {
					field.To = RemovePrefixInput(field.To)
				}
			}

		} else {
			for _, field := range a.Fields {
				if field.Type == FOREACH {
					field.RemovePrefixForMapTo()
				} else {
					field.To = RemovePrefixInput(field.To)
				}
			}
		}

	}
}

func ParseArrayMapping(arrayDatadata interface{}) (*ArrayMapping, error) {
	amapping := &ArrayMapping{}
	switch t := arrayDatadata.(type) {
	case string:
		err := json.Unmarshal([]byte(t), amapping)
		if err != nil {
			return nil, err
		}
	case interface{}:
		s, err := data.CoerceToString(t)
		if err != nil {
			return nil, fmt.Errorf("Convert array mapping value to string error, due to [%s]", err.Error())
		}
		err = json.Unmarshal([]byte(s), amapping)
		if err != nil {
			return nil, err
		}
	}
	return amapping, nil
}

func toInterface(data interface{}) interface{} {

	switch t := data.(type) {
	case string:
		if strings.EqualFold("{}", t) {
			return make(map[string]interface{})
		}
	default:
		if t == nil {
			//TODO maybe consider other types as well
			return make(map[string]interface{})
		}
	}
	return data
}

func getFieldValue(value interface{}, fieldName string) interface{} {
	switch t := value.(type) {
	case map[string]interface{}:
		return t[fieldName]
	default:
		return value
	}
	return value
}

func getArrayExpresssionValue(object interface{}, expressionRef interface{}, inputScope data.Scope, resolver data.Resolver) (interface{}, error) {
	stringVal, ok := expressionRef.(string)
	if !ok {
		//Non string value
		return expressionRef, nil
	}
	exp, err := expression.ParseExpression(stringVal)
	if err == nil {
		//flogo expression
		expValue, err := exp.EvalWithData(object, inputScope, resolver)
		if err != nil {
			err = fmt.Errorf("Execution failed for mapping [%s] due to error - %s", stringVal, err.Error())
			log.Error(err)
			return nil, err
		}
		return expValue, nil
	} else {
		return getArrayValue(object, expressionRef, inputScope, resolver)
	}

}

func getArrayValue(object interface{}, expressionRef interface{}, inputScope data.Scope, resolver data.Resolver) (interface{}, error) {
	var fromValue interface{}

	stringVal, ok := expressionRef.(string)
	if !ok {
		return expressionRef, nil
	}
	if ref.IsArrayMapping(stringVal) {
		reference := ref.GetFieldNameFromArrayRef(stringVal)
		toMapField, err := field.ParseMappingField(reference)
		if err != nil {
			return nil, err
		}

		fromValue, err = flogojson.GetFieldValue(object, toMapField)
		if err != nil {
			return nil, err
		}

	} else if strings.HasPrefix(stringVal, "$") {
		fromRef := ref.NewMappingRef(stringVal)
		var err error
		fromValue, err = fromRef.GetValue(inputScope, resolver)
		if err != nil {
			return nil, fmt.Errorf("Get value from [%s] failed, due to error - %s", stringVal, err.Error())
		}
	} else {
		fromValue = expressionRef
	}

	return fromValue, nil

}

func RemovePrefixInput(str string) string {
	if str != "" && strings.HasPrefix(str, MAP_TO_INPUT) {
		//Remove $INPUT for mapTo
		newMapTo := str[len(MAP_TO_INPUT):]
		if strings.HasPrefix(newMapTo, ".") {
			newMapTo = newMapTo[1:]
		}
		str = newMapTo
	}
	return str
}
