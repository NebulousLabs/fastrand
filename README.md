fastrand
--------

[![GoDoc](https://godoc.org/github.com/NebulousLabs/fastrand?status.svg)](https://godoc.org/github.com/NebulousLabs/fastrand)
[![Go Report Card](http://goreportcard.com/badge/github.com/NebulousLabs/fastrand)](https://goreportcard.com/report/github.com/NebulousLabs/fastrand)

```
go get github.com/NebulousLabs/fastrand
```

`fastrand` implements a cryptographically secure pseudorandom number generator.
The generator is seeded using the system's default entropy source, and
thereafter produces random values via repeated hashing. As a result, `fastrand`
can generate randomness much faster than `crypto/rand`, and generation cannot
fail beyond the initial call to seed the generator.

Also unlike crypto/rand, fastrand will provide significant speedups when called
in parallel.


## Benchmarks ##

```
// 32 byte reads
BenchmarkRead32-4                     	 5000000	       240 ns/op	 132.90 MB/s
BenchmarkReadCrypto32-4               	  300000	      5974 ns/op	   5.36 MB/s

// 512 kb reads
BenchmarkRead512K-4                   	    1000	   1403510 ns/op	 364.80 MB/s
BenchmarkReadCrypto512K-4             	      50	  39849581 ns/op	  12.85 MB/s

// 32 byte reads using 4 threads
BenchmarkRead4Threads-4               	 3000000	       372 ns/op	 343.35 MB/s
BenchmarkReadCrypto4Threads32-4       	  100000	     16060 ns/op	   7.97 MB/s

// 512 kb reads using 4 threads
BenchmarkRead4Threads512k-4           	     500	   2417701 ns/op	 847.09 MB/s
BenchmarkReadCrypto4Threads512kb-4    	      10	 224889487 ns/op	   9.11 MB/s
```

## Security ##

The fastrand packages uses something similar to the Fortuna algorithm, which is
used in FreeBSD as its /dev/random. The techniques used by fastrand are known to
be secure, however the specific implementation has not been reviewed
extensively. Use with caution.

The general strategy is to use crypto/rand at init to get 32 bytes of strong
entropy. From there, the entropy concatenated to a counter and hashed
repeatedly, providing a new 64 bytes of random output each time the counter is
incremented. The counter is 16 bytes, which provides strong guarantees that a
cycle will not be seen throughout the lifetime of the program.

The sync/atomic package is used to ensure that multiple threads calling fastrand
concurrently are always guaranteed to end up with unique counters, allowing
callers to see speedups by calling concurrently, without compromising security.
