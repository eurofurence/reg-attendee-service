package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"strings"
)

type Config struct {
	Token            string `yaml:"idp_token"`          // can take from AUTH cookie in regsys for a user that is staff + director
	IDPUrl           string `yaml:"idp_url"`            // base url with no trailing /
	Jwt              string `yaml:"jwt"`                // can take from admin auth in regsys
	Auth             string `yaml:"auth"`               // can take from admin auth in regsys
	RegsysUrl        string `yaml:"regsys_url"`         // base url including context including /attsrv, no trailing /
	StaffGroupID     string `yaml:"staff_group_id"`     // look up in IDP in url
	DirectorsGroupID string `yaml:"directors_group_id"` // look up in IDP in url
}

func (c *Config) validate() error {
	if c.Token == "" {
		return errors.New("identity provider token empty")
	}
	if !strings.HasPrefix(c.IDPUrl, "https://") || strings.HasSuffix(c.IDPUrl, "/") {
		return errors.New("invalid identity provider url")
	}
	if jwtParts := strings.Split(c.Jwt, "."); len(jwtParts) != 3 {
		return errors.New("invalid jwt cookie, must contain full jwt with all 3 parts")
	}
	if c.Auth == "" {
		return errors.New("invalid auth cookie")
	}
	if !strings.HasPrefix(c.RegsysUrl, "http") || strings.HasSuffix(c.RegsysUrl, "/") {
		return errors.New("invalid regsys base url")
	}
	if c.StaffGroupID == "" {
		return errors.New("staff group id missing")
	}
	if c.DirectorsGroupID == "" {
		return errors.New("directors group id missing")
	}
	return nil
}

func loadValidatedConfig() (Config, error) {
	log.Println("reading configuration")

	result := Config{}

	yamlFile, err := os.ReadFile("cmd/staffbadges/config.yaml")
	if err != nil {
		return result, fmt.Errorf("failed to load config.yaml: %s", err.Error())
	}

	if err := yaml.UnmarshalStrict(yamlFile, &result); err != nil {
		return result, fmt.Errorf("failed to parse config.yaml: %s", err.Error())
	}

	if err := result.validate(); err != nil {
		return result, fmt.Errorf("failed to validate configuration: %s", err.Error())
	}

	log.Println("successfully read and validated configuration")
	return result, nil
}
