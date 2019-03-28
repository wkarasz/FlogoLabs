package ref

import (
	"errors"
	"strings"

	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

var log = logger.GetLogger("mapping-string")

type MappingRef struct {
	ref string
}

func NewMappingRef(ref string) *MappingRef {
	//Compatible TriggerData, the $TriggerData might in function or expression
	if strings.Index(ref, "$TriggerData") >= 0 {
		ref = strings.Replace(ref, "$TriggerData", "$flow", -1)
	}
	return &MappingRef{ref: ref}
}

func (m *MappingRef) GetRef() string {
	return m.ref
}

func (m *MappingRef) Eval(inputScope data.Scope, resovler data.Resolver) (interface{}, error) {
	log.Debugf("Eval mapping field %s", m.ref)

	if inputScope == nil {
		return nil, errors.New("Input scope cannot nil while eval mapping ref")
	}
	value, err := m.GetValue(inputScope, resovler)
	if err != nil {
		log.Errorf("Get From from ref error %+v", err)
	}

	log.Debugf("Eval mapping field result: %+v", value)
	return value, err

}

func (m *MappingRef) GetValue(inputScope data.Scope, resovler data.Resolver) (interface{}, error) {
	return resovler.Resolve(m.ref, inputScope)
}
