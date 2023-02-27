package cache

import (
	"testing"
	"time"

	"github.com/shoenig/test/must"
)

func TestTTLCache_Close(t *testing.T) {
	t.Parallel()

	c := NewTTLCache[string, string](10 * time.Minute)
	c.Set("key1", "v", 1*time.Hour)
	c.Close()

	must.MapEmpty(t, c.m)
}

// nolint:goconst
func TestTTLCache_Get(t *testing.T) {
	t.Parallel()

	c := NewTTLCache[string, string](10 * time.Minute)
	defer c.Close()

	key := "key1"
	c.Set(key, "v", 1*time.Hour)

	must.Eq(t, "v", c.Get(key))
	must.Eq(t, "", c.Get("empty-key"))
}

func TestTTLCache_Remove(t *testing.T) {
	t.Parallel()

	c := NewTTLCache[string, string](10 * time.Minute)
	defer c.Close()

	key := "key1"
	c.Set(key, "v", 1*time.Hour)
	c.Remove(key)

	must.Eq(t, "", c.Get(key))
}

func TestTTLCache_Set(t *testing.T) {
	t.Parallel()

	c := NewTTLCache[string, string](10 * time.Minute)
	defer c.Close()

	key := "key1"
	c.Set(key, "v1", 1*time.Hour)
	must.Eq(t, "v1", c.Get(key))

	c.Set(key, "v2", 1*time.Hour)
	must.Eq(t, "v2", c.Get(key))
}

func TestTTLCache_cleanup(t *testing.T) {
	t.Parallel()

	c := NewTTLCache[string, string](1 * time.Second)
	defer c.Close()

	expiredKey, existingKey := "expired", "existing"
	c.Set(expiredKey, "v1", 1*time.Millisecond)
	c.Set(existingKey, "v2", 1*time.Hour)

	time.Sleep(3 * time.Second)

	must.Eq(t, "", c.Get(expiredKey))
	must.Eq(t, "v2", c.Get(existingKey))
}
