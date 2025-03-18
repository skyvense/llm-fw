package storage

import (
	"database/sql"
	"fmt"

	"llm-fw/common"
	"llm-fw/types"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage 实现了 Storage 接口
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage 创建一个新的 SQLite 存储
func NewSQLiteStorage(dbPath string) (types.Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 创建表
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("创建表失败: %v", err)
	}

	return &SQLiteStorage{
		db: db,
	}, nil
}

// createTables 创建必要的数据库表
func createTables(db *sql.DB) error {
	// 创建请求表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS requests (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			model TEXT NOT NULL,
			prompt TEXT NOT NULL,
			response TEXT,
			tokens_in INTEGER,
			tokens_out INTEGER,
			latency_ms INTEGER,
			status INTEGER,
			error TEXT,
			timestamp DATETIME NOT NULL,
			server TEXT,
			source TEXT
		)
	`)
	if err != nil {
		return err
	}

	// 创建模型统计表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS model_stats (
			model TEXT PRIMARY KEY,
			total_requests INTEGER DEFAULT 0,
			total_tokens_in INTEGER DEFAULT 0,
			total_tokens_out INTEGER DEFAULT 0,
			average_latency REAL DEFAULT 0,
			failed_requests INTEGER DEFAULT 0,
			last_used DATETIME
		)
	`)
	return err
}

// SaveRequest 保存请求记录
func (s *SQLiteStorage) SaveRequest(req *common.Request) error {
	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 插入请求记录
	_, err = tx.Exec(`
		INSERT INTO requests (
			id, user_id, model, prompt, response, tokens_in, tokens_out,
			latency_ms, status, error, timestamp, server, source
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		req.ID, req.UserID, req.Model, req.Prompt, req.Response,
		req.TokensIn, req.TokensOut, req.LatencyMs, req.Status,
		req.Error, req.Timestamp, req.Server, req.Source)
	if err != nil {
		return fmt.Errorf("插入请求记录失败: %v", err)
	}

	// 更新模型统计
	if err := s.updateModelStats(tx, req); err != nil {
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// updateModelStats 更新模型统计信息
func (s *SQLiteStorage) updateModelStats(tx *sql.Tx, req *common.Request) error {
	// 更新或插入模型统计
	_, err := tx.Exec(`
		INSERT INTO model_stats (
			model, total_requests, total_tokens_in, total_tokens_out,
			average_latency, failed_requests, last_used
		) VALUES (?, 1, ?, ?, ?, ?, ?)
		ON CONFLICT(model) DO UPDATE SET
			total_requests = total_requests + 1,
			total_tokens_in = total_tokens_in + ?,
			total_tokens_out = total_tokens_out + ?,
			average_latency = (average_latency * total_requests + ?) / (total_requests + 1),
			failed_requests = failed_requests + ?,
			last_used = ?
	`,
		req.Model, req.TokensIn, req.TokensOut, req.LatencyMs,
		req.Status, req.Timestamp, req.TokensIn, req.TokensOut,
		req.LatencyMs, req.Status, req.Timestamp)
	return err
}

// GetModelStats 获取模型统计信息
func (s *SQLiteStorage) GetModelStats(model string) (*types.ModelStats, error) {
	var stats types.ModelStats
	err := s.db.QueryRow(`
		SELECT total_requests, total_tokens_in, total_tokens_out,
			average_latency, failed_requests, last_used
		FROM model_stats
		WHERE model = ?
	`, model).Scan(&stats.TotalRequests, &stats.TotalTokensIn,
		&stats.TotalTokensOut, &stats.AverageLatency, &stats.FailedRequests,
		&stats.LastUsed)
	if err == sql.ErrNoRows {
		return &types.ModelStats{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询模型统计失败: %v", err)
	}
	return &stats, nil
}

// GetAllModelStats 获取所有模型的统计信息
func (s *SQLiteStorage) GetAllModelStats() (map[string]*types.ModelStats, error) {
	rows, err := s.db.Query(`
		SELECT model, total_requests, total_tokens_in, total_tokens_out,
			average_latency, failed_requests, last_used
		FROM model_stats
	`)
	if err != nil {
		return nil, fmt.Errorf("查询模型统计失败: %v", err)
	}
	defer rows.Close()

	stats := make(map[string]*types.ModelStats)
	for rows.Next() {
		var model string
		var s types.ModelStats
		err := rows.Scan(&model, &s.TotalRequests, &s.TotalTokensIn,
			&s.TotalTokensOut, &s.AverageLatency, &s.FailedRequests,
			&s.LastUsed)
		if err != nil {
			return nil, fmt.Errorf("扫描模型统计失败: %v", err)
		}
		stats[model] = &s
	}
	return stats, nil
}

// GetRecentRequests 获取最近的请求记录
func (s *SQLiteStorage) GetRecentRequests(limit int) ([]*common.Request, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, model, prompt, response, tokens_in, tokens_out,
			latency_ms, status, error, timestamp, server, source
		FROM requests
		ORDER BY timestamp DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("查询最近请求失败: %v", err)
	}
	defer rows.Close()

	var requests []*common.Request
	for rows.Next() {
		var req common.Request
		err := rows.Scan(&req.ID, &req.UserID, &req.Model, &req.Prompt,
			&req.Response, &req.TokensIn, &req.TokensOut, &req.LatencyMs,
			&req.Status, &req.Error, &req.Timestamp, &req.Server,
			&req.Source)
		if err != nil {
			return nil, fmt.Errorf("扫描请求记录失败: %v", err)
		}
		requests = append(requests, &req)
	}
	return requests, nil
}

// GetRequests 获取指定用户的所有请求
func (s *SQLiteStorage) GetRequests(userID string) ([]*common.Request, error) {
	rows, err := s.db.Query(`
		SELECT id, model, prompt, response, tokens_in, tokens_out,
			latency_ms, status, error, timestamp, server, source
		FROM requests
		WHERE user_id = ?
		ORDER BY timestamp DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户请求失败: %v", err)
	}
	defer rows.Close()

	var requests []*common.Request
	for rows.Next() {
		var req common.Request
		err := rows.Scan(&req.ID, &req.Model, &req.Prompt, &req.Response,
			&req.TokensIn, &req.TokensOut, &req.LatencyMs, &req.Status,
			&req.Error, &req.Timestamp, &req.Server, &req.Source)
		if err != nil {
			return nil, fmt.Errorf("扫描请求记录失败: %v", err)
		}
		req.UserID = userID
		requests = append(requests, &req)
	}
	return requests, nil
}

