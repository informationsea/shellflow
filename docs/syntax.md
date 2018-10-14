# Syntax

A syntax of shellflow is very similar to bash script.
Most important difference is annotating input files and output files with parentheses and brackets.

## Input files

All input files should be enclosed with parentheses.

```bash
cat ((INPUT_FILE))
```

If some input files are not appeared in command line, it should be written in comment.

```bash
gcc ((hello.c)) # ((hello.h))   this command depends on hello.c and hello.h
```

When glob is used, all input files should be exists in advanced of shellflow runing.

## Output files

All output files should be enclosed with brackets.

```bash
date >> [[OUTPUT]]
```

## Varaible

A varaible should be enclosed with curly brackets, when the variable is used in command line.

```bash
echo {{variable}}
```

To set value to variable, flowscript, embedded script language in shellflow, or parameter file should be used. When a line starts with `#%`, the line will be parsed as flowscript.

```bash
#% varaible = "foo"
```

## Loop

`for` loop is supported in shellflow. `do` and `done` are required in shellflow.

```bash
for x in a b c; do
    echo {{x}} > {{x}}
done
```

Variable names should be enclosed with curly brackets.

When glob is used, all input files should be exists in advanced of shellflow runing.

## if

`if` statement is not supported currently.
