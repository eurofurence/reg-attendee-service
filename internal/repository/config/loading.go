package config

import (
	"flag"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/system"
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
	configurationData = &conf{Logging: loggingConfig{Severity: "DEBUG"}}
	configurationLock = &sync.RWMutex{}
	flag.StringVar(&configurationFilename, "config", "", "config file path")
}

func parseAndOverwriteContext(yamlFile []byte) error {
	newConfigurationData := &conf{}
	err := yaml.Unmarshal(yamlFile, newConfigurationData)
	if err != nil {
		// cannot use logging package here as this would create a circular dependency (logging needs config)
		log.Printf("[00000000] ERROR failed to parse configuration file '%s': %v", configurationFilename, err)
		return err
	}

	// TODO config validation and defaults, e.g. logging severity

	configurationLock.Lock()
	defer configurationLock.Unlock()

	configurationData = newConfigurationData
	return nil
}

func loadConfiguration() error {
	yamlFile, err := ioutil.ReadFile(configurationFilename)
	if err != nil {
		// cannot use logging package here as this would create a circular dependency (logging needs config)
		log.Printf("[00000000] ERROR failed to load configuration file '%s': %v", configurationFilename, err)
		return err
	}
	err = parseAndOverwriteContext(yamlFile)
	return err
}

// use this for tests to set a hardcoded yaml configuration
func LoadTestingConfigurationFromPathOrAbort(configFilenameForTests string) {
	configurationFilename = configFilenameForTests
	StartupLoadConfiguration()
}

func StartupLoadConfiguration() {
	log.Print("[00000000] INFO  Reading configuration...")
	if configurationFilename == "" {
		// cannot use logging package here as this would create a circular dependency (logging needs config)
		log.Print("[00000000] FATAL Configuration file argument missing. Please specify using -config argument. Aborting.")
		system.Exit(1)
	}
	err := loadConfiguration()
	if err != nil {
		// cannot use logging package here as this would create a circular dependency (logging needs config)
		log.Print("[00000000] FATAL Error reading or parsing configuration file. Aborting.")
		system.Exit(1)
	}
}

func Configuration() *conf {
	configurationLock.RLock()
	defer configurationLock.RUnlock()
	return configurationData
}
