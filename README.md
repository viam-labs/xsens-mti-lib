# Xsens MTi Library
Library and utilities for connecting to Xsens MTi series IMUs.


Steps to compile in canon:
```sh
sudo apt install -y swig
make clean sdk swig build
```
(you can just run `make build` after the first time)

Steps to run:
```sh
./bin/xsens-mti-lib
```
