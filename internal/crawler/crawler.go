package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"vn-admin-api/internal/config"
	"vn-admin-api/internal/database"
	"vn-admin-api/internal/logger"
	"vn-admin-api/internal/models"
)

const (
	URLProvinces = "https://sapnhap.bando.com.vn/pcotinh"
	URLUnits     = "https://sapnhap.bando.com.vn/ptracuu"
	MaxRetries   = 3
	RetryDelay   = 2 * time.Second
)

type Crawler struct {
	repo   *database.Repository
	log    *logger.Logger
	cookie string
	client *http.Client
}

func New(repo *database.Repository, log *logger.Logger, cfg *config.Config) *Crawler {
	return &Crawler{
		repo:   repo,
		log:    log,
		cookie: cfg.APICookie,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Crawler) Run(ctx context.Context) error {
	c.log.Info("Starting crawler process")

	// 1. Fetch Provinces
	c.log.Info("Fetching provinces...")
	provinces, err := c.fetchProvincesWithRetry(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch provinces: %w", err)
	}
	c.log.Info("Found provinces", "count", len(provinces))

	for _, p := range provinces {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		c.log.Info("Processing Province", "name", p.Name, "id", p.ID)

		// Insert Province
		if err := c.repo.UpsertProvince(ctx, p); err != nil {
			c.log.Error("Failed to upsert province", "id", p.ID, "error", err)
			continue
		}

		// 2. Fetch Units for Province
		units, err := c.fetchUnitsWithRetry(ctx, p.ID)
		if err != nil {
			c.log.Error("Failed to fetch units", "province_id", p.ID, "error", err)
			continue
		}
		c.log.Info("Found units", "province_id", p.ID, "count", len(units))

		for _, u := range units {
			if err := c.repo.UpsertAdminUnit(ctx, u); err != nil {
				c.log.Error("Failed to upsert unit", "id", u.ID, "error", err)
			}
		}

		time.Sleep(500 * time.Millisecond) // Polite delay
	}

	c.log.Info("Crawler finished successfully")
	return nil
}

func (c *Crawler) fetchProvincesWithRetry(ctx context.Context) ([]models.Province, error) {
	var provinces []models.Province
	err := c.retry(ctx, func() error {
		p, err := c.fetchProvinces()
		if err != nil {
			return err
		}
		provinces = p
		return nil
	})
	return provinces, err
}

func (c *Crawler) fetchUnitsWithRetry(ctx context.Context, provinceID int) ([]models.AdminUnit, error) {
	var units []models.AdminUnit
	err := c.retry(ctx, func() error {
		u, err := c.fetchUnits(provinceID)
		if err != nil {
			return err
		}
		units = u
		return nil
	})
	return units, err
}

func (c *Crawler) retry(ctx context.Context, op func() error) error {
	var lastErr error
	for i := 0; i < MaxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := op(); err != nil {
			lastErr = err
			c.log.Warn("Operation failed, retrying...", "attempt", i+1, "error", err)
			time.Sleep(RetryDelay * time.Duration(1<<i)) // Exponential backoff
			continue
		}
		return nil
	}
	return fmt.Errorf("max retries reached: %w", lastErr)
}

// Low-level fetch functions

func (c *Crawler) fetchProvinces() ([]models.Province, error) {
	data := url.Values{"id": {"0"}}
	req, _ := http.NewRequest("POST", URLProvinces, strings.NewReader(data.Encode()))
	c.setHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var provinces []models.Province
	if err := json.NewDecoder(resp.Body).Decode(&provinces); err != nil {
		return nil, err
	}
	return provinces, nil
}

func (c *Crawler) fetchUnits(provinceID int) ([]models.AdminUnit, error) {
	data := url.Values{"id": {fmt.Sprintf("%d", provinceID)}}
	req, _ := http.NewRequest("POST", URLUnits, strings.NewReader(data.Encode()))
	c.setHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var units []models.AdminUnit
	if err := json.Unmarshal(body, &units); err != nil {
		return nil, err
	}
	return units, nil
}

func (c *Crawler) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Cookie", c.cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")
	req.Header.Set("Origin", "https://sapnhap.bando.com.vn")
	req.Header.Set("Referer", "https://sapnhap.bando.com.vn/")
}
