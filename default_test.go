package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

const yamlConfig = `atom: config
int_list:
  - 1
  - 2
  - 3
  - "4"
port: 8080
string_list:
  - "added"
  - "a"
  - "list"
debug: true
duration: 5s
map:
  a: "1"
  b: "other"
otherBool: "true"
otherOtherBool: 1
intDuration: 5
floatDuration: 5.5
stringInt: "5"
stringIntEmpty: ""
`

func TestDefaults(t *testing.T) {
	os.Setenv("envVal", "FAIL")
	temp := t.TempDir()
	globalMap := make(map[string]interface{})
	configYaml := filepath.Join(temp, "config.yaml")
	configToml := filepath.Join(temp, "config.toml")
	configJson := filepath.Join(temp, "config.json")
	err := yaml.Unmarshal([]byte(yamlConfig), &globalMap)
	if err != nil {
		t.Fatal(err)
	}

	tomlFile, err := os.Create(configToml)
	if err != nil {
		t.Fatal(err)
	}
	defer tomlFile.Close()
	err = toml.NewEncoder(tomlFile).Encode(globalMap)
	if err != nil {
		t.Fatal(err)
	}

	jsonFile, err := os.Create(configJson)
	defer jsonFile.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = json.NewEncoder(jsonFile).Encode(globalMap)

	err = os.WriteFile(configYaml, []byte(yamlConfig), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		configType string
		configFile string
	}{
		{"yaml", "yaml", configYaml},
		// {"toml", "toml", configToml},
		{"json", "json", configJson},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetConfigType(tt.configType)
			SetConfigFile(tt.configFile)
			ManagerSubTest(t)
		})
	}
}

func ManagerSubTest(t *testing.T) {
	err := ReadInConfig()
	if err != nil {
		t.Error(err)
	}
	atom := GetString("atom")
	if atom != "config" {
		t.Error("atom should be config")
	}
	intSlice := GetIntSlice("int_list")
	if intSlice[0] != 1 {
		t.Error("intSlice[0] should be 1")
	}
	if intSlice[1] != 2 {
		t.Error("intSlice[1] should be 2")
	}
	if intSlice[2] != 3 {
		t.Error("intSlice[2] should be 3")
	}
	intValue := GetInt("port")
	if intValue != 8080 {
		t.Error("intValue should be 8080")
	}
	boolValue := GetBool("debug")
	if boolValue != true {
		t.Error("boolValue should be true")
	}
	boolNonExistentValue := GetBool("non_existent")
	if boolNonExistentValue != false {
		t.Error("boolNonExistentValue should be false")
	}

	stringList := GetStringSlice("string_list")
	if stringList[0] != "added" {
		t.Error("stringList[0] should be added")
	}
	if stringList[1] != "a" {
		t.Error("stringList[1] should be a")
	}
	if stringList[2] != "list" {
		t.Error("stringList[2] should be list")
	}

	durationValue := GetDuration("duration")
	if durationValue != 5*time.Second {
		t.Error("durationValue should be 5s")
	}

	intDurationValue := GetDuration("intDuration")
	if intDurationValue != 5*time.Nanosecond {
		t.Errorf("intDurationValue should be 5ns but is %v", intDurationValue)
	}

	floatDurationValue := GetDuration("floatDuration")
	if floatDurationValue != 5*time.Nanosecond {
		t.Error("floatDurationValue should be 5ns")
	}

	otherBool := GetBool("otherBool")
	if !otherBool {
		t.Error("otherBool should be true")
	}
	otherOtherBool := GetBool("otherOtherBool")
	if !otherOtherBool {
		t.Error("otherOtherBool should be true")
	}

	mapVal := GetStringMap("map")
	if mapVal["a"] != "1" {
		t.Errorf("mapVal['a'] should be '1' but is %v", mapVal["a"])
	}
	if mapVal["b"] != "other" {
		t.Error("mapVal['b'] should be other")
	}

	Set("stringMap", map[string]any{"a": 1, "b": "other"})
	mapVal = GetStringMap("stringMap")
	if mapVal["a"] != 1 {
		t.Errorf("mapVal['a'] should be 1 but is %v", mapVal["a"])
	}
	if mapVal["b"] != "other" {
		t.Error("mapVal['b'] should be other")
	}

	SetDefault("default", "value")
	defaultValue := GetString("default")
	if defaultValue != "value" {
		t.Error("defaultValue should be value")
	}
	SetDefault("debug", false)
	boolValue = GetBool("debug")
	if !boolValue {
		t.Error("boolValue should be true")
	}

	stringInt := GetInt("stringInt")
	if stringInt != 5 {
		t.Error("stringInt should be 5")
	}

	stringIntEmpty := GetInt("stringIntEmpty")
	if stringIntEmpty != 0 {
		t.Error("stringIntEmpty should be 0")
	}

	Set("set", "value")
	setValue := GetString("set")
	if setValue != "value" {
		t.Error("setValue should be value")
	}

	Set("slice", []string{"a", "b", "c"})
	sliceValue := GetStringSlice("slice")
	if len(sliceValue) == 0 {
		t.Error("sliceValue should not be empty")
	} else {
		if sliceValue[0] != "a" {
			t.Error("sliceValue[0] should be a")
		}
		if sliceValue[1] != "b" {
			t.Error("sliceValue[1] should be b")
		}
		if sliceValue[2] != "c" {
			t.Error("sliceValue[2] should be c")
		}

	}
}
