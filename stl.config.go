package stl

import (
	"os"
	"regexp"
)

type XPConfigImpl struct {
	*XPConfigEnvironment
}

type XPConfigEnvironment struct {
	Environment       string
	EnvironmentPrefix string
}

func NewXPConfig(configEnv *XPConfigEnvironment) *XPConfigImpl {
	if configEnv == nil {
		configEnv = &XPConfigEnvironment{}
	}
	return &XPConfigImpl{XPConfigEnvironment: configEnv}
}

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
