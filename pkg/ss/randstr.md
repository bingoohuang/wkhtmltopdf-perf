1. [How to generate a random string of a fixed length in Go?](https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go)
2. [鸟窝 快速产生一个随机字符串](https://colobu.com/2018/09/02/generate-random-string-in-Go/)

```bash
$ go test -bench=. -benchmem                      
YNYTEhJlglgoos: darwin
goarch: amd64
pkg: github.com/bingoohuang/wkp/pkg/ss
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
BenchmarkRunes-12                                        2742140               431.8 ns/op            88 B/op          2 allocs/op
BenchmarkBytes-12                                        3776026               320.6 ns/op            32 B/op          2 allocs/op
BenchmarkBytesRmndr-12                                   4696774               256.6 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMask-12                                    3989793               301.0 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMaskImpr-12                               10719890               115.9 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMaskImprSrc-12                             9758218               108.7 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMaskImprXorshift1024Src-12                10152680               106.9 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMaskImprXorshift256Src-12                 11865195               100.3 ns/op            32 B/op          2 allocs/op
BenchmarkBytesMaskImprXorShift64StarSrc-12               9157275               128.2 ns/op            48 B/op          3 allocs/op
BenchmarkBytesMaskImprXorShift128PlusSrc-12              9337490               129.0 ns/op            48 B/op          3 allocs/op
BenchmarkBytesMaskImprXorShift1024StarSrc-12             9027806               132.2 ns/op            48 B/op          3 allocs/op
BenchmarkSecureRandomAlphaString-12                       893791              1331 ns/op              64 B/op          3 allocs/op
BenchmarkSecureRandomString-12                            935944              1285 ns/op              61 B/op          3 allocs/op
BenchmarkShortID-12                                      1423569               855.9 ns/op            32 B/op          2 allocs/op
BenchmarkGenerate-12                                    20195468                59.50 ns/op           16 B/op          1 allocs/op
BenchmarkRandStr-12                                     23981430                50.56 ns/op           16 B/op          1 allocs/op
PASS
ok      github.com/bingoohuang/wkp/pkg/ss       22.178s
```
