Command line options
====================

run
---

``run`` command runs a workflow.

Options of ``run``
~~~~~~~~~~~~~~~~~~

-  ``-dry-run``

   -  Print jobs to run without execute. When this option selected, only
      updated commands are printed.

-  ``-param PARAM_FILE``

   -  a parameter file

-  ``-script-only``

   -  Only for debug

-  ``-sge``

   -  Run with Sun/Univa Grid Engine

-  ``-skip-sha``

   -  Skip calculate SHA256 (not recommended)

-  ``-rerun``

   -  Rerun all commands even if no input or commands are changed

dot
---

``dot`` command generates command dependecy graph to draw with ``dot``.

Options of ``dot``
~~~~~~~~~~~~~~~~~~

-  ``-param PARAM_FILE``

   -  a parameter file

viewlog
-------

View workflow log.

Options of ``viewlog``
~~~~~~~~~~~~~~~~~~~~~~

-  ``-all``

   -  Show all log

-  ``-failed``

   -  Show failed job only

flowscript
----------

REPL environment of flowscript.
