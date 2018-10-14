Shellflow
=========

[![Build Status](https://travis-ci.org/informationsea/shellflow.svg?branch=master)](https://travis-ci.org/informationsea/shellflow)

Shell Script like scientific workflow management system

Flowscript
----------

Flowscript is simple, expression-oriented and dynamic typing script language. This script is embedded to brace in shellflow.

### Supported types

* string
* int
* file
* array
* map
* function

### Flowscript syntax

* Number
  * example: `1`, `100`
  * Currently, negative number is not supported in flowscript 
* Number Operations
  * example: `1 + 2`, `2 - 3`, ` + 2`, `1 + 2`

### Built-in functions

* `basename(path[, suffix])`
  * return base name of a path. if a suffix is provided and found in the path, this function removes the suffix.
  * example: `basename("hoge/foo") => "foo"`
  * example: `basename("hoge/foo.c", ".c") => "foo"`
* `dirname(path)`
  * return directory name of a path.
  * example: `dirname("hoge/foo") => "hoge"`
* `prefix(text, array)`
  * add prefix to arrayed string
  * example: `prefix("hoge", ["foo", "bar", "hoge"]) => ["hogefoo", "hogebar", "hogehoge"]`
* `zip(array1, array2)`
  * Zip two arrays and create an array of arrays.
  * example: `zip([1,2,3], [4,5,6,7]) => [[1,4], [2,5], [3,6]]`
* `file(path)`
  * Convert to file type from string

Shellflow Syntax
----------------

Almost all lines will be passed to shell directly. Before passing to shell, shellflow evaluates an embedded flowscript in a line and replace braces with the result of the flowscript.

### Input file annotation

Input file name should be surrounded with `((` and `))`. Shell variables should not be contained in the brackets, but flowscript can be contained.

### Output file annotation

Output file name should be surrounded with `[[` and `]]`. Shell variables should not be contained in the brackets, but flowscript can be contained.

### Embedded flowscript

Embedded flowscript should be surrounded with `{{` and `}}`. Braces cannot be omitted even if only one variable specified.

### `flowscript` statement

A line that starts with `#%` will be parsed as raw flowscript expression.

### `if` statement

This statement is not implemented yet.

### `for` statement

This statement is not implemented yet.

Configuration
-------------

This feature is not implemented yet.

Supported Backends
------------------

* Local Execution
* Grid Engine variants (coming soon)

Example
-------
