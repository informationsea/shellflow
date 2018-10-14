Getting Started
===============

Shellflow was designed for rapid developing of research workflow. If you
can write bash script, you don't have to learn a lot of new syntax. Only
You have to add brackets to annotate which files are input or output.

Before starting tutorial
------------------------

In this tutorial, softwares listed in below are required.

-  `bwa <http://bio-bwa.sourceforge.net/>`__
-  `gatk4 <https://software.broadinstitute.org/gatk/download/>`__
-  `shellflow <https://github.com/informationsea/shellflow>`__

Data listed in below are also required.

-  Reference genome (for example:
   `hs37d5 <ftp://ftp.1000genomes.ebi.ac.uk/vol1/ftp/technical/reference/phase2_reference_assembly_sequence>`__)
-  BWA index of the reference genome (``bwa index hs37d5.fa``)
-  Sequence Dictionary File of the reference genome
   (``gatk --java-options "-Xmx4G" CreateSequenceDictionary -R hs37d5.fa``)
-  Some sequnece data (for example:
   `DRR002191 <https://trace.ddbj.nig.ac.jp/DRASearch/run?acc=DRR002191>`__)

1st step: mapping
-----------------

A syntax of shellflow script is very similar to bash shell script. All
you have to do is enclose input files with double parenthesis (``((``
and ``))``) and output files with double brackets (``[[`` and ``]]``).
You can use pipe and redirect in workflow script like usual shell
script.

Content of ``gettingstarted.sf``

.. code:: bash

    bwa mem -t 6 hs37d5.fa <(bzip2 -dc ((DRR002191_1.fastq.bz2))) <(bzip2 -dc ((DRR002191_2.fastq.bz2))) > [[DRR002191.sam]]

.. code:: bash

    $ shellflow run gettingstarted.sf

2nd step: check status
----------------------

.. code:: bash

    $ shellflow viewlog
      #|  State|Success|Failed|Running|Pending|File Changed|Start Date         |Name
      1|   Done|      1|     0|      0|      0|         Yes|2018/10/14 15:00:48|step1.sf

.. code:: bash

    $ shellflow viewlog 1
    Workflow Script Path: /home/okamura/Documents/Programming/GO/workspace/src/github.com/informationsea/shellflow/examples/getting-started/step1/step1.sf
       Workflow Log Path: shellflow-wf/20181014-145901.507-step1.sf-1103fc92-e078-4e47-a316-62c4f16cb935
               Job Start: 2018/10/14 15:00:48
     Changed Input Files:
    ---- Job: 1 ------------
                 State: JobDone
             Exit code: 0
              Reusable: No
                Script: bwa mem -t 6 hs37d5.fa DRR002191_1.fastq.bz2 DRR002191_2.fastq.bz2 > DRR002191.sam
                 Input: DRR002191_1.fastq.bz2 DRR002191_2.fastq.bz2
                Output: DRR002191.sam
     Dependent Job IDs:
         Log directory: shellflow-wf/20181014-145901.507-step1.sf-1103fc92-e078-4e47-a316-62c4f16cb935/job001

.. code:: bash

    $ ls shellflow-wf/20181014-145901.507-step1.sf-1103fc92-e078-4e47-a316-62c4f16cb935/job001
    input.json  local-run-pid.txt  output.json  rc  run.sh  run.stderr  run.stdout  script.sh  script.stderr  script.stdout

3rd step: add more commands
---------------------------

When you want to add a new command depends on previous command, add new
line at last. Shellflow automatically judge which commands depend on
other commands. Unlike Makefile, shellflow assumes all dependent
commands can be found before a command line.

.. code:: bash

    bwa mem -R "@RG\tID:DRR002191\tSM:DRR002191\tPL:illumina\tLB:DRR002191" -t 6 hs37d5.fa <(bzip2 -dc ((DRR002191_1.fastq.bz2))) <(bzip2 -dc ((DRR002191_2.fastq.bz2))) > [[DRR002191.sam]]
    gatk SortSam -I ((DRR002191.sam)) -O [[DRR002191-sorted.bam]] --SORT_ORDER coordinate
    gatk MarkDuplicates -I ((DRR002191-sorted.bam)) -O [[DRR002191-markdup.bam]] -M [[DRR002191-markdup-metrics.txt]]
    gatk BaseRecalibrator --known-sites ((common_all_20180423.vcf.gz)) -I ((DRR002191-markdup.bam)) -O [[DRR002191-bqsr.txt]] -R hs37d5.fa 

Shellflow runs only added commands.

.. code:: bash

    $ shellflow run gettingstarted.sf

4th step: use variable
----------------------

If a line starts with ``#%``, the line is parsed as flowscript, which is
embedded language of shellflow.

