package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type Memory struct {
	memoryByte int64
}

func (m Memory) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.memoryByte)
}

func (m *Memory) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &m.memoryByte)
}

func mustParseNewMemory(s string) Memory {
	m, e := NewMemory(s)
	if e != nil {
		panic(e)
	}
	return m
}

func NewMemory(s string) (Memory, error) {
	var m float64
	var multiply float64 = 1
	var e error
	if strings.HasSuffix(s, "G") {
		m, e = strconv.ParseFloat(s[:len(s)-1], 64)
		multiply = 1024 * 1024 * 1024
	} else if strings.HasSuffix(s, "M") {
		m, e = strconv.ParseFloat(s[:len(s)-1], 64)
		multiply = 1024 * 1024
	} else if strings.HasSuffix(s, "K") || strings.HasSuffix(s, "k") {
		m, e = strconv.ParseFloat(s[:len(s)-1], 64)
		multiply = 1024
	} else {
		m, e = strconv.ParseFloat(s[:len(s)], 64)
	}

	if e != nil {
		return Memory{}, e
	}
	return Memory{int64(m * multiply)}, nil
}

func (m Memory) Byte() int64 {
	return m.memoryByte
}

func (m Memory) KiloByte() float32 {
	return float32(m.memoryByte) / 1024
}

func (m Memory) MegaByte() float32 {
	return m.KiloByte() / 1024
}

func (m Memory) GigaByte() float32 {
	return m.MegaByte() / 1024
}

func (m Memory) String() string {
	if m.memoryByte > 1024*1024*1024 {
		return fmt.Sprintf("%.2fG", m.GigaByte())
	}
	if m.memoryByte > 1024*1024 {
		return fmt.Sprintf("%.2fM", m.MegaByte())
	}
	if m.memoryByte > 1024 {
		return fmt.Sprintf("%.2fk", m.KiloByte())
	}
	return fmt.Sprintf("%d", m.Byte())
}

type CommandConfiguration struct {
	RegExp          string
	SGEOption       []string
	DontInheirtPath bool
	RunImmediate    bool
}

func (v *CommandConfiguration) String() string {
	return fmt.Sprintf("SGEOption: %s / DontInheirtPath: %t / RunImmediate: %t", v.SGEOption, v.DontInheirtPath, v.RunImmediate)
}

type Backend struct {
	Type string
}

type Configuration struct {
	Environment map[string]string
	Backend     Backend
	Command     []CommandConfiguration
}

//go:generate go-assets-builder --package=main --output=assets.go default_config.toml

var ShellflowConfig = os.ExpandEnv("${HOME}/.shellflow.toml")

func LoadConfiguration() (*Configuration, error) {
	// prefer local conf file
	localConfFile, err := os.Open("shellflow.toml")
	if err == nil {
		var conf Configuration
		_, err = toml.DecodeReader(localConfFile, &conf)
		if err != nil {
			return nil, fmt.Errorf("cannot read local TOML. %s", err.Error())
		}
		return &conf, nil
	}

	confFile, err := os.Open(ShellflowConfig)
	if err == nil {
		defer confFile.Close()
	} else if os.IsNotExist(err) {
		// copy default config file from asset
		defaultConf, err := Assets.Open("/default_config.toml")
		if err != nil {
			return nil, fmt.Errorf("Cannot open default config.: %s", err.Error())
		}
		defer defaultConf.Close()
		confFileToWrite, err := os.OpenFile(ShellflowConfig, os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			return nil, err
		}
		defer confFileToWrite.Close()
		_, err = io.Copy(confFileToWrite, defaultConf)
		if err != nil {
			return nil, fmt.Errorf("Cannot write default config data. %s", err.Error())
		}
		err = confFileToWrite.Sync()
		if err != nil {
			return nil, fmt.Errorf("Cannot sync config data. %s", err.Error())
		}

		confFile, err = os.Open(ShellflowConfig)
		if err != nil {
			return nil, fmt.Errorf("Cannot open config file for read. %s", err.Error())
		}
		defer confFile.Close()
	} else if err != nil {
		return nil, err
	}

	var conf Configuration
	_, err = toml.DecodeReader(confFile, &conf)
	if err != nil {
		return nil, fmt.Errorf("cannot read TOML. %s", err.Error())
	}
	return &conf, nil
}
