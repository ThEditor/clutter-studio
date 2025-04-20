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
