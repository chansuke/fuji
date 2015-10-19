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

// broker is an package about define MQTT connecion.
package broker

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	log "github.com/Sirupsen/logrus"
	validator "gopkg.in/validator.v2"

	"github.com/shiguredo/fuji/inidef"
	"github.com/shiguredo/fuji/message"
	"github.com/shiguredo/fuji/utils"
)

type Broker struct {
	GatewayName   string
	Name          string `validate:"max=256,regexp=[^/]+,validtopic"`
	Priority      int    `validate:"min=1,max=3"`
	Host          string `validate:"max=256"`
	Port          int    `validate:"min=1,max=65535"`
	Username      string `validate:"max=256"`
	Password      string `validate:"max=256"`
	RetryInterval int    `validate:"min=0"`
	TopicPrefix   string `validate:"max=256"`
	WillMessage   []byte `validate:"max=256"`
	WillTopic     string `validate:"max=256,validtopic"`
	Tls           bool
	CaCert        string `validate:"max=256"`
	TLSConfig     *tls.Config
	Subscribed    Subscribed // list of subscribed topics

	GwChan chan message.Message

	MQTTClient *MQTT.Client
	connected  bool
}

func (broker *Broker) String() string {
	return fmt.Sprintf("%#v", broker)
}

type Brokers []*Broker

func (bs Brokers) Len() int {
	return len(bs)
}

func (bs Brokers) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func (bs Brokers) Less(i, j int) bool {
	return bs[i].Priority < bs[j].Priority
}

// init is automatically invoked at initial time.
func init() {
	validator.SetValidationFunc("validtopic", inidef.ValidMqttPublishTopic)
}

// NewTLSConfig returns TLS config from CA Cert file path.
func NewTLSConfig(caCertFilePath string) (*tls.Config, error) {
	certPool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile(caCertFilePath)
	if err != nil {
		return nil, inidef.Error("Cert File could not be read.")
	}
	appendCertOk := certPool.AppendCertsFromPEM(pemCerts)
	if appendCertOk != true {
		return nil, inidef.Error("Server Certificate parse failed")
	}

	// only server certificate checked
	return &tls.Config{
		RootCAs:    certPool,
		ClientAuth: tls.NoClientCert,
		ClientCAs:  nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: true,
	}, nil
}

// NewBrokers returns []*Broker from inidef.Config.
// If validation failes, retrun error.
func NewBrokers(conf inidef.Config, gwChan chan message.Message) (Brokers, error) {
	var brokers Brokers

	for _, section := range conf.Sections {
		if section.Type != "broker" {
			continue
		}
		values := section.Values

		willMsg, err := utils.ParsePayload(values["will_message"])
		if err != nil {
			log.Warnf("will_message, %v", err)
		}
		broker := &Broker{
			GatewayName:   conf.GatewayName,
			Name:          section.Name,
			Host:          values["host"],
			Username:      values["username"],
			Password:      values["password"],
			TopicPrefix:   values["topic_prefix"],
			WillMessage:   willMsg,
			Tls:           false,
			CaCert:        "",
			RetryInterval: int(0),
			Subscribed:    NewSubscribed(),
			GwChan:        gwChan,
		}
		priority := 1
		if section.Arg != "" {
			priority, err = strconv.Atoi(section.Arg)
			if err != nil {
				return nil, fmt.Errorf("broker priority parse failed, %v", section.Arg)
			}
		}
		broker.Priority = int(priority)

		port, err := strconv.Atoi(values["port"])
		if err != nil {
			return nil, fmt.Errorf("broker port parse failed, %v", values["port"])
		}
		broker.Port = int(port)

		// OPTIONAL fields
		if values["retry_interval"] != "" {
			retry_interval, err := strconv.Atoi(values["retry_interval"])
			if err != nil {
				return nil, err
			} else {
				broker.RetryInterval = int(retry_interval)
			}
		}

		if values["tls"] == "true" && values["cacert"] != "" {
			// validate certificate
			broker.Tls = true
			broker.CaCert = values["cacert"]
			broker.TLSConfig, err = NewTLSConfig(broker.CaCert)
			if err != nil {
				return nil, err
			}
		}

		if values["will_topic"] != "" {
			broker.WillTopic = strings.Join([]string{broker.TopicPrefix, values["will_topic"]}, "/")
		} else {
			broker.WillTopic = strings.Join([]string{broker.TopicPrefix, broker.GatewayName, "will"}, "/")
		}
		// Validation
		if err := validator.Validate(broker); err != nil {
			return brokers, err
		}
		brokers = append(brokers, broker)
	}

	// sort by Priority
	sort.Sort(brokers)

	return brokers, nil
}

