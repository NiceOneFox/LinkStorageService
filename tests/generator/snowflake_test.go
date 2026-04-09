package generator_test

import (
	"sync"
	"testing"

	"LinkStorageService/internal/generator"
)

func TestSnowflakeGenerator_Structure(t *testing.T) {
	gen, err := generator.NewSnowflakeGenerator(42)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	id := gen.Generate()

	nodeMask := int64(1023)
	seqMask := int64(4095)

	nodeID := (id >> 12) & nodeMask
	sequence := id & seqMask

	if nodeID != 42 {
		t.Errorf("Expected nodeID=42, got %d", nodeID)
	}

	if sequence < 0 || sequence > 4095 {
		t.Errorf("Sequence out of range [0,4095]: got %d", sequence)
	}
}

func TestSnowflakeGenerator_Uniqueness(t *testing.T) {
	gen, err := generator.NewSnowflakeGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	const iterations = 100000
	seen := make(map[int64]bool)

	for i := 0; i < iterations; i++ {
		id := gen.Generate()
		if seen[id] {
			t.Errorf("Duplicate ID generated: %d", id)
		}
		seen[id] = true
	}
}

func TestSnowflakeGenerator_ConcurrentUniqueness(t *testing.T) {
	gen, err := generator.NewSnowflakeGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	const goroutines = 10
	const idsPerGoroutine = 10000

	var wg sync.WaitGroup
	ids := make(chan int64, goroutines*idsPerGoroutine)

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < idsPerGoroutine; i++ {
				ids <- gen.Generate()
			}
		}()
	}

	wg.Wait()
	close(ids)

	seen := make(map[int64]bool)
	for id := range ids {
		if seen[id] {
			t.Errorf("Duplicate ID generated concurrently: %d", id)
		}
		seen[id] = true
	}
}

func TestSnowflakeGenerator_Monotonic(t *testing.T) {
	gen, err := generator.NewSnowflakeGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	var prev int64
	for i := 0; i < 1000; i++ {
		current := gen.Generate()
		if i > 0 && current <= prev {
			t.Errorf("IDs not monotonic: prev=%d, current=%d", prev, current)
		}
		prev = current
	}
}
