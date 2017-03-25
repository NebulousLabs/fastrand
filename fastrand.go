// Package fastrand implements a cryptographically secure pseudorandom number
// generator. The generator is seeded using the system's default entropy
// source, and thereafter produces random values via repeated hashing. As a
// result, fastrand can generate randomness much faster than crypto/rand, and
// generation cannot fail.
package fastrand

import (
	"crypto/rand"
	"io"
	"math"
	"math/big"
	"sync/atomic"
	"unsafe"

	"golang.org/x/crypto/blake2b"
)

// A randReader produces random values via repeated hashing. The entropy field
// is the concatenation of an initial seed and a 128-bit counter. Each time
// the entropy is hashed, the counter is incremented.
type randReader struct {
	progress uint64
	entropy  [32]byte
}

// Reader is a global, shared instance of a cryptographically strong pseudo-
// random generator. It uses blake2b as its hashing function. Reader is safe
// for concurrent use by multiple goroutines.
var Reader io.Reader

// init provides the initial entropy for the reader that will seed all numbers
// coming out of fastrand.
func init() {
	r := &randReader{}
	n, err := rand.Read(r.entropy[:])
	if err != nil || n != len(r.entropy) {
		panic("not enough entropy to fill fastrand reader at startup")
	}
	Reader = r
}

// Read fills b with random data. It always returns len(b), nil.
func (r *randReader) Read(b []byte) (int, error) {
	// Update the progress and use the result + the seed entropy to form a
	// unique seed to hash.
	progress := atomic.AddUint64(&r.progress, 1)

	// If it seems like the progress is about to reset, increment the first
	// entropy bytes. Without this check, the cycle time is 2^64 iterations, but
	// with the check it is a secure 2^128 iterations.
	if progress > 1<<63 {
		atomic.AddUint64((*uint64)(unsafe.Pointer(&r.entropy[0])), 1)
		atomic.StoreUint64(&r.progress, 0)
	}

	// Copy the entropy into a separately allocated slice.
	seed := make([]byte, 40)
	*(*uint64)(unsafe.Pointer(&seed[0])) = progress
	copy(seed[8:], r.entropy[:])

	// We now have a unique seed that we can twiddle to generate entropy.
	n := 0
	for n < len(b) {
		// Hash the seed to get more entropy.
		result := blake2b.Sum512(seed)
		n += copy(b[n:], result[:])

		// Increment the seed so that the next attempt is succesful. The seed
		// most be incremented along the second 8 bytes because the first 8
		// bytes are part of what makes the seed unique to other threads. That
		// leaves still a full 16 bytes of entropy that kept from the original
		// read, more than enough to provide secure output.
		*(*uint64)(unsafe.Pointer(&seed[16]))++
	}
	return n, nil
}

// Read is a helper function that calls Reader.Read on b. It always fills b
// completely.
func Read(b []byte) { Reader.Read(b) }

// Bytes is a helper function that returns n bytes of random data.
func Bytes(n int) []byte {
	b := make([]byte, n)
	Read(b)
	return b
}

// Intn returns a uniform random value in [0,n). It panics if n <= 0.
func Intn(n int) int {
	if n <= 0 {
		panic("fastrand: argument to Intn is <= 0")
	}
	// To eliminate modulo bias, keep selecting at random until we fall within
	// a range that is evenly divisible by n.
	// NOTE: since n is at most math.MaxUint64/2, max is minimized when:
	//    n = math.MaxUint64/4 + 1 -> max = math.MaxUint64 - math.MaxUint64/4
	// This gives an expected 1.333 tries before choosing a value < max.
	max := math.MaxUint64 - math.MaxUint64%uint64(n)
	b := Bytes(8)
	r := *(*uint64)(unsafe.Pointer(&b[0]))
	for r >= max {
		Read(b)
		r = *(*uint64)(unsafe.Pointer(&b[0]))
	}
	return int(r % uint64(n))
}

// BigIntn returns a uniform random value in [0,n). It panics if n <= 0.
//
// TODO: This particular function has not been extended to be faster, as it is
// just a passthrough.
func BigIntn(n *big.Int) *big.Int {
	i, _ := rand.Int(Reader, n)
	return i
}

// Perm returns a random permutation of the integers [0,n).
func Perm(n int) []int {
	m := make([]int, n)
	for i := 1; i < n; i++ {
		j := Intn(i + 1)
		m[i] = m[j]
		m[j] = i
	}
	return m
}
