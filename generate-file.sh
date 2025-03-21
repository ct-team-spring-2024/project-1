mkdir files
# dd if=/dev/zero of=files/largefile.bin bs=1M count=200
dd if=/dev/zero of=files/largefile.bin bs=1M count=50
