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

// gateway connects Broker-Device
package gateway

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	validator "gopkg.in/validator.v2"

	"github.com/shiguredo/fuji/broker"
	"github.com/shiguredo/fuji/config"
	"github.com/shiguredo/fuji/device"
	"github.com/shiguredo/fuji/message"
)

type Gateway struct {
	Name string `validate:"max=256,regexp=[^/]+,validtopic"`

	Devices []device.Devicer
	Brokers broker.Brokers

	MsgChan        chan message.Message   // Broker -> GW
	BrokerChan     chan message.Message   // GW -> Broker
	CmdChan        chan string            // somewhere -> GW
	DeviceChannels []device.DeviceChannel // GW -> device

	MaxRetryCount int `validate:"min=1"`
	RetryInterval int `validate:"min=1"`
}

const (
	DefaultMaxRetryCount    = 3
	DefaultRetryInterval    = 3 // sec
	MaxMsgChanBufferSize    = 20
	MaxBrokerChanBufferSize = 20
)

func init() {
	validator.SetValidationFunc("validtopic", config.ValidMqttPublishTopic)
}

func (gateway Gateway) String() string {
	return fmt.Sprintf("Name: %s\n", gateway.Name)
}

// NewGateway returns Gateway instance with config object
func NewGateway(conf config.Config) (*Gateway, error) {
	var section config.ConfigSection
	for _, s := range conf.Sections {
		if s.Type == "gateway" {
			section = s
		}
	}
	if section.Type == "" {
		return nil, fmt.Errorf("config does not have gateway")
	}

	gw := Gateway{
		Name:           section.Values["name"],
		MsgChan:        make(chan message.Message, MaxMsgChanBufferSize),
		BrokerChan:     make(chan message.Message, MaxBrokerChanBufferSize),
		DeviceChannels: device.NewDeviceChannels(),
		CmdChan:        make(chan string),
		MaxRetryCount:  DefaultMaxRetryCount,
		RetryInterval:  DefaultRetryInterval,
	}

	if m, ok := section.Values["max_retry_count"]; ok {
		max, err := strconv.Atoi(m)
		if err == nil {
			gw.MaxRetryCount = max
		} else {
			return nil, fmt.Errorf("invalid max_retry_count: %s", m)
		}
	}
	if m, ok := section.Values["retry_interval"]; ok {
		max, err := strconv.Atoi(m)
		if err == nil {
			gw.RetryInterval = max
		} else {
			return nil, fmt.Errorf("invalid retry_interval: %s", m)
		}
	}

	// Validation
	if err := gw.Validate(); err != nil {
		return nil, err
	}
	return &gw, nil
}

func (gw *Gateway) Validate() error {
	return validator.Validate(gw)
}

func (gw *Gateway) Start() error {
	return gw.MainLoop()
}

func (gw *Gateway) Stop() {
	gw.CmdChan <- "close"
}

// Publish pass the message to a Broker which is connected
func (gw *Gateway) Publish(msg message.Message) {
	// Brokers are orderd by Priority
	for _, b := range gw.Brokers {
		if msg.BrokerName != b.Name {
			continue
		}

		for i := 0; i < gw.MaxRetryCount; i++ {
			if b.IsConnected() {
				go b.Publish(&msg)
				return
			}
			time.Sleep(time.Duration(gw.RetryInterval) * time.Second)
		}
	}
	log.Errorf("retry failed. msg discarded: %v, topic: %s", msg.BrokerName)
}

// MainLoop loops forever.
func (gw *Gateway) MainLoop() error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

MAINLOOP:
	for {
		select {
		case msg, ok := <-gw.MsgChan:
			// msgChan: messages from devices
			if !ok {
				log.Error("msg from msgChan closed")
				break MAINLOOP
			}
			// use goroutine to avoid blocking
			go gw.Publish(msg)

		case msg, ok := <-gw.BrokerChan:
			// brokerChan: messages from brokers
			if !ok {
				log.Error("msg from brokerChan closed")
				break MAINLOOP
			}
			if msg.Type != message.TypeSubscribed {
				continue
			}
			// send to all device
			for _, dc := range gw.DeviceChannels {
				dc.Chan <- msg
			}
		case signal, _ := <-sigChan:
			// sigChan: signals
			switch signal {
			case syscall.SIGINT:
				log.Warn("SIGINT caught")
				// deadlock if without "go"
				go gw.Stop()
				continue
			default:
				// do nothing
			}
		case cmd, _ := <-gw.CmdChan:
			// cmdChan: messages from command line or ever
			switch cmd {
			case "close":
				log.Warn("close command comes. will be shutdown")
				for _, b := range gw.Brokers {
					b.Close()
				}
				for _, d := range gw.Devices {
					d.Stop()
				}
				return nil
			default:
				log.Warn("unknown command, %v", cmd)
			}
		}
	}
	return nil
}
