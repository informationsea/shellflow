package main

import (
	"time"

	"github.com/jessevdk/go-assets"
)

var _Assets45c329d41a5eab17178890acca5fb6324cd36ece = "[[Command]]\nRegExp = \"mkdir +(-p +)?[^;&]+\"\nRunImmediate = true\n\n[[Command]]\nRegExp = \"example_dont_inheirt_path\"\nDontInheirtPath = true\n\n[[Command]]\nRegExp = \"gatk .*\"\nSGEOption = [\"-l\", \"s_vmem=20G,mem_req=20G\", \"-pe\", \"def_slot\", \"2\"]\n\n[[Command]]\nRegExp = \"samtools .*\"\nSGEOption = [\"-l\", \"s_vmem=10G,mem_req=10G\", \"-pe\", \"def_slot\", \"2\"]\n\n[[Command]]\nRegExp = \"java .*\"\nSGEOption = [\"-l\", \"s_vmem=40G,mem_req=40G\"]\n\n"

// Assets returns go-assets FileSystem
var Assets = assets.NewFileSystem(map[string][]string{"/": []string{"default_config.toml"}}, map[string]*assets.File{
	"/": &assets.File{
		Path:     "/",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1539159433, 1539159433000000000),
		Data:     nil,
	}, "/default_config.toml": &assets.File{
		Path:     "/default_config.toml",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1539160895, 1539160895000000000),
		Data:     []byte(_Assets45c329d41a5eab17178890acca5fb6324cd36ece),
	}}, "")
