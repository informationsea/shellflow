package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestMemory(t *testing.T) {
	{
		m, e := NewMemory("1023")
		if e != nil {
			t.Fatalf("Bad parse result: %s", e.Error())
		}

		if x := m.memoryByte; x != 1023 {
			t.Fatalf("Bad memory byte: %d", x)
		}

		if x := m.Byte(); x != 1023 {
			t.Fatalf("Bad memory byte: %d", x)
		}

		if x := m.KiloByte(); x != 1023/1024. {
			t.Fatalf("Bad memory byte: %f", x)
		}

		if x := m.MegaByte(); x != 1023/1024./1024. {
			t.Fatalf("Bad memory byte: %f", x)
		}

		if x := m.GigaByte(); x != 1023/1024./1024./1024. {
			t.Fatalf("Bad memory byte: %f", x)
		}

		if s := m.String(); s != "1023" {
			t.Fatalf("Bad memory byte: %s", s)
		}
	}

	{
		m, e := NewMemory("1.1k")
		if e != nil {
			t.Fatalf("Bad parse result: %s", e.Error())
		}

		if x := m.memoryByte; x != 1126 {
			t.Fatalf("Bad memory byte: %d", x)
		}

		if x := m.Byte(); x != 1126 {
			t.Fatalf("Bad memory byte: %d", x)
		}

		if x := m.KiloByte(); x != 1126/1024. {
			t.Fatalf("Bad memory byte: %f", x)
		}

		if x := m.MegaByte(); x != 1126/1024./1024. {
			t.Fatalf("Bad memory byte: %f", x)
		}

		if x := m.GigaByte(); x != 1126/1024./1024./1024. {
			t.Fatalf("Bad memory byte: %f", x)
		}

		if s := m.String(); s != "1.10k" {
			t.Fatalf("Bad memory byte: %s", s)
		}
	}

	{
		m, e := NewMemory("1.21M")
		if e != nil {
			t.Fatalf("Bad parse result: %s", e.Error())
		}

		if x := m.memoryByte; x != 1268776 {
			t.Fatalf("Bad memory byte: %d", x)
		}

		if x := m.Byte(); x != 1268776 {
			t.Fatalf("Bad memory byte: %d", x)
		}

		if x := m.KiloByte(); x != 1268776/1024. {
			t.Fatalf("Bad memory byte: %f", x)
		}

		if x := m.MegaByte(); x != 1268776/1024./1024. {
			t.Fatalf("Bad memory byte: %f", x)
		}

		if x := m.GigaByte(); x != 1268776/1024./1024./1024. {
			t.Fatalf("Bad memory byte: %f", x)
		}

		if s := m.String(); s != "1.21M" {
			t.Fatalf("Bad memory byte: %s", s)
		}
	}

	{
		m, e := NewMemory("2.34G")
		if e != nil {
			t.Fatalf("Bad parse result: %s", e.Error())
		}

		if x := m.memoryByte; x != 2512555868 {
			t.Fatalf("Bad memory byte: %d", x)
		}

		if s := m.String(); s != "2.34G" {
			t.Fatalf("Bad memory byte: %s", s)
		}
	}
}

func TestTomlLoad(t *testing.T) {
	var d Configuration
	_, err := toml.DecodeFile("default_config.toml", &d)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	expected := Configuration{
		Environment: nil,
		Backend:     Backend{""},
		Command: []CommandConfiguration{
			CommandConfiguration{
				RegExp:       "mkdir +(-p +)?[^;&]+",
				RunImmediate: true,
			},
			CommandConfiguration{
				RegExp:          "example_dont_inheirt_path",
				DontInheirtPath: true,
			},
			CommandConfiguration{
				RegExp:    "gatk .*",
				SGEOption: []string{"-l", "s_vmem=20G,mem_req=20G", "-pe", "def_slot", "2"},
			},
			CommandConfiguration{
				RegExp:    "samtools .*",
				SGEOption: []string{"-l", "s_vmem=10G,mem_req=10G", "-pe", "def_slot", "2"},
			},
			CommandConfiguration{
				RegExp:    "java .*",
				SGEOption: []string{"-l", "s_vmem=40G,mem_req=40G"},
			},
		},
	}

	if !reflect.DeepEqual(expected, d) {
		j, err := json.MarshalIndent(d, "", "  ")
		if err != nil {
			t.Fatalf("%s", err.Error())
		}
		t.Fatalf("bad conf: %s", j)
	}
}

func TestTomlLoad2(t *testing.T) {
	_, err := LoadConfiguration()
	if err != nil {
		t.Fatalf("error: %s", err.Error())
	}
}
