package json

import (
	"sync"

	"github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/json/field"
)

type JSONData struct {
	container *Container
	rw        sync.RWMutex
}

func SetStringValue(data interface{}, jsonData string, mappingField *field.MappingField) (interface{}, error) {
	jsonParsed, err := ParseJSON([]byte(jsonData))
	if err != nil {
		return nil, err

	}
	container := &JSONData{container: jsonParsed, rw: sync.RWMutex{}}
	err = handleSetValue(data, container, mappingField.Getfields())
	return container.container.object, err
}

func SetFieldValue(data interface{}, jsonData interface{}, mappingField *field.MappingField) (interface{}, error) {
	switch t := jsonData.(type) {
	case string:
		return SetStringValue(data, t, mappingField)
	default:
		jsonParsed, err := Consume(jsonData)
		if err != nil {
			return nil, err
		}
		container := &JSONData{container: jsonParsed, rw: sync.RWMutex{}}
		err = handleSetValue(data, container, mappingField.Getfields())
		if err != nil {
			return nil, err
		}
		return container.container.object, nil
	}

}

func handleSetValue(value interface{}, jsonData *JSONData, fields []string) error {

	jsonData.rw.Lock()
	defer jsonData.rw.Unlock()

	container := jsonData.container
	if hasArrayFieldInArray(fields) {
		arrayFields, fieldNameindex, arrayIndex := getArrayFieldName(fields)
		//No array field found
		if fieldNameindex == -1 {
			if arrayIndex == -2 {
				//Append
				err := container.ArrayAppend(value, arrayFields...)
				if err != nil {
					return err
				}
			} else {
				//set to exist index array
				size, err := container.ArrayCount(arrayFields...)
				if err != nil {
					return err
				}
				if arrayIndex > size-1 {
					err := container.ArrayAppend(value, arrayFields...)
					if err != nil {
						return err
					}
				} else {
					array := container.S(arrayFields...)
					_, err := array.SetIndex(value, arrayIndex)
					if err != nil {
						return err
					}
				}
			}
		} else {
			restFields := fields[fieldNameindex+1:]
			//Has field before [0]
			if arrayFields != nil {
				//check if array field exist
				if container.Exists(arrayFields...) {
					if restFields == nil || len(restFields) <= 0 {
						array, ok := container.Search(arrayFields...).Data().([]interface{})
						if ok {
							if arrayIndex > len(array)-1 {
								array = append(array, value)
							} else {
								array[arrayIndex] = value
							}
						}
						_, err := container.Set(array, arrayFields...)
						return err
					} else {
						var element *Container
						var err error
						count, err := container.ArrayCount(arrayFields...)
						if err != nil {
							return err
						}
						if arrayIndex > count-1 {
							maps := make(map[string]interface{})
							newObject, _ := Consume(maps)
							_, err = newObject.Set(value, restFields...)
							log.Debugf("new object %s", newObject.String())
							if err != nil {
								return err
							}

							err = container.ArrayAppend(maps, arrayFields...)
							if err != nil {
								return err
							}
							if !hasArrayFieldInArray(restFields) {
								return nil
							}

						}

						element, err = container.ArrayElement(arrayIndex, arrayFields...)
						if err != nil {
							return err
						}
						return handleSetValue(value, &JSONData{container: element, rw: sync.RWMutex{}}, restFields)
					}
				}

			} else if fieldNameindex == 0 && getFieldName(fields[fieldNameindex]) == "" {
				//Root only [0]
				array, ok := container.Data().([]interface{})
				if !ok {
					var err error
					array, err = ToArray(container.Data())
					if err != nil {
						array = make([]interface{}, arrayIndex+1)
					}

				}
				container.object = array

				if restFields == nil || len(restFields) <= 0 {
					_, err := container.SetIndex(value, arrayIndex)
					return err
				} else {
					if hasArrayFieldInArray(restFields) {
						//We need make sure the array element is not empty
						count, err := container.ArrayCount(arrayFields...)
						if err != nil {
							return err
						}
						if arrayIndex > count-1 {
							err = container.ArrayAppend(map[string]interface{}{})
							if err != nil {
								return err
							}
						}

						arrayElement, err := container.ArrayElement(arrayIndex)
						if err != nil {
							return err
						}
						if arrayElement.object == nil {
							arrayElement.object = map[string]interface{}{}
						}
						err = handleSetValue(value, &JSONData{container: arrayElement, rw: sync.RWMutex{}}, restFields)
						if err != nil {
							return err
						}
						//Set value back to array
						_, err = container.SetIndex(arrayElement.object, arrayIndex)
						return err
					}

					arrayElement, err := getArrayElementValue(container, arrayFields, restFields, arrayIndex, value)
					if err != nil {
						return err
					}

					_, err = container.SetIndex(arrayElement.object, arrayIndex)
					return err
				}
			}

			//Not exist in the exist container, create new one
			array, err := container.ArrayOfSize(arrayIndex+1, arrayFields...)
			if err != nil {
				return err
			}

			if hasArrayFieldInArray(restFields) {
				//We need make sure the array element is not empty
				arrayElement, err := array.ArrayElement(arrayIndex)
				if err != nil {
					return err
				}
				if arrayElement.object == nil {
					arrayElement.object = map[string]interface{}{}
				}

				err = handleSetValue(value, &JSONData{container: arrayElement, rw: sync.RWMutex{}}, restFields)
				if err != nil {
					return err
				}
				//Set value back to array
				_, err = container.S(arrayFields...).SetIndex(arrayElement.object, arrayIndex)
				return err
			}

			maps := make(map[string]interface{})
			newObject, _ := Consume(maps)
			_, err = newObject.Set(value, restFields...)
			log.Debugf("new object %s", newObject.String())
			if err != nil {
				return err
			}
			_, err = array.SetIndex(newObject.object, arrayIndex)
		}
	} else {
		_, err := jsonData.container.Set(value, fields...)
		if err != nil {
			return err
		}
	}
	return nil
}

func getArrayElementValue(container *Container, arrayFields []string, restFields []string, arrayIndex int, value interface{}) (*Container, error) {
	count, err := container.ArrayCount(arrayFields...)
	if err != nil {
		return nil, err
	}

	if arrayIndex > count-1 {
		maps := make(map[string]interface{})
		newObject, _ := Consume(maps)
		_, err := newObject.Set(value, restFields...)
		log.Debugf("new object %s", newObject.String())
		if err != nil {
			return nil, err
		}
		value = newObject.object
	}

	arrayElement, err := container.ArrayElement(arrayIndex)
	if err != nil {
		return nil, err
	}
	if arrayElement.object == nil {
		arrayElement.object = map[string]interface{}{}
	}
	_, err = arrayElement.Set(value, restFields...)
	return arrayElement, err
}
