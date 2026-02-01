package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"vn-admin-api/internal/models"

	"github.com/redis/go-redis/v9"
)

// Cache is the interface for caching - implemented by Redis or in-memory
type Cache interface {
	GetProvinces(ctx context.Context) ([]models.Province, bool)
	SetProvinces(ctx context.Context, provinces []models.Province) error
	GetUnits(ctx context.Context, provinceID int) ([]models.AdminUnit, bool)
	SetUnits(ctx context.Context, provinceID int, units []models.AdminUnit) error
	Ping(ctx context.Context) error
}

// =============================================================================
// Redis Cache Implementation
// =============================================================================

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(url string, ttl time.Duration) (*RedisCache, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &RedisCache{client: client, ttl: ttl}, nil
}

func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *RedisCache) GetProvinces(ctx context.Context) ([]models.Province, bool) {
	data, err := c.client.Get(ctx, "provinces").Bytes()
	if err != nil {
		return nil, false
	}

	var provinces []models.Province
	if err := json.Unmarshal(data, &provinces); err != nil {
		return nil, false
	}
	return provinces, true
}

func (c *RedisCache) SetProvinces(ctx context.Context, provinces []models.Province) error {
	data, err := json.Marshal(provinces)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "provinces", data, c.ttl).Err()
}

func (c *RedisCache) GetUnits(ctx context.Context, provinceID int) ([]models.AdminUnit, bool) {
	key := fmt.Sprintf("units:%d", provinceID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	var units []models.AdminUnit
	if err := json.Unmarshal(data, &units); err != nil {
		return nil, false
	}
	return units, true
}

func (c *RedisCache) SetUnits(ctx context.Context, provinceID int, units []models.AdminUnit) error {
	key := fmt.Sprintf("units:%d", provinceID)
	data, err := json.Marshal(units)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

var _ Cache = (*RedisCache)(nil)

// =============================================================================
// Memory Cache Implementation (Fallback)
// =============================================================================

type MemoryCache struct {
	mu          sync.RWMutex
	provinces   []models.Province
	provinceExp time.Time

	unitsMu sync.RWMutex
	units   map[int]cachedUnits

	ttl time.Duration
}

type cachedUnits struct {
	data   []models.AdminUnit
	expiry time.Time
}

func NewMemoryCache(ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		ttl:   ttl,
		units: make(map[int]cachedUnits),
	}
}

func (c *MemoryCache) Ping(ctx context.Context) error {
	return nil // Memory is always available
}

func (c *MemoryCache) GetProvinces(ctx context.Context) ([]models.Province, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if time.Now().Before(c.provinceExp) && len(c.provinces) > 0 {
		return c.provinces, true
	}
	return nil, false
}

func (c *MemoryCache) SetProvinces(ctx context.Context, provinces []models.Province) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.provinces = provinces
	c.provinceExp = time.Now().Add(c.ttl)
	return nil
}

func (c *MemoryCache) GetUnits(ctx context.Context, provinceID int) ([]models.AdminUnit, bool) {
	c.unitsMu.RLock()
	defer c.unitsMu.RUnlock()

	if cached, ok := c.units[provinceID]; ok {
		if time.Now().Before(cached.expiry) {
			return cached.data, true
		}
	}
	return nil, false
}

func (c *MemoryCache) SetUnits(ctx context.Context, provinceID int, units []models.AdminUnit) error {
	c.unitsMu.Lock()
	defer c.unitsMu.Unlock()

	c.units[provinceID] = cachedUnits{
		data:   units,
		expiry: time.Now().Add(c.ttl),
	}
	return nil
}

var _ Cache = (*MemoryCache)(nil)