.. code:: bash

    #% SAMPLE_ID = "DRR002191"
    bwa mem -R "@RG\tID:{{SAMPLE_ID}}\tSM:{{SAMPLE_ID}}\tPL:illumina\tLB:{{SAMPLE_ID}}" -t 6 hs37d5.fa <(bzip2 -dc (({{SAMPLE_ID}}_1.fastq.bz2))) <(bzip2 -dc (({{SAMPLE_ID}}_2.fastq.bz2))) > [[{{SAMPLE_ID}}.sam]]
    gatk SortSam -I (({{SAMPLE_ID}}.sam)) -O [[{{SAMPLE_ID}}-sorted.bam]] --SORT_ORDER coordinate
    gatk MarkDuplicates -I (({{SAMPLE_ID}}-sorted.bam)) -O [[{{SAMPLE_ID}}-markdup.bam]] -M [[{{SAMPLE_ID}}-markdup-metrics.txt]]
    gatk BaseRecalibrator --known-sites ((common_all_20180423.vcf.gz)) -I (({{SAMPLE_ID}}-markdup.bam)) -O [[{{SAMPLE_ID}}-bqsr.txt]] -R hs37d5.fa 

5th step: use loop
------------------

.. code:: bash

    for SAMPLE_ID in DRR002191 DRR002192; do
        bwa mem -R "@RG\tID:{{SAMPLE_ID}}\tSM:{{SAMPLE_ID}}\tPL:illumina\tLB:{{SAMPLE_ID}}" -t 6 hs37d5.fa <(bzip2 -dc (({{SAMPLE_ID}}_1.fastq.bz2))) <(bzip2 -dc (({{SAMPLE_ID}}_2.fastq.bz2))) > [[{{SAMPLE_ID}}.sam]]
        gatk SortSam -I (({{SAMPLE_ID}}.sam)) -O [[{{SAMPLE_ID}}-sorted.bam]] --SORT_ORDER coordinate
        gatk MarkDuplicates -I (({{SAMPLE_ID}}-sorted.bam)) -O [[{{SAMPLE_ID}}-markdup.bam]] -M [[{{SAMPLE_ID}}-markdup-metrics.txt]]
        gatk BaseRecalibrator --known-sites ((common_all_20180423.vcf.gz)) -I (({{SAMPLE_ID}}-markdup.bam)) -O [[{{SAMPLE_ID}}-bqsr.txt]] -R hs37d5.fa 
    done

.. code:: bash

    #% SAMPLES = ["DRR002191", "DRR002192"]
    for SAMPLE_ID in {{SAMPLES}}; do
        bwa mem -R "@RG\tID:{{SAMPLE_ID}}\tSM:{{SAMPLE_ID}}\tPL:illumina\tLB:{{SAMPLE_ID}}" -t 6 hs37d5.fa <(bzip2 -dc (({{SAMPLE_ID}}_1.fastq.bz2))) <(bzip2 -dc (({{SAMPLE_ID}}_2.fastq.bz2))) > [[{{SAMPLE_ID}}.sam]]
        gatk SortSam -I (({{SAMPLE_ID}}.sam)) -O [[{{SAMPLE_ID}}-sorted.bam]] --SORT_ORDER coordinate
        gatk MarkDuplicates -I (({{SAMPLE_ID}}-sorted.bam)) -O [[{{SAMPLE_ID}}-markdup.bam]] -M [[{{SAMPLE_ID}}-markdup-metrics.txt]]
        gatk BaseRecalibrator --known-sites ((common_all_20180423.vcf.gz)) -I (({{SAMPLE_ID}}-markdup.bam)) -O [[{{SAMPLE_ID}}-bqsr.txt]] -R hs37d5.fa 
    done

6th step: map all FASTQ in a directory
--------------------------------------

.. code:: bash

    for FILENAME in *_1.fastq.bz2; do
        #% SAMPLE_ID = basename(FILENAME, "_1.fastq.bz2")
        bwa mem -R "@RG\tID:{{SAMPLE_ID}}\tSM:{{SAMPLE_ID}}\tPL:illumina\tLB:{{SAMPLE_ID}}" -t 6 hs37d5.fa <(bzip2 -dc (({{SAMPLE_ID}}_1.fastq.bz2))) <(bzip2 -dc (({{SAMPLE_ID}}_2.fastq.bz2))) > [[{{SAMPLE_ID}}.sam]]
        gatk SortSam -I (({{SAMPLE_ID}}.sam)) -O [[{{SAMPLE_ID}}-sorted.bam]] --SORT_ORDER coordinate
        gatk MarkDuplicates -I (({{SAMPLE_ID}}-sorted.bam)) -O [[{{SAMPLE_ID}}-markdup.bam]] -M [[{{SAMPLE_ID}}-markdup-metrics.txt]]
        gatk BaseRecalibrator --known-sites ((common_all_20180423.vcf.gz)) -I (({{SAMPLE_ID}}-markdup.bam)) -O [[{{SAMPLE_ID}}-bqsr.txt]] -R hs37d5.fa 
    done
