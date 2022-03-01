package stl

import (
	"os"
	"regexp"
)

// XPConfigImpl 配置工具对象
type XPConfigImpl struct {
	*XPConfigEnvironment
}

// XPConfigEnvironment 配置环境对象
type XPConfigEnvironment struct {
	Environment       string
	EnvironmentPrefix string
}

// NewXPConfig 创建配置工具对象
func NewXPConfig(configEnv *XPConfigEnvironment) *XPConfigImpl {
	if configEnv == nil {
		configEnv = &XPConfigEnvironment{}
	}
	return &XPConfigImpl{XPConfigEnvironment: configEnv}
}

// GetEnvironment 获取配置的环境
func (instance *XPConfigImpl) GetEnvironment() string {
	if instance.Environment == "" {
		if env := os.Getenv("XPCONFIG_ENV"); env != "" {
			return env
		}

		if isTest, _ := regexp.MatchString("/_test/", os.Args[0]); isTest {
			return "test"
		}

		return "development"
	}

	return instance.Environment
}

// Load 从文件加载配置
func (instance *XPConfigImpl) Load(config interface{}, files ...string) error {
	for _, file := range instance.getConfigurationFiles(files...) {
		if err := processFile(config, file); err != nil {
			return err
		}
	}

	if prefix := instance.getEnvironmentPrefix(config); prefix == "-" {
		return processTags(config)
	} else {
		return processTags(config, prefix)
	}
}
