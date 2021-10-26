1. [How to generate a random string of a fixed length in Go?](https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go)
2. [鸟窝 快速产生一个随机字符串](https://colobu.com/2018/09/02/generate-random-string-in-Go/)

```bash
$ go test -bench=. -benchmem
goos: darwin
goarch: amd64
pkg: github.com/bingoohuang/wkp/pkg/ss
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
BenchmarkRunes-12                                        2499286               480.8 ns/op            88 B/op          2 allocs/op
BenchmarkBytes-12                                        3592778               336.6 ns/op            32 B/op          2 allocs/op
BenchmarkBytesRmndr-12                                   3781815               292.9 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMask-12                                    3840357               336.0 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMaskImpr-12                               10026261               122.6 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMaskImprSrc-12                            11770927               111.3 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMaskImprXorshift1024Src-12                11951936               100.3 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMaskImprXorshift256Src-12                 12221281                96.66 ns/op           32 B/op          2 allocs/op
BenchmarkBytesMaskImprXorShift64StarSrc-12               8884897               134.6 ns/op            48 B/op          3 allocs/op
BenchmarkBytesMaskImprXorShift128PlusSrc-12              9110806               133.1 ns/op            48 B/op          3 allocs/op
BenchmarkBytesMaskImprXorShift1024StarSrc-12             8824548               136.1 ns/op            48 B/op          3 allocs/op
BenchmarkSecureRandomAlphaString-12                       900402              1354 ns/op              64 B/op          3 allocs/op
BenchmarkSecureRandomString-12                            930405              1348 ns/op              61 B/op          3 allocs/op
BenchmarkShortID-12                                      1407378               857.1 ns/op            32 B/op          2 allocs/op
BenchmarkGenerate-12                                    19989634                59.91 ns/op           16 B/op          1 allocs/op
BenchmarkRandStr-12                                     23874699                52.95 ns/op           16 B/op          1 allocs/op
BenchmarkUniuriNewLen-12                                 1354945               888.0 ns/op            56 B/op          3 allocs/op
PASS
ok      github.com/bingoohuang/wkp/pkg/ss       24.899s
```
