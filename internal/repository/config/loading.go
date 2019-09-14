package config

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"sync"
)

var (
	configurationData     *conf
	configurationLock     *sync.RWMutex
	configurationFilename string
)

func init() {
	configurationLock = &sync.RWMutex{}
	flag.StringVar(&configurationFilename, "config", "", "config file path")
}

func parseAndOverwriteContext(yamlFile []byte) error {
	newConfigurationData := &conf{}
	err := yaml.Unmarshal(yamlFile, newConfigurationData)
	if err != nil {
		log.Printf("failed to parse configuration file '%s': %v", configurationFilename, err)
		return err
	}

	configurationLock.Lock()
	defer configurationLock.Unlock()

	configurationData = newConfigurationData
	return nil
}

func LoadConfiguration() error {
	yamlFile, err := ioutil.ReadFile(configurationFilename)
	if err != nil {
		log.Printf("failed to load configuration file '%s': %v", configurationFilename, err)
		return err
	}
	err = parseAndOverwriteContext(yamlFile)
	return err
}

// use this for tests to set a hardcoded yaml configuration
func InitializeConfiguration(yaml string) error {
	return parseAndOverwriteContext([]byte(yaml))
}

func StartupLoadConfiguration() {
	log.Print("Reading configuration...")
	if configurationFilename == "" {
		log.Fatal("Configuration file argument missing. Please specify using -config argument. Aborting.")
	}
	err := LoadConfiguration()
	if err != nil {
		log.Fatal("Error reading or parsing configuration file. Aborting.")
	}
}

func Configuration() *conf {
	configurationLock.RLock()
	defer configurationLock.RUnlock()
	return configurationData
}
