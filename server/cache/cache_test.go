package cache

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	c := New[string, int](time.Second)
	c.Set("a", 1)

	val, ok := c.Get("a")
	if !ok {
		t.Fatal("expected value to be found")
	}
	if val != 1 {
		t.Fatalf("expected 1, got %d", val)
	}
}

func TestCache_NotFound(t *testing.T) {
	c := New[string, int](time.Second)

	_, ok := c.Get("missing")
	if ok {
		t.Fatal("expected missing key to not be found")
	}
}

func TestCache_Expired(t *testing.T) {
	c := New[string, int](time.Millisecond*50)
	c.Set("a", 1)

	time.Sleep(60 * time.Millisecond)

	_, ok := c.Get("a")
	if ok {
		t.Fatal("expected expired key to not be found")
	}
}

func TestCache_SetWithTTL(t *testing.T) {
	c := New[string, int](time.Hour)
	c.SetWithTTL("a", 1, time.Millisecond*50)

	// should be present immediately
	val, ok := c.Get("a")
	if !ok || val != 1 {
		t.Fatal("expected value immediately after SetWithTTL")
	}

	time.Sleep(60 * time.Millisecond)

	_, ok = c.Get("a")
	if ok {
		t.Fatal("expected key to expire after custom TTL")
	}
}

func TestCache_Delete(t *testing.T) {
	c := New[string, int](time.Hour)
	c.Set("a", 1)
	c.Delete("a")

	_, ok := c.Get("a")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestCache_Clear(t *testing.T) {
	c := New[string, int](time.Hour)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Clear()

	if c.Size() != 0 {
		t.Fatal("expected cache to be empty after Clear")
	}

	_, ok := c.Get("a")
	if ok {
		t.Fatal("expected a to be cleared")
	}
}

func TestCache_Size(t *testing.T) {
	c := New[string, int](time.Hour)
	if c.Size() != 0 {
		t.Fatal("expected size 0 initially")
	}

	c.Set("a", 1)
	if c.Size() != 1 {
		t.Fatalf("expected size 1, got %d", c.Size())
	}

	c.Set("b", 2)
	if c.Size() != 2 {
		t.Fatalf("expected size 2, got %d", c.Size())
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := New[int, string](time.Hour)

	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				c.Set(n, "value")
				c.Get(n)
			}
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestCache_Generics(t *testing.T) {
	type result struct {
		Name string
		Code int
	}

	c := New[string, result](time.Hour)
	c.Set("key1", result{Name: "test", Code: 200})

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected value")
	}
	if val.Name != "test" || val.Code != 200 {
		t.Fatalf("unexpected value: %+v", val)
	}
}
