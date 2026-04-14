## Run copy util manually with given arguments
```
go build -o gopy
mkdir dstdir
./gopy -from testdata/input.txt -to dstdir/out_offset0_limit0.txt -offset 0 -limit 0
./gopy -from testdata/input.txt -to dstdir/out_offset0_limit10.txt -offset 0 -limit 10
./gopy -from testdata/input.txt -to dstdir/out_offset0_limit1000.txt -offset 0 -limit 1000
./gopy -from testdata/input.txt -to dstdir/out_offset0_limit10000.txt -offset 0 -limit 10000
./gopy -from testdata/input.txt -to dstdir/out_offset100_limit1000.txt -offset 100 -limit 1000
./gopy -from testdata/input.txt -to dstdir/out_offset6000_limit1000.txt -offset 6000 -limit 1000
```