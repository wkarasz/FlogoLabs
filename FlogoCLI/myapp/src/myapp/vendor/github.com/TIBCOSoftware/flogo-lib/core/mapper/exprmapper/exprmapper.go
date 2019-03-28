package exprmapper

import (
	"fmt"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/core/mapper/assign"
	"github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/expression"
	"github.com/TIBCOSoftware/flogo-lib/logger"

	//Pre registry all function for now
	_ "github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/function/array/length"
	_ "github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/function/number/random"
	_ "github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/function/string/concat"
	_ "github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/function/string/equals"
	_ "github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/function/string/equalsignorecase"
	_ "github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/function/string/length"
	_ "github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/function/string/substring"
)

var log = logger.GetLogger("expr-mapper")

const (
	MAP_TO_INPUT = "$INPUT"
)

func MapExpreesion(mapping *data.MappingDef, inputScope, outputScope data.Scope, resolver data.Resolver) error {
	mappingValue, err := GetExpresssionValue(mapping.Value, inputScope, resolver)
	if err != nil {
		return err
	}
	err = assign.SetValueToOutputScope(mapping.MapTo, outputScope, mappingValue)
	if err != nil {
		err = fmt.Errorf("Set value %+v to output [%s] error - %s", mappingValue, mapping.MapTo, err.Error())
		log.Error(err)
		return err
	}
	log.Debugf("Set value %+v to %s Done", mappingValue, mapping.MapTo)
	return nil
}

func GetExpresssionValue(mappingV interface{}, inputScope data.Scope, resolver data.Resolver) (interface{}, error) {
	mappingValue, ok := mappingV.(string)
	if !ok {
		return mappingV, nil
	}
	exp, err := expression.ParseExpression(mappingValue)
	if err == nil {
		//flogo expression
		log.Debugf("[%s] is an valid expression", mappingValue)
		expValue, err := exp.EvalWithScope(inputScope, resolver)
		if err != nil {
			return nil, fmt.Errorf("Execution failed for mapping [%s] due to error - %s", mappingValue, err.Error())
		}
		return expValue, nil
	} else {
		log.Debugf("[%s] is not an expression, take it as assign", mappingValue)
		return assign.GetMappingValue(mappingV, inputScope, resolver)
	}
}