// GetAllRequests 获取所有请求
func (s *SQLiteStorage) GetAllRequests() ([]*common.Request, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, model, prompt, response, tokens_in, tokens_out,
			latency_ms, status, error, timestamp, server, source
		FROM requests
		ORDER BY timestamp DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("查询所有请求失败: %v", err)
	}
	defer rows.Close()

	var requests []*common.Request
	for rows.Next() {
		var req common.Request
		err := rows.Scan(&req.ID, &req.UserID, &req.Model, &req.Prompt,
			&req.Response, &req.TokensIn, &req.TokensOut, &req.LatencyMs,
			&req.Status, &req.Error, &req.Timestamp, &req.Server,
			&req.Source)
		if err != nil {
			return nil, fmt.Errorf("扫描请求记录失败: %v", err)
		}
		requests = append(requests, &req)
	}
	return requests, nil
}

// GetRequestByID 根据ID获取请求
func (s *SQLiteStorage) GetRequestByID(requestID string) (*common.Request, error) {
	var req common.Request
	err := s.db.QueryRow(`
		SELECT id, user_id, model, prompt, response, tokens_in, tokens_out,
			latency_ms, status, error, timestamp, server, source
		FROM requests
		WHERE id = ?
	`, requestID).Scan(&req.ID, &req.UserID, &req.Model, &req.Prompt,
		&req.Response, &req.TokensIn, &req.TokensOut, &req.LatencyMs,
		&req.Status, &req.Error, &req.Timestamp, &req.Server,
		&req.Source)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("请求未找到")
	}
	if err != nil {
		return nil, fmt.Errorf("查询请求失败: %v", err)
	}
	return &req, nil
}

// DeleteRequest 删除请求
func (s *SQLiteStorage) DeleteRequest(requestID string) error {
	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 获取请求信息
	var req common.Request
	err = tx.QueryRow(`
		SELECT model, tokens_in, tokens_out, latency_ms, status
		FROM requests
		WHERE id = ?
	`, requestID).Scan(&req.Model, &req.TokensIn, &req.TokensOut,
		&req.LatencyMs, &req.Status)
	if err == sql.ErrNoRows {
		return fmt.Errorf("请求未找到")
	}
	if err != nil {
		return fmt.Errorf("查询请求失败: %v", err)
	}

	// 删除请求记录
	_, err = tx.Exec("DELETE FROM requests WHERE id = ?", requestID)
	if err != nil {
		return fmt.Errorf("删除请求记录失败: %v", err)
	}

	// 更新模型统计
	_, err = tx.Exec(`
		UPDATE model_stats SET
			total_requests = total_requests - 1,
			total_tokens_in = total_tokens_in - ?,
			total_tokens_out = total_tokens_out - ?,
			average_latency = average_latency - ?,
			failed_requests = failed_requests - ?
		WHERE model = ?
	`, req.TokensIn, req.TokensOut, req.LatencyMs, req.Status, req.Model)
	if err != nil {
		return fmt.Errorf("更新模型统计失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// Close 关闭存储连接
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// NewHistoryManager 创建一个新的历史记录管理器
func (s *SQLiteStorage) NewHistoryManager(size int) types.HistoryManager {
	return NewHistoryManager(s, size)
}
