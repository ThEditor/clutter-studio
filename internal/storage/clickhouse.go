package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/google/uuid"
)

type ClickHouseStorage struct {
	db *sql.DB
}

func NewClickHouseStorage(dsn string) (*ClickHouseStorage, error) {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	storage := &ClickHouseStorage{db: db}

	return storage, nil
}

func (s *ClickHouseStorage) Close() error {
	return s.db.Close()
}

type EventData struct {
	VisitorIP        string
	VisitorUserAgent string
	SiteID           string
	Referrer         string
	Page             string
}

type DeviceStats struct {
	DeviceType string `json:"device_type"`
	Count      int    `json:"count"`
}

type PageStats struct {
	Page  string `json:"page"`
	Count int    `json:"count"`
}

type ReferrerStats struct {
	Referrer string `json:"referrer"`
	Count    int    `json:"count"`
}

type VisitorStats struct {
	Day            time.Time `json:"day"`
	UniqueVisitors int       `json:"unique_visitors"`
}

func (s *ClickHouseStorage) GetSiteEventData(siteID uuid.UUID) ([]EventData, error) {
	rows, err := s.db.Query(`
		SELECT 
			visitor_ip,
			visitor_user_agent,
			site_id,
			referrer,
			created_on,
			page
		FROM events
		WHERE site_id = ?
	`, siteID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []EventData
	for rows.Next() {
		var event EventData
		var createdOn time.Time
		if err := rows.Scan(
			&event.VisitorIP,
			&event.VisitorUserAgent,
			&event.SiteID,
			&event.Referrer,
			&createdOn,
			&event.Page,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return events, nil
}

func (s *ClickHouseStorage) GetUniqueVisitors(siteID uuid.UUID) (int, error) {
	var uniqueVisitors int
	err := s.db.QueryRow(`
		SELECT uniqExact(visitor_ip || visitor_user_agent) AS unique_visitors
		FROM events
		WHERE site_id = ?
	`, siteID.String()).Scan(&uniqueVisitors)
	if err != nil {
		return 0, fmt.Errorf("failed to get unique visitors: %w", err)
	}
	return uniqueVisitors, nil
}

func (s *ClickHouseStorage) GetPageViews(siteID uuid.UUID) (int, error) {
	var pageViews int
	err := s.db.QueryRow(`
		SELECT count(*) AS page_views
		FROM events
		WHERE site_id = ?
	`, siteID.String()).Scan(&pageViews)
	if err != nil {
		return 0, fmt.Errorf("failed to get page views: %w", err)
	}
	return pageViews, nil
}

func (s *ClickHouseStorage) GetTopReferrers(siteID uuid.UUID, limit int) ([]ReferrerStats, error) {
	rows, err := s.db.Query(`
		SELECT referrer, count(*) AS count
		FROM events
		WHERE site_id = ?
		GROUP BY referrer
		ORDER BY count DESC
		LIMIT ?
	`, siteID.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top referrers: %w", err)
	}
	defer rows.Close()

	var results []ReferrerStats
	for rows.Next() {
		var stats ReferrerStats
		if err := rows.Scan(&stats.Referrer, &stats.Count); err != nil {
			return nil, fmt.Errorf("failed to scan referrer stats: %w", err)
		}
		results = append(results, stats)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over referrer stats: %w", err)
	}
	return results, nil
}

func (s *ClickHouseStorage) GetTopPages(siteID uuid.UUID, limit int) ([]PageStats, error) {
	rows, err := s.db.Query(`
		SELECT page, count(*) AS count
		FROM events
		WHERE site_id = ?
		GROUP BY page
		ORDER BY count DESC
		LIMIT ?
	`, siteID.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top pages: %w", err)
	}
	defer rows.Close()

	var results []PageStats
	for rows.Next() {
		var stats PageStats
		if err := rows.Scan(&stats.Page, &stats.Count); err != nil {
			return nil, fmt.Errorf("failed to scan page stats: %w", err)
		}
		results = append(results, stats)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over page stats: %w", err)
	}
	return results, nil
}

func (s *ClickHouseStorage) GetDeviceStats(siteID uuid.UUID) ([]DeviceStats, error) {
	rows, err := s.db.Query(`
		SELECT
		  device_type,
		  count(*) AS total
		FROM (
		  SELECT
			CASE
			  WHEN visitor_user_agent ILIKE '%Mobile%' AND visitor_user_agent ILIKE '%Tablet%' THEN 'Tablet'
			  WHEN visitor_user_agent ILIKE '%Tablet%' THEN 'Tablet'
			  WHEN visitor_user_agent ILIKE '%Mobile%' THEN 'Mobile'
			  ELSE 'Desktop'
			END AS device_type
		  FROM events
		  WHERE site_id = ?
		)
		GROUP BY device_type
		ORDER BY total DESC
	`, siteID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get device stats: %w", err)
	}
	defer rows.Close()

	var results []DeviceStats
	for rows.Next() {
		var stats DeviceStats
		if err := rows.Scan(&stats.DeviceType, &stats.Count); err != nil {
			return nil, fmt.Errorf("failed to scan device stats: %w", err)
		}
		results = append(results, stats)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over device stats: %w", err)
	}
	return results, nil
}

func (s *ClickHouseStorage) GetVisitorGraph(siteID uuid.UUID, startDate, endDate string) ([]VisitorStats, error) {
	rows, err := s.db.Query(`
		SELECT
		  toDate(created_on) AS day,
		  uniqExact(visitor_ip || visitor_user_agent) AS unique_visitors
		FROM events
		WHERE site_id = ?
		  AND created_on >= ?
		  AND created_on < ?
		GROUP BY day
		ORDER BY day ASC
	`, siteID.String(), startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get visitor graph data: %w", err)
	}
	defer rows.Close()

	var results []VisitorStats
	for rows.Next() {
		var stats VisitorStats
		if err := rows.Scan(&stats.Day, &stats.UniqueVisitors); err != nil {
			return nil, fmt.Errorf("failed to scan visitor stats: %w", err)
		}
		results = append(results, stats)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over visitor stats: %w", err)
	}
	return results, nil
}
