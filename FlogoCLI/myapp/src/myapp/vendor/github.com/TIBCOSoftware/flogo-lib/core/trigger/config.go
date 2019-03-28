package trigger

import (
	"github.com/TIBCOSoftware/flogo-lib/core/action"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/util"
)

// Config is the configuration for a Trigger
type Config struct {
	Name     string                 `json:"name"`
	Id       string                 `json:"id"`
	Ref      string                 `json:"ref"`
	Settings map[string]interface{} `json:"settings"`
	Output   map[string]interface{} `json:"output"`
	Handlers []*HandlerConfig       `json:"handlers"`

	// Deprecated: Use Output
	Outputs map[string]interface{} `json:"outputs"`
}

func (c *Config) FixUp(metadata *Metadata) {

	//for backwards compatibility
	if len(c.Output) == 0 {
		c.Output = c.Outputs
	}

	for k, v := range c.Settings {
		strValue, ok := v.(string)
		if ok {
			if strValue != "" && strValue[0] == '$' {
				// Let resolver resolve value
				val, err := data.GetBasicResolver().Resolve(strValue, nil)
				if err != nil {
					val = strValue
				}
				c.Settings[k] = val
			}
		}
	}

	// fix up top-level outputs
	for name, value := range c.Output {

		attr, ok := metadata.Output[name]

		if ok {
			newValue, err := data.CoerceToValue(value, attr.Type())

			if err != nil {
				//todo handle error
			} else {
				c.Output[name] = newValue
			}
		}
	}

	idGen, _ := util.NewGenerator()

	// fix up handler outputs
	for _, hc := range c.Handlers {

		hc.parent = c

		//for backwards compatibility
		if hc.ActionId == "" {
			hc.ActionId = idGen.NextAsString()
		}

		//for backwards compatibility
		if len(hc.Output) == 0 {
			hc.Output = hc.Outputs
		}

		for k, v := range hc.Settings {
			strValue, ok := v.(string)
			if ok {
				if strValue != "" && strValue[0] == '$' {
					// Let resolver resolve value
					val, err := data.GetBasicResolver().Resolve(strValue, nil)
					if err != nil {
						val = strValue
					}
					hc.Settings[k] = val
				}
			}
		}
		// fix up outputs
		for name, value := range hc.Output {

			attr, ok := metadata.Output[name]

			if ok {
				newValue, err := data.CoerceToValue(value, attr.Type())

				if err != nil {
					//todo handle error
				} else {
					hc.Output[name] = newValue
				}
			}
		}
	}
}

func (c *Config) GetSetting(setting string) string {

	strVal, err := data.CoerceToString(c.Settings[setting])

	if err != nil {
		return ""
	}

	return strVal
}

type HandlerConfig struct {
	parent   *Config
	Name     string                 `json:"name,omitempty"`
	Settings map[string]interface{} `json:"settings"`
	Output   map[string]interface{} `json:"output"`
	Action   *ActionConfig

	// Deprecated: Use Action (*action.Config)
	ActionId string `json:"actionId"`
	// Deprecated: Use Action (*action.Config)
	ActionMappings *data.IOMappings `json:"actionMappings,omitempty"`
	// Deprecated: Use Action (*action.Config)
	ActionOutputMappings []*data.MappingDef `json:"actionOutputMappings,omitempty"`
	// Deprecated: Use Action (*action.Config)
	ActionInputMappings []*data.MappingDef `json:"actionInputMappings,omitempty"`
	// Deprecated: Use Output
	Outputs map[string]interface{} `json:"outputs"`
}

// ActionConfig is the configuration for the Action
type ActionConfig struct {
	*action.Config

	Mappings *data.IOMappings `json:"mappings"`

	Act action.Action
}

func (hc *HandlerConfig) GetSetting(setting string) string {

	strVal, err := data.CoerceToString(hc.Settings[setting])

	if err != nil {
		return ""
	}

	return strVal
}
