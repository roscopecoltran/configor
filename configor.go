package configor

import (
	"fmt"
	"os"
	"regexp"
)

type Configor struct {
	*Config
}

type Config struct {
	Environment string
	ENVPrefix   string
	Debug       bool
	Verbose     bool
}

// New initialize a Configor
func New(config *Config) *Configor {
	if config == nil {
		config = &Config{}
	}

	if os.Getenv("CONFIGOR_DEBUG_MODE") != "" {
		config.Debug = true
	}

	if os.Getenv("CONFIGOR_VERBOSE_MODE") != "" {
		config.Verbose = true
	}

	return &Configor{Config: config}
}

// GetEnvironment get environment
func (configor *Configor) GetEnvironment() string {
	if configor.Environment == "" {
		if env := os.Getenv("CONFIGOR_ENV"); env != "" {
			return env
		}
		if isTest, _ := regexp.MatchString("/_test/", os.Args[0]); isTest {
			return "test"
		}

		return "development"
	}
	return configor.Environment
}

// Load will unmarshal configurations to struct from files that you provide
func (configor *Configor) Load(config interface{}, files ...string) error {
	defer func() {
		if configor.Config.Debug || configor.Config.Verbose {
			fmt.Printf("Configuration:\n  %#v\n", config)
		}
	}()

	for _, file := range configor.getConfigurationFiles(files...) {
		if configor.Config.Debug || configor.Config.Verbose {
			fmt.Printf("Loading configurations from file '%v'...\n", file)
		}
		if err := processFile(config, file); err != nil {
			return err
		}
	}

	prefix := configor.getENVPrefix(config)
	if prefix == "-" {
		return configor.processTags(config)
	}
	return configor.processTags(config, prefix)
}

// nodes := []string{"contacts", "db", "oauth2"}
// configor.Dump(Config, "yaml", "contacts", "db", "oauth2")
func Dump(config interface{}, nodes string, formats string, prefixPath string) error {
	err := os.MkdirAll(prefixPath, 0700)
	if err != nil {
		return err
	}
	if config == nil {
		config = &Config{}
	}
	exportNodes := getAttributesListToExport(nodes)
	// fmt.Println("exportNodes: \n", exportNodes)
	exportFormats := getAttributesListToExport(formats)
	// fmt.Println("exportFormats: \n", exportFormats)
	exportNodesCount := len(exportNodes)
	for _, f := range exportFormats {
		// fmt.Printf("exportNodesCount: %b \n", exportNodesCount)
		switch {
		case exportNodesCount == 0:
			nodeName := "config"
			// fmt.Printf("nodeName: %s, format: %s \n", exportNodesCount, f)
			data, err := encodeFile(config, "config", f)
			if err != nil {
				return err
			}
			filePath := getConfigDumpFilePath(prefixPath, nodeName, f)
			// fmt.Printf("filePath: %s \n", filePath)
			if err := writeFile(filePath, data); err != nil {
				return err
			}
		case exportNodesCount > 0:
			for _, n := range exportNodes {
				// fmt.Printf("nodeName: %s, format: %s \n", n, f)
				data, err := encodeFile(config, n, f)
				if err != nil {
					return err
				}
				filePath := getConfigDumpFilePath(prefixPath, n, f)
				// fmt.Printf("filePath: %s \n", filePath)
				if err := writeFile(filePath, data); err != nil {
					return err
				}
			}
		}
	}
	return nil
	//return errors.New("error occured while selecting the node to export")
}

// ENV return environment
func ENV() string {
	return New(nil).GetEnvironment()
}

// Load will unmarshal configurations to struct from files that you provide
func Load(config interface{}, files ...string) error {
	return New(nil).Load(config, files...)
}
