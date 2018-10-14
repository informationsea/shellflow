# Configuration

Shellflow can be configured GridEngine options or other options with TOML file.

## Command

### RegExp

If a command to run is matched to a regular expression, options listed in below will be applied.

### RunImmediate

If this option is setted to true, command will be run immediately without using job scheduler.

### DontInheirtPath

If this option is setted to true, Shellflow will not inheirt `PATH` and `LD_LIBRARY_PATH` when running with job scheduler.

### SGEOption

This options will be passed to Univa/Sun Grid Engine `qsub`.

## Configuration Example

```toml
[[Command]]
RegExp = "mkdir +(-p +)?[^;&]+"
RunImmediate = true

[[Command]]
RegExp = "example_dont_inheirt_path"
DontInheirtPath = true

[[Command]]
RegExp = "gatk .*"
SGEOption = ["-l", "s_vmem=20G,mem_req=20G", "-pe", "def_slot", "2"]

[[Command]]
RegExp = "samtools .*"
SGEOption = ["-l", "s_vmem=10G,mem_req=10G", "-pe", "def_slot", "2"]

[[Command]]
RegExp = "java .*"
SGEOption = ["-l", "s_vmem=40G,mem_req=40G"]
```