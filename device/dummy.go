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
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	validator "gopkg.in/validator.v2"

	"github.com/shiguredo/fuji/broker"
	"github.com/shiguredo/fuji/config"
	"github.com/shiguredo/fuji/message"
	"github.com/shiguredo/fuji/utils"
)

// DummyDevice is an dummy device which outputs only specified payload.
type DummyDevice struct {
	Name       string `validate:"max=256,regexp=[^/]+,validtopic"`
	Broker     []*broker.Broker
	BrokerName string
	QoS        byte `validate:"min=0,max=2"`
	InputPort  InputPortType
	Interval   int    `validate:"min=1"`
	Payload    []byte `validate:"max=4096"`
	Type       string `validate:"max=256"`
	Retain     bool
	Subscribe  bool
	DeviceChan DeviceChannel // GW -> device
}

// String retruns dummy device information
func (dummyDevice *DummyDevice) String() string {
	return fmt.Sprintf("%#v", dummyDevice)
}

// NewDummyDevice creates dummy device which outputs specified string/binary payload.
func NewDummyDevice(section config.ConfigSection, brokers []*broker.Broker, devChan DeviceChannel) (DummyDevice, error) {
	ret := DummyDevice{
		Name:       section.Name,
		DeviceChan: devChan,
	}
	values := section.Values
	bname, ok := section.Values["broker"]
	if !ok {
		return ret, fmt.Errorf("broker does not set")
	}

	for _, b := range brokers {
		if b.Name == bname {
			ret.Broker = brokers
		}
	}
	if ret.Broker == nil {
		return ret, fmt.Errorf("broker does not exists: %s", bname)
	}
	ret.BrokerName = bname

	qos, err := strconv.Atoi(values["qos"])
	if err != nil {
		return ret, fmt.Errorf("qos parse failed, %v", err)
	}
	ret.QoS = byte(qos)

	interval, err := strconv.Atoi(values["interval"])
	if err != nil {
		return ret, err
	} else {
		ret.Interval = int(interval)
	}
	ret.Type = values["type"]
	ret.Payload, err = utils.ParsePayload(values["payload"])
	if err != nil {
		log.Warnf("invalid payload, but continue")
	}
	ret.Retain = false
	if values["retain"] == "true" {
		ret.Retain = true
	}

	sub, ok := values["subscribe"]
	if ok && sub == "true" {
		ret.Subscribe = true
	}

	// Validation
	if err := ret.Validate(); err != nil {
		return ret, err
	}
	return ret, nil
}

func (device *DummyDevice) Validate() error {
	validator := validator.NewValidator()
	validator.SetValidationFunc("validtopic", config.ValidMqttPublishTopic)
	if err := validator.Validate(device); err != nil {
		return err
	}
	return nil
}

// Start starts dummy goroutine
func (device DummyDevice) Start(channel chan message.Message) error {
	log.Info("start dummy device")
	go device.MainLoop(channel)

	return nil
}

// MainLoop is an mainloop of dummy device.
func (device DummyDevice) MainLoop(channel chan message.Message) error {
	ticker := time.NewTicker(time.Duration(device.Interval) * time.Second)

	for {
		select {
		case <-ticker.C:
			msg := message.Message{
				Sender:     device.Name,
				Type:       device.Type,
				QoS:        device.QoS,
				Retained:   device.Retain,
				Body:       []byte(device.Payload),
				BrokerName: device.BrokerName,
			}
			channel <- msg
		case msg, _ := <-device.DeviceChan.Chan:
			if !strings.HasSuffix(msg.Topic, device.Name) {
				continue
			}

			log.Infof("msg reached to device, %v", msg)
		}
	}
	return nil
}

// DeviceType retunes device type.
func (device DummyDevice) DeviceType() string {
	return "dummy"
}

func (device DummyDevice) Stop() error {
	log.Warnf("closing dummy device: %v", device.Name)
	return nil
}

func (device DummyDevice) AddSubscribe() error {
	if !device.Subscribe {
		return nil
	}
	for _, b := range device.Broker {
		b.AddSubscribed(device.Name, device.QoS)
	}
	return nil
}
