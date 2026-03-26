package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	Server ServerConfig `mapstructure:"server" validate:"required"`
}

type ServerConfig struct {
	Port int `mapstructure:"port" validate:"required,min=1,max=65535"`
}

func TestLoad(t *testing.T) {
	// 需要配置文件支持
	cfg, err := Load[TestConfig](
		WithName("test-config"),
		WithType("yaml"),
		WithPaths("./testdata"),
		WithValidate(true),
	)
	assert.NoError(t, err)
	assert.Equal(t, 8080, cfg.Server.Port)
}

func TestMustLoad(t *testing.T) {
	cfg := MustLoad[TestConfig](
		WithName("test-config"),
		WithType("yaml"),
		WithPaths("./testdata"),
	)
	assert.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Server.Port)
}
