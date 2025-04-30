all:
	cd cmd/client/;go build;
	cd cmd/server/;go build;

1GB:
	dd if=/dev/urandom of=1GB.bin bs=64M count=16 iflag=fullblock

1MB:
	dd if=/dev/urandom of=1MB.bin bs=1M count=1 iflag=fullblock

64MB:
	dd if=/dev/urandom of=64MB.bin bs=1M count=64 iflag=fullblock

100MB:
	dd if=/dev/urandom of=100MB.bin bs=1M count=100 iflag=fullblock
