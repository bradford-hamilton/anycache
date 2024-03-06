package anycache

import (
	"reflect"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		capacity int
	}
	tests := []struct {
		name    string
		args    args
		want    *AnyCache
		wantErr bool
	}{
		{
			name:    "zero capacity",
			args:    args{capacity: 0},
			want:    nil,
			wantErr: true,
		},
		{
			name: "arbitrary valid capacity 1",
			args: args{capacity: 1},
			want: &AnyCache{
				items: map[any]any{},
				mu:    &sync.RWMutex{},
			},
			wantErr: false,
		},
		{
			name: "arbitrary valid capacity 100",
			args: args{capacity: 100},
			want: &AnyCache{
				items: map[any]any{},
				mu:    &sync.RWMutex{},
			},
			wantErr: false,
		},
		{
			name: "arbitrary valid capacity 10000",
			args: args{capacity: 10000},
			want: &AnyCache{
				items: map[any]any{},
				mu:    &sync.RWMutex{},
			},
			wantErr: false,
		},
		{
			name: "max capacity",
			args: args{capacity: maxCacheSize},
			want: &AnyCache{
				items: map[any]any{},
				mu:    &sync.RWMutex{},
			},
			wantErr: false,
		},
		{
			name:    "over capacity",
			args:    args{capacity: maxCacheSize + 1},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.capacity)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnyCache_Set(t *testing.T) {
	t.Run("add new item", func(t *testing.T) {
		cache, err := New(10)
		if err != nil {
			t.Fatal("Failed to create AnyCache:", err)
		}

		item, overwritten := cache.Set("key1", "value1")

		assertFalse(t, overwritten, "new item should not overwrite")
		assertEqual(t, "key1", item.Key, "incorrect key")
		assertEqual(t, "value1", item.Value, "incorrect value")
	})

	t.Run("overwrite item", func(t *testing.T) {
		cache, err := New(10)
		if err != nil {
			t.Fatal("Failed to create AnyCache:", err)
		}

		_, _ = cache.Set("key1", "value1")
		item, overwritten := cache.Set("key1", "value2")

		assertTrue(t, overwritten, "existing item should be overwritten")
		assertEqual(t, "key1", item.Key, "incorrect key after overwrite")
		assertEqual(t, "value2", item.Value, "incorrect value after overwrite")
	})

	t.Run("concurrent access", func(t *testing.T) {
		cache, err := New(10)
		if err != nil {
			t.Fatal("Failed to create AnyCache:", err)
		}

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				cache.Set(i, i)
			}(i)
		}
		wg.Wait()

		if len(cache.items) != 100 {
			t.Errorf("expected 100 items in cache, got %d", len(cache.items))
		}
	})
}

func TestAnyCache_Get(t *testing.T) {
	t.Run("retrieve existing item", func(t *testing.T) {
		cache, err := New(10)
		if err != nil {
			t.Fatal("Failed to create AnyCache:", err)
		}

		_, _ = cache.Set("key1", "value1")
		item, hit := cache.Get("key1")

		assertTrue(t, hit, "should be a cache hit")
		assertEqual(t, "key1", item.Key, "incorrect key retrieved")
		assertEqual(t, "value1", item.Value, "incorrect value retrieved")
	})

	t.Run("attempt to retrieve non-existing item", func(t *testing.T) {
		cache, err := New(10)
		if err != nil {
			t.Fatal("Failed to create AnyCache:", err)
		}

		item, hit := cache.Get("nonExistingKey")

		assertFalse(t, hit, "should be a cache miss")
		assertEqual(t, Item{}, item, "should return zero value for Item on miss")
	})

	t.Run("concurrent access with gets", func(t *testing.T) {
		cache, err := New(100)
		if err != nil {
			t.Fatal("Failed to create AnyCache:", err)
		}

		// Prepopulate cache
		for i := 0; i < 50; i++ {
			cache.Set(i, i)
		}

		var wg sync.WaitGroup
		hitCount := 0
		mu := sync.Mutex{}

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				_, hit := cache.Get(i)
				if hit {
					mu.Lock()
					hitCount++
					mu.Unlock()
				}
			}(i)
		}
		wg.Wait()

		if hitCount != 50 {
			t.Errorf("expected 50 hits, got %d", hitCount)
		}
	})
}

func TestAnyCache_Keys(t *testing.T) {
	t.Run("retrieve keys from populated cache", func(t *testing.T) {
		cache, err := New(10)
		if err != nil {
			t.Fatal("Failed to create AnyCache:", err)
		}

		keysToSet := []any{"key1", "key2", "key3"}
		for _, key := range keysToSet {
			cache.Set(key, "value")
		}

		returnedKeys := cache.Keys()

		// Verify all keys are present
		for _, key := range keysToSet {
			found := false
			for _, returnedKey := range returnedKeys {
				if returnedKey == key {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("key %v was not found in the returned keys", key)
			}
		}

		// Verify no extra keys are present
		if len(returnedKeys) != len(keysToSet) {
			t.Errorf("expected %d keys, got %d", len(keysToSet), len(returnedKeys))
		}
	})

	t.Run("concurrent access to keys", func(t *testing.T) {
		cache, err := New(100)
		if err != nil {
			t.Fatal("Failed to create AnyCache:", err)
		}

		// Prepopulate cache
		for i := 0; i < 50; i++ {
			cache.Set(i, i)
		}

		var wg sync.WaitGroup
		for i := 50; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				cache.Set(i, i)
			}(i)
		}

		// Attempt to read keys while the cache is being modified
		go func() {
			wg.Wait() // Wait for all sets to complete
		}()

		returnedKeys := cache.Keys()

		if len(returnedKeys) < 50 || len(returnedKeys) > 100 {
			t.Errorf("unexpected number of keys returned: got %d, expected between 50 and 100", len(returnedKeys))
		}
	})
}

func TestAnyCache_Flush(t *testing.T) {
	cache, err := New(10)
	if err != nil {
		t.Fatal("Failed to create AnyCache:", err)
	}

	// Prepopulate cache
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	if cache.Len() != 2 {
		t.Fatalf("expected 2 items in cache before flush, got %d", cache.Len())
	}

	// Test flushing the cache
	cache.Flush()
	if cache.Len() != 0 {
		t.Errorf("expected 0 items in cache after flush, got %d", cache.Len())
	}
}

func TestAnyCache_Len(t *testing.T) {
	cache, err := New(10)
	if err != nil {
		t.Fatal("Failed to create AnyCache:", err)
	}

	if cache.Len() != 0 {
		t.Fatalf("expected 0 items in cache initially, got %d", cache.Len())
	}

	// Add items to the cache
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	if cache.Len() != 2 {
		t.Errorf("expected 2 items in cache after adding items, got %d", cache.Len())
	}

	// Flush and check length
	cache.Flush()
	if cache.Len() != 0 {
		t.Errorf("expected 0 items in cache after flush, got %d", cache.Len())
	}
}

// assertEqual checks if two values are equal and reports a test error if not.
func assertEqual(t *testing.T, expected, actual any, msg string) {
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// assertTrue checks if a value is true and reports a test error if not.
func assertTrue(t *testing.T, actual bool, msg string) {
	if !actual {
		t.Errorf("%s: expected true, got %v", msg, actual)
	}
}

// assertFalse checks if a value is false and reports a test error if not.
func assertFalse(t *testing.T, actual bool, msg string) {
	if actual {
		t.Errorf("%s: expected false, got %v", msg, actual)
	}
}
