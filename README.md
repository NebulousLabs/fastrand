fastrand
--------

[![GoDoc](https://godoc.org/github.com/NebulousLabs/fastrand?status.svg)](https://godoc.org/github.com/NebulousLabs/fastrand)
[![Go Report Card](http://goreportcard.com/badge/github.com/NebulousLabs/fastrand)](https://goreportcard.com/report/github.com/NebulousLabs/fastrand)

```
go get github.com/NebulousLabs/fastrand
```

`fastrand` implements a cryptographically secure pseudorandom number
generator. The generator is seeded using the system's default entropy source,
and thereafter produces random values via repeated hashing. As a result,
`fastrand` can generate randomness much faster than `crypto/rand`, and
generation cannot fail.


## Benchmarks ##

```
BenchmarkRead32-4             	10000000	       207 ns/op	 153.94 MB/s
BenchmarkRead512K-4           	     500	   3230813 ns/op	 158.47 MB/s
BenchmarkReadCrypto32-4       	  500000	      2697 ns/op	  11.86 MB/s
BenchmarkReadCrypto512K-4     	      50	  32161930 ns/op	  15.92 MB/s
```
