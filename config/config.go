package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// ConfigLoader 配置加载器
type ConfigLoader struct {
	name        string
	env         string
	configType  string
	paths       []string
	envPrefix   string
	envReplacer *strings.Replacer
	watch       bool
	validate    bool
	onChange    func(ConfigChangeEvent)
}

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	Path string
}

// LoadOption 配置加载选项
type LoadOption func(*ConfigLoader)

// WithName 配置文件名
func WithName(name string) LoadOption {
	return func(l *ConfigLoader) {
		l.name = name
	}
}

// WithEnv 运行环境
func WithEnv(env string) LoadOption {
	return func(l *ConfigLoader) {
		l.env = env
	}
}

// WithType 配置文件类型
func WithType(t string) LoadOption {
	return func(l *ConfigLoader) {
		l.configType = t
	}
}

// WithPaths 配置文件搜索路径
func WithPaths(paths ...string) LoadOption {
	return func(l *ConfigLoader) {
		l.paths = paths
	}
}

// WithEnvPrefix 环境变量前缀
func WithEnvPrefix(prefix string) LoadOption {
	return func(l *ConfigLoader) {
		l.envPrefix = prefix
	}
}

// WithEnvReplacer 环境变量替换规则
func WithEnvReplacer(replacer *strings.Replacer) LoadOption {
	return func(l *ConfigLoader) {
		l.envReplacer = replacer
	}
}

// WithWatch 启用配置热重载
func WithWatch(watch bool) LoadOption {
	return func(l *ConfigLoader) {
		l.watch = watch
	}
}

// WithValidate 启用配置验证
func WithValidate(validate bool) LoadOption {
	return func(l *ConfigLoader) {
		l.validate = validate
	}
}

// WithOnChange 配置变更回调
func WithOnChange(onChange func(ConfigChangeEvent)) LoadOption {
	return func(l *ConfigLoader) {
		l.onChange = onChange
	}
}

// Load 加载配置
func Load[T any](opts ...LoadOption) (*T, error) {
	loader := &ConfigLoader{
		name:       "config",
		env:        os.Getenv("APP_ENV"),
		configType: "yaml",
		paths:      []string{".", "./configs"},
	}

	for _, opt := range opts {
		opt(loader)
	}

	v := viper.New()

	configName := loader.name
	if loader.env != "" {
		configName = fmt.Sprintf("%s.%s", loader.name, loader.env)
	}

	v.SetConfigName(configName)
	v.SetConfigType(loader.configType)

	for _, path := range loader.paths {
		v.AddConfigPath(path)
	}

	if loader.envPrefix != "" {
		v.SetEnvPrefix(loader.envPrefix)
		v.AutomaticEnv()
		if loader.envReplacer != nil {
			v.SetEnvKeyReplacer(loader.envReplacer)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		var configNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configNotFound) {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		return nil, fmt.Errorf("config file not found: %s.%s", configName, loader.configType)
	}

	if loader.watch {
		v.WatchConfig()
		if loader.onChange != nil {
			v.OnConfigChange(func(e fsnotify.Event) {
				loader.onChange(ConfigChangeEvent{Path: e.Name})
			})
		}
	}

	var cfg T
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if loader.validate {
		validate := validator.New()
		if err := validate.Struct(&cfg); err != nil {
			return nil, fmt.Errorf("config validation failed: %w", err)
		}
	}

	return &cfg, nil
}

// MustLoad 加载配置（失败则 panic）
func MustLoad[T any](opts ...LoadOption) *T {
	cfg, err := Load[T](opts...)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

// Prevent unused import error
var _ = time.Second
