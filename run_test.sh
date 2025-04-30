#!/bin/bash

file="1GB.bin"
MAX=5

for (( i=0; i<MAX; i++ )) ; {
	echo "#### Round $i ###"

	echo "### Round $i: NLB Test"
	./client -a 37.153.116.128:443 -t true -f ./$file  -b 1600 -n NLB

	echo "### Round $i: Target Test"
	./client -a 37.153.116.251:9000 -t true -f ./$file  -b 1600 -n Target

	echo "### Round $i: CLB Test"
	./client -a 37.153.116.43:443 -t true -f ./$file  -b 1600 -n CLB
}

