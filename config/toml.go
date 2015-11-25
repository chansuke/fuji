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
	"errors"
	"reflect"
	"regexp"
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
	Status  SectionMap `toml:"status"`
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

// SearchSection finds the section matched condition args.
func SearchSection(sections *[]ConfigSection, t, arg string) *ConfigSection {
	for _, section := range *sections {
		if section.Type == t && section.Arg == arg {
			return &section
		}
	}
	return nil
}

// SearchDeviceType find the device section matched type name string
func SearchDeviceType(sections *[]ConfigSection, arg string) *ConfigSection {
	for _, section := range *sections {
		if section.Type == "device" && section.Values["type"] == arg {
			return &section
		}
	}
	return nil
}
