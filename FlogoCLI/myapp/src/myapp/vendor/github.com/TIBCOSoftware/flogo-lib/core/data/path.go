package data

import (
	"github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/json"
)

func PathGetValue(value interface{}, path string) (interface{}, error) {
	return json.GetPathValue(value, path)
}
