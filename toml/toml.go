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

package toml

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	ReU0   *regexp.Regexp
	ReWild *regexp.Regexp
)

type Config struct {
	GatewayName string
	BrokerNames []string

	Sections []ConfigSection
}

type ValueMap map[string]string

type ConfigSection struct {
	Title string
	Type  string
	Name  string
	Arg   string

	Values ValueMap
}

type ConfigToml struct {
	Gateway SectionMap `toml:"gateway"`
	Brokers SectionMap `toml:"broker"`
	Devices SectionMap `toml:"device"`
}

type SectionMap map[string]interface{}

type AnyError interface{}

type Error string

func (e Error) Error() string {
	return string(e)
}

// NilOrString defines the value is nil or empty
type NilOrString interface{}

// init is automatically invoked at initial time.
func init() {
	ReU0 = regexp.MustCompile("\u0000")
	ReWild = regexp.MustCompile("[+#]+")
}

func IsNil(str NilOrString) bool {
	if str == nil {
		return true
	}
	return false
}
func String(str NilOrString) string {
	stringValue, ok := str.(string)
	if ok == false {
		return ("nil")
	}
	return stringValue
}

func addConfigSections(configSections []ConfigSection, title string, sectionMap SectionMap) []ConfigSection {
	for name, values := range sectionMap {
		t := strings.Split(name, "/")
		if len(t) > 2 {
			log.Errorf("invalid section(slash), %v", t)
			continue
		}

		values_ := values.([]map[string]interface{})
		valueMap := make(map[string]string)

		for _, m := range values_ {
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

// ValidMqttPublishTopic validates the Topic is validate or not
// This is used with validator packages.
func ValidMqttPublishTopic(v interface{}, param string) error {
	str := reflect.ValueOf(v)
	if str.Kind() != reflect.String {
		return errors.New("ValidMqttPublishTopic only validates strings")
	}
	if !utf8.ValidString(str.String()) {
		return errors.New("not a valid UTF8 string")
	}

	if ReU0.FindString(str.String()) != "" {
		return errors.New("Topic SHALL NOT include \\U0000 character")
	}

	if ReWild.FindString(str.String()) != "" {
		return errors.New("SHALL NOT MQTT pub-topic include wildard character")
	}
	return nil
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
	valueMap := make(ValueMap)
	for name, value := range configToml.Gateway {

		if name == "name" {
			config.GatewayName = value.(string)
			if config.GatewayName == "" {
				return config, fmt.Errorf("gateway has not name")
			}
		}

		switch value.(type) {
		case int64:
			valueMap[name] = strconv.FormatInt(value.(int64), 10)
		case bool:
			valueMap[name] = strconv.FormatBool(value.(bool))
		default:
			valueMap[name] = value.(string)
		}

	}

	rt := ConfigSection{
		Title:  "gateway",
		Type:   "gateway",
		Values: valueMap,
	}
	sections = append(sections, rt)

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
