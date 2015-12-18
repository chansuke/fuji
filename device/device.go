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

package device

import (
	log "github.com/Sirupsen/logrus"

	"github.com/shiguredo/fuji/broker"
	"github.com/shiguredo/fuji/config"
	"github.com/shiguredo/fuji/message"
)

type Devicer interface {
	Start(chan message.Message) error
	DeviceType() string
	Stop() error
	AddSubscribe() error
}

// NewDevices is a factory method to create various kind of devices from config.Config
func NewDevices(conf config.Config, brokers []*broker.Broker) ([]Devicer, []DeviceChannel, error) {
	var ret []Devicer
	var devChannels []DeviceChannel

	var err error
	for _, section := range conf.Sections {
		if section.Type != "device" {
			continue
		}

		var device Devicer

		devChan := NewDeviceChannel()
		devChannels = append(devChannels, devChan)

		switch section.Values["type"] {
		case "dummy":
			device, err = NewDummyDevice(section, brokers, devChan)
			if err != nil {
				log.Errorf("could not create dummy device, %v", err)
				continue
			}
		case "serial":
			device, err = NewSerialDevice(section, brokers, devChan)
			if err != nil {
				log.Errorf("could not create serial device, %v", err)
				continue
			}
		default:
			log.Warnf("unknown device type, %v", section.Arg)
			continue
		}
		ret = append(ret, device)
	}

	return ret, devChannels, nil
}
