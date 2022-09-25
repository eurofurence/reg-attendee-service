package config

import (
	"errors"
	"flag"
	"github.com/eurofurence/reg-attendee-service/internal/repository/system"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/url"
	"sort"
	"sync"
)

var (
	configurationData     *conf
	configurationLock     *sync.RWMutex
	configurationFilename string
	dbMigrate             bool
	ecsLogging            bool
)

var (
	ErrorConfigArgumentMissing = errors.New("configuration file argument missing. Please specify using -config argument. Aborting")
	ErrorConfigFile            = errors.New("failed to read or parse configuration file. Aborting")
)

func init() {
	configurationData = &conf{Logging: loggingConfig{Severity: "DEBUG"}}
	configurationLock = &sync.RWMutex{}

	flag.StringVar(&configurationFilename, "config", "", "config file path")
	flag.BoolVar(&dbMigrate, "migrate-database", false, "migrate database on startup")
	flag.BoolVar(&ecsLogging, "ecs-json-logging", false, "switch to structured json logging")
}

// ParseCommandLineFlags is exposed separately so you can skip it for tests
func ParseCommandLineFlags() {
	log.Print("[00000000] INFO  Parsing command line flags...")
	flag.Parse()
}

func parseAndOverwriteConfig(yamlFile []byte) error {
	newConfigurationData := &conf{}
	err := yaml.UnmarshalStrict(yamlFile, newConfigurationData)
	if err != nil {
		// cannot use logging package here as this would create a circular dependency (logging needs config)
		log.Printf("[00000000] ERROR failed to parse configuration file '%s': %v", configurationFilename, err)
		return err
	}

	setConfigurationDefaults(newConfigurationData)

	errs := url.Values{}
	validateServerConfiguration(errs, newConfigurationData.Server)
	validateLoggingConfiguration(errs, newConfigurationData.Logging)
	validateSecurityConfiguration(errs, newConfigurationData.Security)
	validateDatabaseConfiguration(errs, newConfigurationData.Database)
	validateFlagsConfiguration(errs, newConfigurationData.Choices.Flags)
	validatePackagesConfiguration(errs, newConfigurationData.Choices.Packages)
	validateOptionsConfiguration(errs, newConfigurationData.Choices.Options)
	validateBirthdayConfiguration(errs, newConfigurationData.Birthday)
	validateRegistrationStartTime(errs, newConfigurationData.GoLive)

	if len(errs) != 0 {
		var keys []string
		for key, _ := range errs {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, k := range keys {
			key := k
			val := errs[k]
			// cannot use logging package here as this would create a circular dependency (logging needs config)
			log.Printf("[00000000] ERROR configuration error: %s: %s", key, val[0])
		}
		return errors.New("configuration validation error")
	}

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
	err = parseAndOverwriteConfig(yamlFile)
	return err
}

// use this for tests to set a hardcoded yaml configuration
func LoadTestingConfigurationFromPathOrAbort(configFilenameForTests string) {
	configurationFilename = configFilenameForTests
	if err := StartupLoadConfiguration(); err != nil {
		system.Exit(1)
	}
}

// use this for tests
func EnableTestingMigrateDatabase() {
	dbMigrate = true
}

func StartupLoadConfiguration() error {
	log.Print("[00000000] INFO  Reading configuration...")
	if configurationFilename == "" {
		// cannot use logging package here as this would create a circular dependency (logging needs config)
		log.Print("[00000000] FATAL Configuration file argument missing. Please specify using -config argument. Aborting.")
		return ErrorConfigArgumentMissing
	}
	err := loadConfiguration()
	if err != nil {
		// cannot use logging package here as this would create a circular dependency (logging needs config)
		log.Print("[00000000] FATAL Error reading or parsing configuration file. Aborting.")
		return ErrorConfigFile
	}
	return nil
}

func Configuration() *conf {
	configurationLock.RLock()
	defer configurationLock.RUnlock()
	return configurationData
}