func (b *Broker) IsConnected() bool {
	if b.MQTTClient != nil && b.MQTTClient.IsConnected() && b.connected {
		return true
	}
	return false
}

func (b *Broker) onConnectionLost(client *MQTT.Client, reason error) {
	log.Errorf("MQTT broker disconnected(%s): %s", b.Name, reason)
	b.connected = false
}

func (b *Broker) onMessageReceived(client *MQTT.Client, m MQTT.Message) {
	log.Debugf("topic:%s / msg:%s", m.Topic(), m.Payload())

	msg := message.Message{
		Sender: b.Name,
		Type:   message.TypeSubscribed,
		Body:   m.Payload(),
		Topic:  m.Topic(),
	}
	b.GwChan <- msg
}

func (b *Broker) SubscribeOnConnect(client *MQTT.Client) {
	log.Infof("client connected")
	b.connected = true

	if b.Subscribed.Length() > 0 {
		// subscribe
		token := client.SubscribeMultiple(b.Subscribed.List(), b.onMessageReceived)
		token.Wait()
		if token.Error() != nil {
			log.Error(token.Error())
		}
	}
}

// MQTTClientSetup setup MQTTOptions and connect ot broker.
func (b *Broker) MQTTClientSetup(gwName string) error {
	cli, err := MQTTConnect(gwName, b)
	if err != nil {
		return err
	}

	if token := cli.Connect(); token.Wait() && token.Error() != nil {
		log.Errorf("Failed to start MQTT client: %v", token.Error())
		return token.Error()
	}

	b.MQTTClient = cli
	return nil
}

func (b *Broker) Publish(msg *message.Message) error {
	if b.MQTTClient == nil || !b.IsConnected() {
		log.Warn("message got but Broker not connected")
		return nil
	}

	topic, err := b.GenerateTopic(msg)
	if err != nil {
		return err
	}

	log.Debugf("message got: %v", topic)
	token := b.MQTTClient.Publish(topic.Str, msg.QoS, msg.Retained, msg.Body)
	log.Debugf("message published: %v", topic)
	token.Wait()
	if token.Error() != nil {
		log.Errorf("Failed to publish: %v", token.Error())
		return token.Error()
	}

	return nil
}

// GenerateTopic generates topic from topicprefix, gwname and message.
func (b *Broker) GenerateTopic(msg *message.Message) (message.TopicString, error) {
	var topicString string
	switch msg.Sender {
	case "status": // status device topic structure is difference
		topicString = strings.Join([]string{b.TopicPrefix, msg.Topic}, "/")
	default:
		topicString = strings.Join([]string{b.TopicPrefix, b.GatewayName, msg.Sender, msg.Type}, "/")
	}

	topic := message.TopicString{
		Str: topicString,
	}
	if err := topic.Validate(); err != nil {
		log.Errorf("topic validation error, %v", err)
		return topic, err
	}

	return topic, nil
}

func (b *Broker) Close() error {
	if b.MQTTClient != nil {
		b.MQTTClient.Disconnect(250) // msec wait
	}
	return nil
}

func (b *Broker) FourceClose() error {
	if b.MQTTClient != nil {
		b.MQTTClient.Disconnect(5) // msec wait
	}
	return nil
}

// MQTTConnect returns MQTTClient with options.
func MQTTConnect(gwName string, b *Broker) (*MQTT.Client, error) {
	opts := MQTT.NewClientOptions()

	defaulturl := fmt.Sprintf("tcp://%s:%d", b.Host, b.Port)
	if b.Tls {
		defaulturl := fmt.Sprintf("ssl://%s:%d", b.Host, b.Port)
		opts.AddBroker(defaulturl)
		opts.SetClientID(gwName)
		opts.SetTLSConfig(b.TLSConfig)
	} else {
		opts.AddBroker(defaulturl)
		opts.SetClientID(gwName)
	}
	log.Infof("broker connecting to: %v", defaulturl)

	opts.SetUsername(b.Username)
	opts.SetPassword(b.Password)
	if !inidef.IsNil(b.WillMessage) {
		willTopic := b.WillTopic
		willQoS := 0
		opts.SetBinaryWill(willTopic, b.WillMessage, byte(willQoS), true)
	}
	opts.SetOnConnectHandler(b.SubscribeOnConnect)
	opts.SetConnectionLostHandler(b.onConnectionLost)

	client := MQTT.NewClient(opts)
	return client, nil
}

func GetBrokerNames(brokers []*Broker) []string {
	var ret []string
	for _, b := range brokers {
		ret = append(ret, b.Name)
	}
	return ret
}
