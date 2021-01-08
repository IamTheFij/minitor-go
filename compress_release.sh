#! /bin/sh

find ./dist -type f -perm +111 -execdir tar -czvf {}.tar.gz {} \;
