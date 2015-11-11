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

package gateway

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shiguredo/fuji/config"
)

func TestNewGateway(t *testing.T) {
	assert := assert.New(t)

	conf, err := config.LoadConfig("../tests/testing_conf.toml")
	gw, err := NewGateway(conf)
	assert.Nil(err)
	assert.Equal("ham", gw.Name)
	assert.NotNil(gw.CmdChan)
	assert.NotNil(gw.MsgChan)
	assert.NotNil(gw.BrokerChan)
}

func TestNewGatewayInvalidName(t *testing.T) {
	assert := assert.New(t)

	{ // includes plus
		configStr := `
[gateway]
name = "bone+lessham"
`
		conf, err := config.LoadConfigByte([]byte(configStr))
		_, err = NewGateway(conf)
		assert.NotNil(err)
	}
	{ // includes sharp
		configStr := `
[gateway]
name = "bone#lessham"`
		conf, err := config.LoadConfigByte([]byte(configStr))
		_, err = NewGateway(conf)
		assert.NotNil(err)
	}
	{ // too long
		configStr := `
[gateway]
name = "bonelesshaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
`
		conf, err := config.LoadConfigByte([]byte(configStr))
		_, err = NewGateway(conf)
		assert.NotNil(err)
	}
	{ // \\U0000 string
		configStr := fmt.Sprintf(`
[gateway]
name = 	"na%cme"
`, '\u0000')
		conf, err := config.LoadConfigByte([]byte(configStr))
		_, err = NewGateway(conf)
		assert.NotNil(err)
	}
}

func TestNewGatewayMaxRetryCount(t *testing.T) {
	assert := assert.New(t)

	{ // default
		configStr := `
[gateway]
name = "sango"
`
		conf, err := config.LoadConfigByte([]byte(configStr))
		gw, err := NewGateway(conf)
		assert.Nil(err)
		assert.Equal(3, gw.MaxRetryCount)
	}
	{ // specified
		configStr := `
[gateway]
name = "sango"
max_retry_count = 10
`
		conf, err := config.LoadConfigByte([]byte(configStr))
		gw, err := NewGateway(conf)
		assert.Nil(err)
		assert.Equal(10, gw.MaxRetryCount)
	}
	{ // minus fail validation
		configStr := `
[gateway]
name = "sango"
max_retry_count = -10
`
		conf, err := config.LoadConfigByte([]byte(configStr))
		_, err = NewGateway(conf)
		assert.NotNil(err)
	}
	{ // invalid int
		configStr := `
[gateway]
name = "sango"
max_retry_count = aabbcc
`
		conf, err := config.LoadConfigByte([]byte(configStr))
		_, err = NewGateway(conf)
		assert.NotNil(err)
	}
}

func TestNewGatewayRetryInterval(t *testing.T) {
	assert := assert.New(t)

	{ // default
		configStr := `
[gateway]
name = "sango"
`
		conf, err := config.LoadConfigByte([]byte(configStr))
		gw, err := NewGateway(conf)
		assert.Nil(err)
		assert.Equal(3, gw.RetryInterval)
	}
	{ // specified
		configStr := `
[gateway]
name = "sango"
retry_interval = 10
`
		conf, err := config.LoadConfigByte([]byte(configStr))
		gw, err := NewGateway(conf)
		assert.Nil(err)
		assert.Equal(10, gw.RetryInterval)
	}
	{ // minus fail validation
		configStr := `
[gateway]
name = "sango"
retry_interval = -10
`
		conf, err := config.LoadConfigByte([]byte(configStr))
		_, err = NewGateway(conf)
		assert.NotNil(err)
	}
	{ // invalid int
		configStr := `
[gateway]
name = "sango"
retry_interval = aabbcc
`
		conf, err := config.LoadConfigByte([]byte(configStr))
		_, err = NewGateway(conf)
		assert.NotNil(err)
	}
}
