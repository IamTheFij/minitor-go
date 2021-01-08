#! /bin/bash

find ./dist -type f -executable -execdir tar -czvf {}.tar.gz {} \;
