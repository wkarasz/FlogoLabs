package MyTimerTrigger

import (
	"io/ioutil"
	"encoding/json"
	"testing"
	"github.com/TIBCOSoftware/flogo-lib/core/trigger"
)

func getJsonMetadata() string {
	jsonMetadataBytes, err := ioutil.ReadFile("trigger.json")
	if err != nil {
		panic("No Json Metadata found for trigger.json path")
	}
	return string(jsonMetadataBytes)
}

const testConfig string = `{
  "id": "mytrigger",
  "settings": {
  },
  "handlers": [
    {
      "settings": {
        "notImmediate": "true",
	"startDate": "2018-01-01T12:00:00Z00:00",
	"repeating": "false",
	"seconds": "5"
      }
    }
  ]
}`

/*
func TestCreate(t *testing.T) {

	// New factory
	md := trigger.NewMetadata(getJsonMetadata())
	f := NewFactory(md)

	if f == nil {
		t.Fail()
	}

	// New Trigger
	config := trigger.Config{}
	json.Unmarshal([]byte(testConfig), config)
	trg := f.New(&config)

	if trg == nil {
		t.Fail()
	}
}
*/

func TestInit(t *testing.T) {
	// New  factory
	//f := &TimerFactory{}
	//tgr := f.New("flogo-timer")

	//runner := &TestRunner{}

	config := trigger.Config{}
	json.Unmarshal([]byte(testConfig), &config)
	//tgr.Init(config, runner)

}
