package fastrand

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	mrand "math/rand"
	"testing"
	"time"
)

// TestRead tests that Read produces output with sufficiently high entropy.
func TestRead(t *testing.T) {
	const size = 10e3

	var b bytes.Buffer
	zip, _ := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if _, err := zip.Write(Bytes(size)); err != nil {
		t.Fatal(err)
	}
	if err := zip.Close(); err != nil {
		t.Fatal(err)
	}
	if b.Len() < size {
		t.Error("supposedly high entropy bytes have been compressed!")
	}
}

// TestIntn tests the Intn function.
func TestIntn(t *testing.T) {
	const iters = 10000
	var counts [10]int
	for i := 0; i < iters; i++ {
		counts[Intn(len(counts))]++
	}
	exp := iters / len(counts)
	lower, upper := exp-(exp/10), exp+(exp/10)
	for i, n := range counts {
		if !(lower < n && n < upper) {
			t.Errorf("Expected range of %v-%v for index %v, got %v", lower, upper, i, n)
		}
	}
}

// TestPerm tests the Perm function.
func TestPerm(t *testing.T) {
	chars := "abcde" // string to be permuted
	createPerm := func() string {
		s := make([]byte, len(chars))
		for i, j := range Perm(len(chars)) {
			s[i] = chars[j]
		}
		return string(s)
	}

	// create (factorial(len(chars)) * 100) permutations
	permCount := make(map[string]int)
	for i := 0; i < 12000; i++ {
		permCount[createPerm()]++
	}

	// we should have seen each permutation approx. 100 times
	for p, n := range permCount {
		if n < 50 || n > 150 {
			t.Errorf("saw permutation %v times: %v", n, p)
		}
	}
}

// BenchmarkIntn benchmarks the Intn function for large ints.
func BenchmarkIntn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Intn(4e9)
	}
}

// BenchmarkIntnSmall benchmarks the Intn function for small ints.
func BenchmarkIntnSmall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Intn(4e3)
	}
}

// BenchmarkRead benchmarks the speed of Read for small slices.
func BenchmarkRead32(b *testing.B) {
	b.SetBytes(32)
	buf := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		Read(buf)
	}
}

// BenchmarkRead64K benchmarks the speed of Read for larger slices.
func BenchmarkRead64K(b *testing.B) {
	b.SetBytes(64e3)
	buf := make([]byte, 64e3)
	for i := 0; i < b.N; i++ {
		Read(buf)
	}
}

// BenchmarkReadContention benchmarks the speed of Read when 4 other
// goroutines are calling Intn in a tight loop.
func BenchmarkReadContention(b *testing.B) {
	b.SetBytes(32)
	for j := 0; j < 4; j++ {
		go func() {
			for {
				Intn(1)
				time.Sleep(time.Microsecond)
			}
		}()
	}
	buf := make([]byte, 32)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Read(buf)
	}
}

// BenchmarkReadCrypto benchmarks the speed of (crypto/rand).Read for small
// slices. This establishes a lower limit for BenchmarkRead32.
func BenchmarkReadCrypto32(b *testing.B) {
	b.SetBytes(32)
	buf := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		rand.Read(buf)
	}
}

// BenchmarkReadCrypto64K benchmarks the speed of (crypto/rand).Read for larger
// slices. This establishes a lower limit for BenchmarkRead64K.
func BenchmarkReadCrypto64K(b *testing.B) {
	b.SetBytes(64e3)
	buf := make([]byte, 64e3)
	for i := 0; i < b.N; i++ {
		rand.Read(buf)
	}
}

// BenchmarkReadMath benchmarks the speed of (math/rand).Read for small
// slices. This establishes an upper limit for BenchmarkRead32.
func BenchmarkReadMath32(b *testing.B) {
	b.SetBytes(32)
	buf := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		mrand.Read(buf)
	}
}

// BenchmarkReadMath64K benchmarks the speed of (math/rand).Read for larger
// slices. This establishes an upper limit for BenchmarkRead64K.
func BenchmarkReadMath64K(b *testing.B) {
	b.SetBytes(64e3)
	buf := make([]byte, 64e3)
	for i := 0; i < b.N; i++ {
		mrand.Read(buf)
	}
}

// BenchmarkPerm benchmarks the speed of Perm for small slices.
func BenchmarkPerm32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Perm(32)
	}
}

// BenchmarkPermLarge benchmarks the speed of Perm for large slices.
func BenchmarkPermLarge4k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Perm(4e3)
	}
}
