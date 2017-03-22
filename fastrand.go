// Package fastrand implements a cryptographically secure pseudorandom number
// generator. The generator is seeded using the system's default entropy
// source, and thereafter produces random values via repeated hashing. As a
// result, fastrand can generate randomness much faster than crypto/rand, and
// generation cannot fail.
package fastrand

import (
	"crypto/rand"
	"hash"
	"math"
	"sync"
	"unsafe"

	"github.com/minio/blake2b-simd"
)

// A randReader produces random values via repeated hashing. The entropy field
// is the concatenation of an initial seed and a 128-bit counter. Each time
// the entropy is hashed, the counter is incremented.
type randReader struct {
	entropy  []byte
	h        hash.Hash
	hashSize int
	buf      [32]byte
	mu       sync.Mutex
}

// Read fills b with random data. It always returns len(b), nil.
func (r *randReader) Read(b []byte) (int, error) {
	r.mu.Lock()
	for i := 0; i < len(b); i += r.hashSize {
		// Increment counter.
		*(*uint64)(unsafe.Pointer(&r.entropy[0]))++
		if *(*uint64)(unsafe.Pointer(&r.entropy[0])) == 0 {
			*(*uint64)(unsafe.Pointer(&r.entropy[8]))++
		}
		// Hash the counter + initial seed.
		r.h.Reset()
		r.h.Write(r.entropy)
		r.h.Sum(r.buf[:0])

		// Fill out 'b'.
		copy(b[i:], r.buf[:])
	}
	r.mu.Unlock()
	return len(b), nil
}

// Reader is a global, shared instance of a cryptographically strong pseudo-
// random generator. It uses blake2b as its hashing function. Reader is safe
// for concurrent use by multiple goroutines.
var Reader = func() *randReader {
	// Use 64 bytes in case the first 32 aren't completely random.
	base := make([]byte, 64)
	_, err := rand.Read(base)
	if err != nil {
		panic("fastrand: no entropy available")
	}
	e := blake2b.Sum256(base)
	return &randReader{
		entropy:  append(make([]byte, 16), e[:]...),
		h:        blake2b.New256(),
		hashSize: len(e),
	}
}()

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
