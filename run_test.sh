#!/bin/bash

file="1GB.bin"

echo "### NLB Test"
./client -a 37.153.116.128:443 -t true -f ./$file  -b 1600

echo "### Target Test"
./client -a 37.153.116.251:9000 -t true -f ./$file  -b 1600 

echo "### CLB Test"
./client -a 37.153.116.43:443 -t true -f ./$file  -b 1600


