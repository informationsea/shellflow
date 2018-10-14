#!/usr/bin/python3

import argparse
import random


def _main():
    parser = argparse.ArgumentParser("Create mutated fasta")
    parser.add_argument("input", type=argparse.FileType("r"))
    parser.add_argument(
        "--output", "-o", type=argparse.FileType("w"), required=True)
    parser.add_argument("--mutation-rate", type=float, default=0.00001)
    parser.add_argument("--seed", type=int)
    options = parser.parse_args()

    if options.seed:
        random.seed(options.seed)

    for line in options.input:
        if line.startswith(">"):
            options.output.write(line)
            continue
        for one in line:
            if one in "ATCG" and random.random() < options.mutation_rate:
                options.output.write("ATCG"[random.randint(0, 3)])
            else:
                options.output.write(one)


if __name__ == "__main__":
    _main()
