package configor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	// "github.com/fsamin/go-dump"
	"github.com/joho/godotenv"

	// "github.com/k0kubun/pp"
	yaml "gopkg.in/yaml.v2"
)

var defaultEnvFiles []string = []string{".env"}
var EnvKeys map[string]string

func (configor *Configor) getENVPrefix(config interface{}) string {
	if configor.Config.ENVPrefix == "" {
		if prefix := os.Getenv("CONFIGOR_ENV_PREFIX"); prefix != "" {
			return prefix
		}
		return "Configor"
	}
	return configor.Config.ENVPrefix
}

func getConfigurationFileWithENVPrefix(file, env string) (string, error) {
	var (
		envFile string
		extname = path.Ext(file)
	)

	if extname == "" {
		envFile = fmt.Sprintf("%v.%v", file, env)
	} else {
		envFile = fmt.Sprintf("%v.%v%v", strings.TrimSuffix(file, extname), env, extname)
	}

	if fileInfo, err := os.Stat(envFile); err == nil && fileInfo.Mode().IsRegular() {
		return envFile, nil
	}
	return "", fmt.Errorf("failed to find file %v", file)
}

func (configor *Configor) getConfigurationFiles(files ...string) []string {
	var results []string

	for i := len(files) - 1; i >= 0; i-- {
		foundFile := false
		file := files[i]

		// check configuration
		if fileInfo, err := os.Stat(file); err == nil && fileInfo.Mode().IsRegular() {
			foundFile = true
			results = append(results, file)
		}

		// check configuration with env
		if file, err := getConfigurationFileWithENVPrefix(file, configor.GetEnvironment()); err == nil {
			foundFile = true
			results = append(results, file)
		}

		// check example configuration
		if !foundFile {
			if example, err := getConfigurationFileWithENVPrefix(file, "example"); err == nil {
				fmt.Printf("Failed to find configuration %v, using example file %v\n", file, example)
				results = append(results, example)
			} else {
				fmt.Printf("Failed to find configuration %v\n", file)
			}
		}
	}
	return results
}

func processFile(config interface{}, file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	if err := godotenv.Load(); err != nil {
		return err
	}
	// replace KeyHolders before loading the YAML/JSON/TOML file
	EnvKeys, _ = godotenv.Read(defaultEnvFiles...)
	dataStr := string(data)
	for k, v := range EnvKeys {
		holderKey := fmt.Sprintf("{ENV.%s}", strings.Replace(k, "\"", "", -1))
		dataStr = strings.Replace(dataStr, holderKey, v, -1)
	}
	data = []byte(dataStr)
	switch {
	case strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml"):
		return yaml.Unmarshal(data, config)
	case strings.HasSuffix(file, ".toml"):
		return toml.Unmarshal(data, config)
	case strings.HasSuffix(file, ".json"):
		return json.Unmarshal(data, config)
	default:
		if toml.Unmarshal(data, config) != nil {
			if json.Unmarshal(data, config) != nil {
				if yaml.Unmarshal(data, config) != nil {
					return errors.New("failed to decode config")
				}
			}
		}
		return nil
	}
}

// write env file (to do later !)
// env, err := godotenv.Unmarshal("KEY=value")
// err := godotenv.Write(env, "./.env")

func getConfigDumpFilePath(prefixPath string, nodeName string, format string) string {
	return fmt.Sprintf("%s/%s.%s", prefixPath, nodeName, format)
}

func getAttributesListToExport(attrs string) []string {
	return strings.Split(attrs, ",")
}

func getPrefixForStruct(prefixes []string, fieldStruct *reflect.StructField) []string {
	if fieldStruct.Anonymous && fieldStruct.Tag.Get("anonymous") == "true" {
		return prefixes
	}
	return append(prefixes, fieldStruct.Name)
}

func writeFile(filePath string, data []byte) error {
	if err := ioutil.WriteFile(filePath, data, 0600); err != nil {
		return err
	}
	return nil
}

