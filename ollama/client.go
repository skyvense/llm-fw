package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	URL    string
	Client *http.Client
}

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func NewClient(url string) *Client {
	return &Client{
		URL:    url,
		Client: &http.Client{},
	}
}

func (c *Client) Generate(model, prompt string) (*GenerateResponse, error) {
	reqBody := GenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/generate", c.URL), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
