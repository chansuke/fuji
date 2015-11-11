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
	"io"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	serial "github.com/tarm/serial"
	validator "gopkg.in/validator.v2"

	"github.com/shiguredo/fuji/broker"
	"github.com/shiguredo/fuji/config"
	"github.com/shiguredo/fuji/message"
)

type SerialDevice struct {
	Name       string `validate:"max=256,regexp=[^/]+,validtopic"`
	Broker     []*broker.Broker
	BrokerName string
	QoS        byte `validate:"min=0,max=2"`
	InputPort  InputPortType
	Serial     string `validate:"max=256"`
	Baud       int    `validate:"min=0"`
	Size       int    `validate:"min=0,max=256"`
	Type       string `validate:"max=256"`
	Interval   int    `validate:"min=0"`
	Retain     bool
	Subscribe  bool
	DeviceChan DeviceChannel // GW -> device
}

func (device SerialDevice) String() string {
	var brokers []string
	for _, broker := range device.Broker {
		brokers = append(brokers, fmt.Sprintf("%s\n", broker))
	}
	return fmt.Sprintf("%#v", device)
}

// NewSerialDevice read config.ConfigSection and returnes SerialDevice.
// If config validation failed, return error
func NewSerialDevice(section config.ConfigSection, brokers []*broker.Broker, devChan DeviceChannel) (SerialDevice, error) {
	ret := SerialDevice{
		Name:       section.Name,
		DeviceChan: devChan,
		Interval:   1,
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
		return ret, err
	} else {
		ret.QoS = byte(qos)
	}
	// TODO: check it is true or not
	// ret.InputPort = InputPortType(INPUT_PORT_SERIAL)
	ret.InputPort = InputPortType(INPUT_PORT_DUMMY)
	ret.Serial = values["serial"]
	baud, err := strconv.Atoi(values["baud"])
	if err != nil {
		return ret, err
	} else {
		ret.Baud = int(baud)
	}
	if values["size"] == "" {
		ret.Size = 0
	} else {
		sizev, err := strconv.Atoi(values["size"])
		if err != nil {
			return ret, err
		} else {
			ret.Size = int(sizev)
		}
	}
	ret.Type = values["type"]
	ret.Retain = false
	if values["retain"] == "true" {
		ret.Retain = true
	}

	sub, ok := values["subscribe"]
	if ok && sub == "true" {
		ret.Subscribe = true
	}

	if err := ret.Validate(); err != nil {
		return ret, err
	}

	return ret, nil
}

func (device *SerialDevice) Validate() error {
	validator := validator.NewValidator()
	validator.SetValidationFunc("validtopic", config.ValidMqttPublishTopic)
	if err := validator.Validate(device); err != nil {
		return err
	}
	return nil
}

func readSizedSerialPortLoop(bufSize int, port *serial.Port, readpipe chan []byte) error {
	readBuf := make([]byte, 512)
	var sumBuf = []byte{}
	var renewBuf = []byte{}
	sendBuf := make([]byte, 256)

	for {
		num, err := port.Read(readBuf)
		if err == io.EOF {
			continue
		}
		if err != nil {
			return fmt.Errorf("cannnot open serial port: serialPort: %v, Error: %v", port, err)
		}
		if num > 0 {
			log.Debugf("readBuf: %v, len: %v", readBuf, len(readBuf))
			for index := range readBuf[:num] {
				sumBuf = append(sumBuf, readBuf[index])
			}
			for len(sumBuf) >= bufSize {
				sendBuf = sumBuf[:bufSize]
				readpipe <- sendBuf

				// Truncate sumBuf by Size
				log.Debugf("sumBuf: %v, len: %v", sumBuf, len(sumBuf))
				renewBuf = []byte{}
				for index := bufSize; index < len(sumBuf); index++ {
					renewBuf = append(renewBuf, sumBuf[index])
				}
				sumBuf = renewBuf
				log.Debugf("renewed sumBuf: %v, len: %v / Size: %v", sumBuf, len(sumBuf), bufSize)
			}
		}
	}
}

func readFreesizedSerialPortLoop(port *serial.Port, readpipe chan []byte) error {
	readBuf := make([]byte, 256)
	var sumBuf = []byte{}

	readPointer := 0

	defer port.Close()

	for {
		num, err := port.Read(readBuf)
		if err == io.EOF {
			// No more data comes
			if readPointer > 0 {
				log.Debugf("read data to send: %v", sumBuf)
				readpipe <- sumBuf
				readPointer = 0
				sumBuf = []byte{}
				log.Debugf("sumBuf cleared: %v", sumBuf)
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("cannnot open serial port: serialPort: %v, Error: %v", port, err)
		}
		if num > 0 {
			readPointer += num
			for index := range readBuf[:num] {
				sumBuf = append(sumBuf, readBuf[index])
			}
			log.Debugf("read partial data: %v to sumBuf: %v", readBuf, sumBuf)
			continue
		}
	}
}

func (device SerialDevice) Start(channel chan message.Message) error {
	serialConfig := &serial.Config{Name: device.Serial, Baud: device.Baud, ReadTimeout: time.Millisecond * 50}
	serialPort, err := serial.OpenPort(serialConfig)
	if err != nil {
		return fmt.Errorf("serial device start failed, serialConfig: %v, serialPort: %v, Error: %v", serialConfig, serialPort, err)
	}

	readPipe := make(chan []byte)

	if device.Size > 0 {
		go readSizedSerialPortLoop(device.Size, serialPort, readPipe)
	} else {
		go readFreesizedSerialPortLoop(serialPort, readPipe)
	}

	log.Info("start serial device")

	writeBuf := make([]byte, 256)
	msgBuf := make([]byte, 256)

	go func() error {
		for {
			select {
			case msgBuf = <-readPipe:
				log.Debugf("msgBuf to send: %v", msgBuf)
				msg := message.Message{
					Sender:     device.Name,
					Type:       device.Type,
					QoS:        device.QoS,
					Retained:   device.Retain,
					BrokerName: device.BrokerName,
					Body:       msgBuf,
				}
				channel <- msg
			case msg, _ := <-device.DeviceChan.Chan:
				log.Infof("msg topic:, %v / %v", msg.Topic, device.Name)
				if !strings.HasSuffix(msg.Topic, device.Name) {
					continue
				}
				log.Infof("msg reached to device, %v", msg)
				writeBuf = msg.Body
				num, err := serialPort.Write(writeBuf)
				if err != nil {
					log.Error(err)
					return err
				}
				log.Infof("written length: %d", num)
			}
		}
	}()
	return nil
}

func (device SerialDevice) Stop() error {
	log.Infof("closing serial: %v", device.Name)
	return nil
}

func (device SerialDevice) DeviceType() string {
	return "serial"
}

func (device SerialDevice) AddSubscribe() error {
	if !device.Subscribe {
		return nil
	}
	for _, b := range device.Broker {
		b.AddSubscribed(device.Name, device.QoS)
	}
	return nil
}
