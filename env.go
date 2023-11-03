package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

type configType string

const (
	ConfigTypeTOML configType = "toml"
	ConfigTypeYAML configType = "yaml"
	ConfigTypeJSON configType = "json"
)

type ConfigMap struct {
	Key   string
	Value any
}

type ConfigManager struct {
	configName       string
	configFileUsed   string
	configType       configType
	envPrefix        string
	mapConfig        map[string]ConfigMap
	defaultConfig    map[string]ConfigMap
	envConfig        map[string]ConfigMap
	combinedConfig   map[string]ConfigMap
	mutex            sync.RWMutex
	explicitDefaults bool
}

var ErrConfigFileNotFound = errors.New("config File Not Found")

func NewConfigManager(automaticEnv bool) *ConfigManager {
	cm := ConfigManager{}
	cm.envConfig = make(map[string]ConfigMap)
	cm.mapConfig = make(map[string]ConfigMap)
	cm.defaultConfig = make(map[string]ConfigMap)
	cm.combinedConfig = make(map[string]ConfigMap)
	cm.envPrefix = ""
	envSet := os.Environ()
	for _, env := range envSet {
		kv := strings.Split(env, "=")
		lower := strings.ToLower(kv[0])
		cm.envConfig[lower] = ConfigMap{Key: kv[0], Value: kv[1]}
	}
	return &cm
}

func (c *ConfigManager) ConfigFileUsed() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.configFileUsed
}

func (c *ConfigManager) UseExplicitDefaults(enable bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.explicitDefaults = enable
}

func (c *ConfigManager) collapse() {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	ccm := make(map[string]ConfigMap)
	for k, v := range c.defaultConfig {
		ccm[k] = v
		if _, ok := c.envConfig[k]; ok {
			ccm[k] = c.envConfig[k]
		}
	}
	for k, v := range c.mapConfig {
		ccm[k] = v
	}
	c.combinedConfig = ccm
}

func (c *ConfigManager) WriteConfig() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	switch c.configType {
	case ConfigTypeTOML:
		f, err := os.Create(c.configFileUsed)
		if err != nil {
			return err
		}
		defer f.Close()
		enc := toml.NewEncoder(f)
		err = enc.Encode(c.combinedConfig)
		return err
	case ConfigTypeYAML:
		f, err := os.Create(c.configFileUsed)
		if err != nil {
			return err
		}
		defer f.Close()
		enc := yaml.NewEncoder(f)
		err = enc.Encode(c.combinedConfig)
		return err
	case ConfigTypeJSON:
		f, err := os.Create(c.configFileUsed)
		if err != nil {
			return err
		}
		defer f.Close()
		enc := json.NewEncoder(f)
		return enc.Encode(c.combinedConfig)
	default:
		return fmt.Errorf("config type %s not supported", c.configType)
	}
}

func (c *ConfigManager) GetBool(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if v, ok := c.combinedConfig[strings.ToLower(key)]; ok {
		val := v.Value
		switch val := val.(type) {
		case bool:
			return val
		case string:
			if strings.ToLower(val) == "true" {
				return true
			}
			return false
		case int:
			if val == 0 {
				return false
			}
			return true
		case float32, float64:
			if val == 0 {
				return false
			}
			return true
		case nil:
			return false
		case time.Duration:
			if val == 0 || val < 0 {
				return false
			}
			return true
		default:
			return val.(bool)
		}
	}
	return false
}

func (c *ConfigManager) GetDuration(key string) time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if v, ok := c.combinedConfig[strings.ToLower(key)]; ok {
		val := v.Value
		switch val := val.(type) {
		case time.Duration:
			return val
		case string:
			d, err := time.ParseDuration(val)
			if err != nil {
				return 0
			}
			return d
		case int:
			return time.Duration(val)
		case float32:
			return time.Duration(val)
		case float64:
			return time.Duration(val)
		case nil:
			return 0
		default:
			return val.(time.Duration)
		}
	}
	return 0
}

func (c *ConfigManager) GetString(key string) string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if v, ok := c.combinedConfig[strings.ToLower(key)]; ok {
		switch val := v.Value.(type) {
		case string:
			return val
		default:
			return fmt.Sprintf("%v", v.Value)
		}
	}
	return ""
}

func (c *ConfigManager) GetStringMap(key string) map[string]any {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if v, ok := c.combinedConfig[strings.ToLower(key)]; ok {
		switch val := v.Value.(type) {
		case map[string]any:
			return val
		default:
			return nil
		}
	}
	return nil
}

func (c *ConfigManager) GetStringSlice(key string) []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if v, ok := c.combinedConfig[strings.ToLower(key)]; ok {
		switch val := v.Value.(type) {
		case []string:
			return val
		default:
			return nil
		}
	}
	return nil
}

func (c *ConfigManager) GetInt(key string) int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if v, ok := c.combinedConfig[strings.ToLower(key)]; ok {
		switch val := v.Value.(type) {
		case int:
			return val
		case string:
			i, err := strconv.Atoi(val)
			if err != nil {
				return 0
			}
			return i
		case float32:
			return int(val)
		case float64:
			return int(val)
		case nil:
			return 0
		default:
			return val.(int)
		}
	}
	return 0
}

func (c *ConfigManager) SetConfigType(configType string) error {
	switch configType {
	case "toml":
		c.configType = ConfigTypeTOML
	case "yaml":
		c.configType = ConfigTypeYAML
	case "json":
		c.configType = ConfigTypeJSON
	default:
		return fmt.Errorf("config type %s not supported", configType)
	}
	return nil
}

func (c *ConfigManager) SetEnvPrefix(prefix string) {
	c.envPrefix = prefix
}

func (c *ConfigManager) GetIntSlice(key string) []int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if v, ok := c.combinedConfig[strings.ToLower(key)]; ok {
		switch val := v.Value.(type) {
		case []int:
			return val
		default:
			return nil
		}
	}
	return nil
}

func (c *ConfigManager) ReadInConfig() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	// assume config = map[string]any
	confFileData, err := readFile(c.configFileUsed, c.configType)
	if err != nil {
		return err
	}
	conf := make(map[string]ConfigMap)
	for k, v := range confFileData {
		lower := strings.ToLower(k)
		conf[lower] = ConfigMap{Key: k, Value: v}
	}
	c.mapConfig = conf
	return nil
}

func readFile(filename string, fileType configType) (map[string]any, error) {
	fileData := make(map[string]any)
	switch fileType {
	case ConfigTypeTOML:
		_, err := toml.DecodeFile(filename, &fileData)
		return fileData, err
	case ConfigTypeYAML:
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		d := yaml.NewDecoder(f)
		err = d.Decode(&fileData)
		return fileData, err
	case ConfigTypeJSON:
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		err = json.NewDecoder(f).Decode(&fileData)
		return fileData, err
	default:
		return nil, fmt.Errorf("config type %s not supported", fileType)
	}
}

func (c *ConfigManager) SetConfigName(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.configName = name
}

func (c *ConfigManager) SetConfigFile(file string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.configFileUsed = file
}
