Flowscript
==========

Flowscript is embedded language in shellflow. It is used in enclosed
region with double curly brackets ``{{}}`` and lines starts with ``#%``.
You can try flowscript in REPL with running ``shellflow flowscript``.

Syntax
------

String
~~~~~~

Strings shoulde be enclosed with double quote ``"``

Example: ``"value"``

Built-in functions
------------------

basename
~~~~~~~~

Return base name of a path. if a suffix is provided and found in the
path, this function removes the suffix.

-  ``basename("hoge/foo.c") => "foo.c"``
-  ``basename("hoge/foo.c", ".c") => "foo"``

dirname
~~~~~~~

Return directory name of a path.

-  ``dirname("bar/hoge/foo.c") => "var/hoge"``

prefix
~~~~~~

add prefix to arrayed string

``prefix("hoge", ["foo", "bar", "hoge"]) => ["hogefoo", "hogebar", "hogehoge"]``

zip
~~~

Zip two arrays and create an array of arrays.

``zip([1,2,3], [4,5,6,7]) => [[1,4], [2,5], [3,6]]``
