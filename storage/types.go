package storage

import (
	"llm-fw/handlers"
	"time"
)

// Request defines the structure for request records
type Request struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Model     string    `json:"model"`
	Prompt    string    `json:"prompt"`
	Response  string    `json:"response"`
	TokensIn  int       `json:"tokens_in"`
	TokensOut int       `json:"tokens_out"`
	Server    string    `json:"server"`
	Timestamp time.Time `json:"timestamp"`
}

// ToHandlerRequest converts storage.Request to handlers.Request
func (r *Request) ToHandlerRequest() *handlers.Request {
	return &handlers.Request{
		ID:        r.ID,
		UserID:    r.UserID,
		Model:     r.Model,
		Prompt:    r.Prompt,
		Response:  r.Response,
		TokensIn:  r.TokensIn,
		TokensOut: r.TokensOut,
		Server:    r.Server,
	}
}

// FromHandlerRequest converts handlers.Request to storage.Request
func FromHandlerRequest(r *handlers.Request) *Request {
	return &Request{
		ID:        r.ID,
		UserID:    r.UserID,
		Model:     r.Model,
		Prompt:    r.Prompt,
		Response:  r.Response,
		TokensIn:  r.TokensIn,
		TokensOut: r.TokensOut,
		Server:    r.Server,
		Timestamp: time.Now(),
	}
}