func isEmptyStruct(object interface{}) bool {
	//First check normal definitions of empty
	if object == nil {
		return true
	} else if object == "" {
		return true
	} else if object == false {
		return true
	}
	//Then see if it's a struct
	if reflect.ValueOf(object).Kind() == reflect.Struct {
		// and create an empty copy of the struct object to compare against
		empty := reflect.New(reflect.TypeOf(object)).Elem().Interface()
		if reflect.DeepEqual(object, empty) {
			return true
		}
	}
	return false
}

func encodeFile(config interface{}, node string, format string) ([]byte, error) {
	/*
		ConfigCheck := reflect.Indirect(reflect.ValueOf(config))
		if ConfigCheck.Kind() != reflect.Struct {
			return nil, errors.New("invalid config, should be struct")
		}

		ConfigValue := reflect.ValueOf(config)
		if ConfigValue.Kind() != reflect.Struct {
			return nil, errors.New("invalid config, should be struct")
		}

		out := &bytes.Buffer{}
		fmt.Println("Fdump config:")
		dump.Fdump(out, config)
		fmt.Println(out)

		fmt.Println("Dump ToMap() config:")
		m, _ := dump.ToMap(config, dump.WithDefaultLowerCaseFormatter())
		fmt.Println(m)

		s, _ := dump.Sdump(config)
		fmt.Println("SDump config: \n", s)

		ConfigNode := reflect.Indirect(ConfigValue).FieldByName(node)
	*/
	switch format {
	case "json":
		data, err := json.MarshalIndent(config, "", "\t")
		if err != nil {
			return nil, err
		}
		return data, nil
	case "toml":
		var dataBytes bytes.Buffer
		if err := toml.NewEncoder(&dataBytes).Encode(config); err != nil {
			return nil, err
		}
		// fmt.Println(dataBytes.String())
		return []byte(dataBytes.String()), nil
	case "yaml":
		data, err := yaml.Marshal(config)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	return nil, errors.New("Unkown format to export")

}

func processTags(config interface{}, prefixes ...string) error {
	configValue := reflect.Indirect(reflect.ValueOf(config))
	if configValue.Kind() != reflect.Struct {
		return errors.New("invalid config, should be struct")
	}

	configType := configValue.Type()
	for i := 0; i < configType.NumField(); i++ {
		var (
			envNames    []string
			fieldStruct = configType.Field(i)
			field       = configValue.Field(i)
			envName     = fieldStruct.Tag.Get("env") // read configuration from shell env
		)

		if !field.CanAddr() || !field.CanInterface() {
			continue
		}

		if envName == "" {
			envNames = append(envNames, strings.Join(append(prefixes, fieldStruct.Name), "_"))                  // Configor_DB_Name
			envNames = append(envNames, strings.ToUpper(strings.Join(append(prefixes, fieldStruct.Name), "_"))) // CONFIGOR_DB_NAME
		} else {
			envNames = []string{envName}
		}

		// Load From Shell ENV
		for _, env := range envNames {
			if value := os.Getenv(env); value != "" {
				if err := yaml.Unmarshal([]byte(value), field.Addr().Interface()); err != nil {
					return err
				}
				break
			}
		}

		if isBlank := reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()); isBlank {
			// Set default configuration if blank
			if value := fieldStruct.Tag.Get("default"); value != "" {
				if err := yaml.Unmarshal([]byte(value), field.Addr().Interface()); err != nil {
					return err
				}
			} else if fieldStruct.Tag.Get("required") == "true" {
				// return error if it is required but blank
				return errors.New(fieldStruct.Name + " is required, but blank")
			}
		}

		for field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		if field.Kind() == reflect.Struct {
			if err := processTags(field.Addr().Interface(), getPrefixForStruct(prefixes, &fieldStruct)...); err != nil {
				return err
			}
		}

		if field.Kind() == reflect.Slice {
			for i := 0; i < field.Len(); i++ {
				if reflect.Indirect(field.Index(i)).Kind() == reflect.Struct {
					if err := processTags(field.Index(i).Addr().Interface(), append(getPrefixForStruct(prefixes, &fieldStruct), fmt.Sprint(i))...); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
