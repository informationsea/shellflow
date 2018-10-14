#include <stdio.h>

#include "lib1.h"
#include "lib2.h"

int main(int argc, char **argv) {
    printf("lib1: %d\n", lib1());
    printf("lib2: %d\n", lib2());
    return 0;
}