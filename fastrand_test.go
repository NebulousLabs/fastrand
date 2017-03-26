package fastrand

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"io"
	"math"
	"math/big"
	mrand "math/rand"
	"sync"
	"testing"
	"time"
)

// panics returns true if the function fn panicked.
func panics(fn func()) (panicked bool) {
	defer func() {
		panicked = (recover() != nil)
	}()
	fn()
	return
}

// TestIntnPanics tests that Intn panics if n <= 0.
func TestIntnPanics(t *testing.T) {
	// Test n < 0.
	if !panics(func() { Intn(-1) }) {
		t.Error("expected panic for n < 0")
	}

	// Test n = 0.
	if !panics(func() { Intn(0) }) {
		t.Error("expected panic for n == 0")
	}

	// Test n > 0.
	if panics(func() { Intn(1) }) {
		t.Error("did not expect panic for n > 0")
	}
}

// TestBigIntnPanics tests that BigIntn panics if n <= 0.
func TestBigIntnPanics(t *testing.T) {
	// Test n < 0.
	if !panics(func() { BigIntn(big.NewInt(-1)) }) {
		t.Error("expected panic for n < 0")
	}

	// Test n = 0.
	if !panics(func() { BigIntn(big.NewInt(0)) }) {
		t.Error("expected panic for n == 0")
	}

	// Test n > 0.
	if panics(func() { BigIntn(big.NewInt(1)) }) {
		t.Error("did not expect panic for n > 0")
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

// TestRandConcurrent checks that there are no race conditions when using the
// rngs concurrently.
func TestRandConcurrent(t *testing.T) {
	// Spin up one goroutine for each exported function. Each goroutine calls
	// its function in a tight loop.

	funcs := []func(){
		// Read some random data into a large byte slice.
		func() { Read(make([]byte, 16e3)) },

		// Call io.Copy on the global reader.
		func() { io.CopyN(new(bytes.Buffer), Reader, 16e3) },

		// Call Intn
		func() { Intn(math.MaxUint64/4 + 1) },

		// Call BigIntn on a 256-bit int
		func() { BigIntn(new(big.Int).SetBytes(Bytes(32))) },

		// Call Perm
		func() { Perm(150) },
	}

	closeChan := make(chan struct{})
	var wg sync.WaitGroup
	for i := range funcs {
		wg.Add(1)
		go func(i int) {
			for {
				select {
				case <-closeChan:
					wg.Done()
					return
				default:
				}

				funcs[i]()
			}
		}(i)
	}

	// Allow goroutines to run for a moment.
	time.Sleep(100 * time.Millisecond)

	// Close the channel and wait for everything to clean up.
	close(closeChan)
	wg.Wait()
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

// BenchmarkIntn benchmarks the Intn function for small ints.
func BenchmarkIntn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Intn(4e3)
	}
}

// BenchmarkIntnLarge benchmarks the Intn function for large ints.
func BenchmarkIntnLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// constant chosen to trigger resampling (see Intn)
		_ = Intn(math.MaxUint64/4 + 1)
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

// BenchmarkRead512K benchmarks the speed of Read for larger slices.
func BenchmarkRead512K(b *testing.B) {
	b.SetBytes(512e3)
	buf := make([]byte, 512e3)
	for i := 0; i < b.N; i++ {
		Read(buf)
	}
}

// BenchmarkRead4Threads benchmarks the speed of Read when it's being using
// across four threads.
func BenchmarkRead4Threads(b *testing.B) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 32)
			<-start
			for i := 0; i < b.N; i++ {
				Read(buf)
			}
			wg.Done()
		}()
	}
	b.SetBytes(4 * 32)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkRead4Threads512k benchmarks the speed of Read when it's being using
// across four threads with 512kb read sizes.
func BenchmarkRead4Threads512k(b *testing.B) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 512e3)
			<-start
			for i := 0; i < b.N; i++ {
				Read(buf)
			}
			wg.Done()
		}()
	}
	b.SetBytes(4 * 512e3)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkRead64Threads benchmarks the speed of Read when it's being using
// across 64 threads.
func BenchmarkRead64Threads(b *testing.B) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 32)
			<-start
			for i := 0; i < b.N; i++ {
				Read(buf)
			}
			wg.Done()
		}()
	}
	b.SetBytes(64 * 32)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkRead64Threads512k benchmarks the speed of Read when it's being using
// across 64 threads with 512kb read sizes.
func BenchmarkRead64Threads512k(b *testing.B) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 512e3)
			<-start
			for i := 0; i < b.N; i++ {
				Read(buf)
			}
			wg.Done()
		}()
	}
	b.SetBytes(64 * 512e3)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
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

// BenchmarkReadCrypto4Threads32 benchmarks the speed of rand.Read when its
// being used across 4 threads with 32 byte read sizes.
func BenchmarkReadCrypto4Threads32(b *testing.B) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 32)
			<-start
			for i := 0; i < b.N; i++ {
				_, err := rand.Read(buf)
				if err != nil {
					b.Fatal(err)
				}
			}
			wg.Done()
		}()
	}
	b.SetBytes(4 * 32)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkReadCrypto4Threads512k benchmarks the speed of rand.Read when its
// being used across 4 threads with 512 kb read sizes.
func BenchmarkReadCrypto4Threads512kb(b *testing.B) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 512e3)
			<-start
			for i := 0; i < b.N; i++ {
				_, err := rand.Read(buf)
				if err != nil {
					b.Fatal(err)
				}
			}
			wg.Done()
		}()
	}
	b.SetBytes(4 * 512e3)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkReadCrypto64Threads32 benchmarks the speed of rand.Read when its
// being used across 4 threads with 32 byte read sizes.
func BenchmarkReadCrypto64Threads32(b *testing.B) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 32)
			<-start
			for i := 0; i < b.N; i++ {
				_, err := rand.Read(buf)
				if err != nil {
					b.Fatal(err)
				}
			}
			wg.Done()
		}()
	}
	b.SetBytes(64 * 32)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkReadCrypto64Threads512k benchmarks the speed of rand.Read when its
// being used across 4 threads with 512 kb read sizes.
func BenchmarkReadCrypto64Threads512kb(b *testing.B) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			buf := make([]byte, 512e3)
			<-start
			for i := 0; i < b.N; i++ {
				_, err := rand.Read(buf)
				if err != nil {
					b.Fatal(err)
				}
			}
			wg.Done()
		}()
	}
	b.SetBytes(64 * 512e3)

	// Signal all threads to begin
	b.ResetTimer()
	close(start)
	// Wait for all threads to exit
	wg.Wait()
}

// BenchmarkReadCrypto512K benchmarks the speed of (crypto/rand).Read for larger
// slices. This establishes a lower limit for BenchmarkRead512K.
func BenchmarkReadCrypto512K(b *testing.B) {
	b.SetBytes(512e3)
	buf := make([]byte, 512e3)
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

// BenchmarkReadMath512K benchmarks the speed of (math/rand).Read for larger
// slices. This establishes an upper limit for BenchmarkRead512K.
func BenchmarkReadMath512K(b *testing.B) {
	b.SetBytes(512e3)
	buf := make([]byte, 512e3)
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
