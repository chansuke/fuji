// Copyright 2015 Shiguredo Inc. <fuji@shiguredo.jp>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"strconv"
	"strings"
)

func buildUniqueValueMap(values map[string]interface{}) map[string]string {
	valueMap := make(map[string]string)

	for k, v := range values {
		switch v.(type) {
		case int64:
			valueMap[k] = strconv.FormatInt(v.(int64), 10)
		case bool:
			valueMap[k] = strconv.FormatBool(v.(bool))
		default:
			valueMap[k] = v.(string)
		}
	}

	return valueMap
}

func buildMultipleValueMap(values []map[string]interface{}) map[string]string {
	valueMap := make(map[string]string)

	for _, m := range values {
		for k, v := range m {
			switch v.(type) {
			case int64:
				valueMap[k] = strconv.FormatInt(v.(int64), 10)
			case bool:
				valueMap[k] = strconv.FormatBool(v.(bool))
			default:
				valueMap[k] = v.(string)
			}
		}
	}

	return valueMap
}

func getGatewayName(gatewaySectionMap SectionMap) (string, error) {
	for name, value := range gatewaySectionMap {
		if name == "name" {
			gatewayName := value.(string)
			if gatewayName == "" {
				return "", fmt.Errorf("gateway has not name")
			}
			return gatewayName, nil
		}
	}
	return "", nil
}

func addGatewaySection(configSections []ConfigSection, gatewaySectionMap SectionMap) []ConfigSection {
	valueMap := make(ValueMap)
	for name, value := range gatewaySectionMap {

		switch value.(type) {
		case int64:
			valueMap[name] = strconv.FormatInt(value.(int64), 10)
		case bool:
			valueMap[name] = strconv.FormatBool(value.(bool))
		default:
			valueMap[name] = value.(string)
		}

	}

	if len(valueMap) > 0 {
		rt := ConfigSection{
			Title:  "gateway",
			Type:   "gateway",
			Values: valueMap,
		}
		configSections = append(configSections, rt)
	}

	return configSections
}

func addStatusSections(configSections []ConfigSection, statusSectionMap SectionMap) []ConfigSection {
	valueMap := make(ValueMap)
	for name, value := range statusSectionMap {

		switch value.(type) {
		case int64:
			valueMap[name] = strconv.FormatInt(value.(int64), 10)
		case bool:
			valueMap[name] = strconv.FormatBool(value.(bool))
		case []map[string]interface{}:
			// do nothing
		default:
			valueMap[name] = value.(string)
		}

	}
	if len(valueMap) > 0 {
		rt := ConfigSection{
			Title:  "status",
			Type:   "status",
			Values: valueMap,
		}
		configSections = append(configSections, rt)
	}

	for name, value := range statusSectionMap {

		valueMap := make(ValueMap)
		switch value.(type) {
		case []map[string]interface{}:
			{
				m := value.([]map[string]interface{})
				for _, v := range m {
					for k, vv := range v {
						switch vv.(type) {
						case int64:
							valueMap[k] = strconv.FormatInt(vv.(int64), 10)
						case bool:
							valueMap[k] = strconv.FormatBool(vv.(bool))
						default:
							valueMap[k] = vv.(string)
						}
					}
				}
			}
		}

		if len(valueMap) > 0 {
			rt := ConfigSection{
				Title:  "status",
				Type:   "status",
				Name:   name,
				Values: valueMap,
			}
			configSections = append(configSections, rt)
		}
	}

	return configSections
}

func addConfigSections(configSections []ConfigSection, title string, sectionMap SectionMap) []ConfigSection {
	for name, values := range sectionMap {
		t := strings.Split(name, "/")
		if len(t) > 2 {
			log.Errorf("invalid section(slash), %v", t)
			continue
		}

		var valueMap map[string]string
		switch title {
		case "device":
			valueMap = buildUniqueValueMap(values.(map[string]interface{}))
		case "broker":
			valueMap = buildMultipleValueMap(values.([]map[string]interface{}))
		default:
			valueMap = buildMultipleValueMap(values.([]map[string]interface{}))
		}

		if len(valueMap) == 0 {
			continue
		}

		rt := ConfigSection{
			Title:  title,
			Type:   title,
			Name:   t[0],
			Values: valueMap,
		}

		if len(t) == 2 { // if args exists, store it
			rt.Arg = t[1]
		}

		configSections = append(configSections, rt)
	}

	return configSections
}

// LoadConfig loads toml format file from confPath arg and returns []ConfigSection.
// ConfigSection has a Type, Name and arg.
// example:
// [broker."sango"]
// [broker."sango/1"]
// [broker."sango/2"]
//
// ret = [
//   ConfigSection{Type: "broker", Name: "sango"},
//   ConfigSection{Type: "broker", Name: "sango", Arg: "1"},
//   ConfigSection{Type: "broker", Name: "sango", Arg: "2"},
// ]
func LoadConfig(confPath string) (Config, error) {
	dat, err := ioutil.ReadFile(confPath)
	if err != nil {
		return Config{}, err
	}

	return LoadConfigByte(dat)
}

// LoadConfigByte returns []ConfigSection from []byte.
// This is invoked from LoadConfig.
func LoadConfigByte(conf []byte) (Config, error) {
	config := Config{}
	var configToml ConfigToml

	if err := toml.Unmarshal(conf, &configToml); err != nil {
		return config, err
	}

	var sections []ConfigSection
	var bn []string

	// gateway section
	gatewayName, err := getGatewayName(configToml.Gateway)
	if err != nil {
		return config, err
	}
	config.GatewayName = gatewayName
	sections = addGatewaySection(sections, configToml.Gateway)

	// status section
	sections = addStatusSections(sections, configToml.Status)

	// broker sections
	sections = addConfigSections(sections, "broker", configToml.Brokers)

	// device sections
	sections = addConfigSections(sections, "device", configToml.Devices)

	// broker names
	for name, _ := range configToml.Brokers {
		t := strings.Split(name, "/")
		if len(t) > 2 {
			continue
		}
		bn = append(bn, t[0])
	}

	config.Sections = sections
	config.BrokerNames = bn

	return config, nil
}
