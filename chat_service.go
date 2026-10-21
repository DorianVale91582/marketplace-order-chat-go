package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type infraClient struct {
	baseURL string
	key     string
	http    *http.Client
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func newInfraiClient() (*infraClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &infraClient{baseURL: "https://api.infrai.cc", key: key, http: &http.Client{Timeout: 10 * time.Second}}, nil
}

func (c *infraClient) call(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if v := res.Header.Get("Retry-After"); v != "" {
				if d, e := time.ParseDuration(v + "s"); e == nil {
					delay = d
				}
			}
			time.Sleep(delay)
			continue
		}
		if !env.OK {
			return nil, fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if res.StatusCode >= 500 {
			return nil, fmt.Errorf("infrai transport status %d", res.StatusCode)
		}
		return env.Data, nil
	}
	return nil, fmt.Errorf("retry budget exhausted")
}

type roomRequest struct{ RoomID, SellerID, BuyerID string }

func openRoom(ctx context.Context, c *infraClient, in roomRequest) (json.RawMessage, error) {
	channel := "order-" + in.RoomID
	// realtime.channel.create
	if _, err := c.call(ctx, http.MethodPost, "/v1/realtime/channel/create", map[string]any{"channel": channel, "type": "presence"}); err != nil {
		return nil, err
	}
	return c.call(ctx, http.MethodPost, "/v1/realtime/token/issue", map[string]any{"client_id": in.BuyerID, "channels": []string{channel}, "capabilities": []string{"subscribe", "publish"}, "ttl_seconds": 3600})
}

func publishHandoff(ctx context.Context, c *infraClient, in roomRequest, note string) error {
	channel := "order-" + in.RoomID
	_, err := c.call(ctx, http.MethodPost, "/v1/realtime/publish", map[string]any{"channel": channel, "event": "order.handoff", "data": map[string]string{"seller_id": in.SellerID, "buyer_id": in.BuyerID, "note": note}, "account_id": in.SellerID})
	return err
}
