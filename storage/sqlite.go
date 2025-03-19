package storage

import (
	"database/sql"
	"fmt"
	"sync"

	"llm-fw/types"

	_ "github.com/mattn/go-sqlite3"
)

// Request represents a generation request
type Request = types.Request

// ModelStats represents statistics for a model
type ModelStats = types.ModelStats

// SQLiteStorage implements the types.Storage interface using SQLite
type SQLiteStorage struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	s := &SQLiteStorage{db: db}
	if err := s.initDB(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	return s, nil
}

// SaveRequest saves a request to the database
func (s *SQLiteStorage) SaveRequest(req *types.Request) error {
	_, err := s.db.Exec(`
		INSERT INTO requests (
			id, user_id, model, prompt, response, tokens_in, tokens_out, server, latency_ms, status, error, timestamp, source
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		req.ID,
		req.UserID,
		req.Model,
		req.Prompt,
		req.Response,
		req.TokensIn,
		req.TokensOut,
		req.Server,
		req.LatencyMs,
		req.Status,
		req.Error,
		req.Timestamp,
		req.Source,
	)
	return err
}

// GetRequest retrieves a request by ID
func (s *SQLiteStorage) GetRequest(id string) (*types.Request, error) {
	var req types.Request
	err := s.db.QueryRow(`
		SELECT id, user_id, model, prompt, response, tokens_in, tokens_out, server, latency_ms, status, error, timestamp, source
		FROM requests
		WHERE id = ?
	`, id).Scan(
		&req.ID,
		&req.UserID,
		&req.Model,
		&req.Prompt,
		&req.Response,
		&req.TokensIn,
		&req.TokensOut,
		&req.Server,
		&req.LatencyMs,
		&req.Status,
		&req.Error,
		&req.Timestamp,
		&req.Source,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// ListRequests retrieves requests with limit
func (s *SQLiteStorage) ListRequests(limit int) ([]*types.Request, error) {
	rows, err := s.db.Query(`
		SELECT id, model, prompt, response, tokens_in, tokens_out, latency_ms, timestamp
		FROM requests
		ORDER BY timestamp DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*types.Request
	for rows.Next() {
		var req types.Request
		err := rows.Scan(
			&req.ID,
			&req.Model,
			&req.Prompt,
			&req.Response,
			&req.TokensIn,
			&req.TokensOut,
			&req.LatencyMs,
			&req.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}
	return requests, nil
}

// DeleteRequest deletes a request by ID
func (s *SQLiteStorage) DeleteRequest(id string) error {
	_, err := s.db.Exec("DELETE FROM requests WHERE id = ?", id)
	return err
}

// SaveModelStats saves model statistics to the database
func (s *SQLiteStorage) SaveModelStats(model string, stats *types.ModelStats) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO model_stats (
			model, total_requests, failed_requests, total_tokens_in, total_tokens_out, average_latency, last_used
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		model,
		stats.TotalRequests,
		stats.FailedRequests,
		stats.TotalTokensIn,
		stats.TotalTokensOut,
		stats.AverageLatency,
		stats.LastUsed,
	)
	return err
}

// GetModelStats retrieves model statistics from the database
func (s *SQLiteStorage) GetModelStats(model string) (*types.ModelStats, error) {
	var stats types.ModelStats
	err := s.db.QueryRow(`
		SELECT total_requests, failed_requests, total_tokens_in, total_tokens_out, average_latency, last_used 
		FROM model_stats 
		WHERE model = ?
	`, model).Scan(
		&stats.TotalRequests,
		&stats.FailedRequests,
		&stats.TotalTokensIn,
		&stats.TotalTokensOut,
		&stats.AverageLatency,
		&stats.LastUsed,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetAllModelStats retrieves statistics for all models
func (s *SQLiteStorage) GetAllModelStats() (map[string]*types.ModelStats, error) {
	rows, err := s.db.Query("SELECT model, total_requests, failed_requests, total_tokens_in, total_tokens_out, average_latency, last_used FROM model_stats")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]*types.ModelStats)
	for rows.Next() {
		var stat types.ModelStats
		var model string
		err := rows.Scan(&model, &stat.TotalRequests, &stat.FailedRequests, &stat.TotalTokensIn, &stat.TotalTokensOut, &stat.AverageLatency, &stat.LastUsed)
		if err != nil {
			return nil, err
		}
		stats[model] = &stat
	}
	return stats, nil
}

// ListModelStats retrieves all model statistics
func (s *SQLiteStorage) ListModelStats() (map[string]*types.ModelStats, error) {
	rows, err := s.db.Query(`
		SELECT model, total_requests, failed_requests, total_tokens_in, total_tokens_out, average_latency, last_used
		FROM model_stats
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]*types.ModelStats)
	for rows.Next() {
		var stat types.ModelStats
		var model string
		err := rows.Scan(&model, &stat.TotalRequests, &stat.FailedRequests, &stat.TotalTokensIn, &stat.TotalTokensOut, &stat.AverageLatency, &stat.LastUsed)
		if err != nil {
			return nil, err
		}
		stats[model] = &stat
	}
	return stats, nil
}

// DeleteModelStats deletes model statistics
func (s *SQLiteStorage) DeleteModelStats(model string) error {
	_, err := s.db.Exec("DELETE FROM model_stats WHERE model = ?", model)
	return err
}

// SaveModelStatsHistory saves model statistics history
func (s *SQLiteStorage) SaveModelStatsHistory(history *types.ModelStatsHistory) error {
	_, err := s.db.Exec(`
		INSERT INTO model_stats_history (
			id, model, total_requests, failed_requests, total_tokens_in, total_tokens_out, average_latency, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		history.ID,
		history.Model,
		history.TotalRequests,
		history.FailedRequests,
		history.TotalTokensIn,
		history.TotalTokensOut,
		history.AverageLatency,
		history.Timestamp,
	)
	return err
}

// GetModelStatsHistory retrieves model statistics history
func (s *SQLiteStorage) GetModelStatsHistory(model string, limit int) ([]*types.ModelStatsHistory, error) {
	rows, err := s.db.Query(`
		SELECT id, model, total_requests, failed_requests, total_tokens_in, total_tokens_out, average_latency, timestamp
		FROM model_stats_history
		WHERE model = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`, model, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*types.ModelStatsHistory
	for rows.Next() {
		var h types.ModelStatsHistory
		err := rows.Scan(
			&h.ID,
			&h.Model,
			&h.TotalRequests,
			&h.FailedRequests,
			&h.TotalTokensIn,
			&h.TotalTokensOut,
			&h.AverageLatency,
			&h.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, &h)
	}
	return history, nil
}

// ListModelStatsHistory retrieves all model statistics history
func (s *SQLiteStorage) ListModelStatsHistory(limit int) ([]*types.ModelStatsHistory, error) {
	rows, err := s.db.Query(`
		SELECT id, model, total_requests, failed_requests, total_tokens_in, total_tokens_out, average_latency, timestamp
		FROM model_stats_history
		ORDER BY timestamp DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*types.ModelStatsHistory
	for rows.Next() {
		var h types.ModelStatsHistory
		err := rows.Scan(
			&h.ID,
			&h.Model,
			&h.TotalRequests,
			&h.FailedRequests,
			&h.TotalTokensIn,
			&h.TotalTokensOut,
			&h.AverageLatency,
			&h.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, &h)
	}
	return history, nil
}

// DeleteModelStatsHistory deletes model statistics history
func (s *SQLiteStorage) DeleteModelStatsHistory(model string) error {
	_, err := s.db.Exec("DELETE FROM model_stats_history WHERE model = ?", model)
	return err
}

// Cleanup removes old data
func (s *SQLiteStorage) Cleanup() error {
	_, err := s.db.Exec(`
		DELETE FROM model_stats_history
		WHERE timestamp < datetime('now', '-30 days')
	`)
	return err
}

// NewHistoryManager creates a new history manager
func (s *SQLiteStorage) NewHistoryManager(size int) types.HistoryManager {
	return NewHistoryManager(s, size)
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// initDB initializes the database
func (s *SQLiteStorage) initDB() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS requests (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			model TEXT NOT NULL,
			prompt TEXT NOT NULL,
			response TEXT NOT NULL,
			tokens_in INTEGER NOT NULL,
			tokens_out INTEGER NOT NULL,
			server TEXT NOT NULL,
			latency_ms REAL NOT NULL,
			status INTEGER NOT NULL,
			error TEXT,
			timestamp DATETIME NOT NULL,
			source TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS model_stats (
			model TEXT PRIMARY KEY,
			total_requests INTEGER NOT NULL DEFAULT 0,
			failed_requests INTEGER NOT NULL DEFAULT 0,
			total_tokens_in INTEGER NOT NULL DEFAULT 0,
			total_tokens_out INTEGER NOT NULL DEFAULT 0,
			average_latency REAL NOT NULL DEFAULT 0,
			last_used DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS model_stats_history (
			id TEXT PRIMARY KEY,
			model TEXT NOT NULL,
			total_requests INTEGER NOT NULL,
			failed_requests INTEGER NOT NULL,
			total_tokens_in INTEGER NOT NULL,
			total_tokens_out INTEGER NOT NULL,
			average_latency REAL NOT NULL,
			timestamp DATETIME NOT NULL
		);
	`)
	return err
}

// GetAllRequests retrieves all requests
func (s *SQLiteStorage) GetAllRequests() ([]*types.Request, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, model, prompt, response, tokens_in, tokens_out, server, latency_ms, status, error, timestamp, source
		FROM requests
		ORDER BY timestamp DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*types.Request
	for rows.Next() {
		var req types.Request
		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.Model,
			&req.Prompt,
			&req.Response,
			&req.TokensIn,
			&req.TokensOut,
			&req.Server,
			&req.LatencyMs,
			&req.Status,
			&req.Error,
			&req.Timestamp,
			&req.Source,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}
	return requests, nil
}

// GetRecentRequests retrieves the most recent requests
func (s *SQLiteStorage) GetRecentRequests(limit int) ([]*types.Request, error) {
	return s.ListRequests(limit)
}

// GetRequestByID retrieves a specific request by ID
func (s *SQLiteStorage) GetRequestByID(requestID string) (*types.Request, error) {
	return s.GetRequest(requestID)
}

// GetRequests retrieves all requests for a specific user
func (s *SQLiteStorage) GetRequests(userID string) ([]*types.Request, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, model, prompt, response, tokens_in, tokens_out, server, latency_ms, status, error, timestamp, source
		FROM requests
		WHERE user_id = ?
		ORDER BY timestamp DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*types.Request
	for rows.Next() {
		var req types.Request
		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.Model,
			&req.Prompt,
			&req.Response,
			&req.TokensIn,
			&req.TokensOut,
			&req.Server,
			&req.LatencyMs,
			&req.Status,
			&req.Error,
			&req.Timestamp,
			&req.Source,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}
	return requests, nil
}
