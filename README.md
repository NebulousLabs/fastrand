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
BenchmarkRead32-4           	 5000000	       249 ns/op	 128.47 MB/s
BenchmarkRead64K-4          	    3000	    497977 ns/op	 128.52 MB/s
BenchmarkReadCrypto32-4     	  500000	      2737 ns/op	  11.69 MB/s
BenchmarkReadCrypto64K-4    	     300	   4059629 ns/op	  15.76 MB/s
```