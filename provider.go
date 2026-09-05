// provider.go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type ImageGenRequest struct {
	Prompt        string `json:"prompt"`
	NumImages     int    `json:"num_images"`
	Img2imgBase64 string `json:"img2img_base64,omitempty"`
	Transparent   bool   `json:"transparent"`
	ResponseFormat string `json:"response_format,omitempty"`
}

type ImageProvider interface {
	Name() string
	Generate(ctx context.Context, req *ImageGenRequest) ([]string, error)
}

type GatewayProvider struct {
	name     string
	endpoint string
	apiKey   string
	model    string
}

func NewGatewayProvider(name, endpoint, apiKey, model string) ImageProvider {
	if endpoint == "" {
		endpoint = "http://127.0.0.1:50066/v1"
	}
	if name == "" {
		name = "Grok2API"
	}
	return &GatewayProvider{
		name:     name,
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    model,
	}
}

func (g *GatewayProvider) Name() string {
	return g.name
}

// ==================== 自动获取模型 ====================
func (g *GatewayProvider) getAutoModel() string {
	if g.model != "" {
		return g.model
	}

	modelsURL := strings.TrimSuffix(g.endpoint, "/") + "/models"
	client := &http.Client{Timeout: 10 * time.Second}

	req, _ := http.NewRequest("GET", modelsURL, nil)
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("⚠️ 无法获取模型列表: %v", err)
		return "quikronimage-2.0"
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Data) > 0 {
		log.Printf("🧠 自动获取到模型: %s", result.Data[0].ID)
		return result.Data[0].ID
	}
	return "quikronimage-2.0"
}

// ==================== 生成逻辑 ====================
func (g *GatewayProvider) Generate(ctx context.Context, req *ImageGenRequest) ([]string, error) {
	log.Printf("🔄 [Grok2API] 开始生成图片，模型: %s", g.model)

	// 自动获取模型（如果未指定）
	if g.model == "" {
		g.model = g.getAutoModel()
	}

	// 设置响应格式（透明图使用 b64_json）
	if req.ResponseFormat == "" {
		if req.Transparent {
			req.ResponseFormat = "b64_json"
		} else {
			req.ResponseFormat = "url"
		}
	}

	url := strings.TrimSuffix(g.endpoint, "/") + "/image_generations"

	payload := map[string]interface{}{
		"model":           g.model,
		"prompt":          req.Prompt,
		"n":               req.NumImages,
		"quality":         "standard",
		"response_format": req.ResponseFormat,
	}

	if req.Img2imgBase64 != "" {
		payload["image"] = req.Img2imgBase64
	}

	jsonData, _ := json.Marshal(payload)

	reqHTTP, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	reqHTTP.Header.Set("Authorization", "Bearer "+g.apiKey)
	reqHTTP.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		return nil, fmt.Errorf("Grok2API 调用失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("📨 [Grok2API] 返回: %s", string(body[:min(800, len(body))]))

	var result struct {
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
		Data []struct {
			URL     string `json:"url"`
			B64JSON string `json:"b64_json,omitempty"`
		} `json:"data"`
		Output []string `json:"output"`
	}

	json.Unmarshal(body, &result)

	var urls []string
	if req.Transparent {
		for _, item := range result.Data {
			if item.B64JSON != "" {
				urls = append(urls, "data:image/png;base64,"+item.B64JSON)
			}
		}
	} else {
		for _, item := range result.Images {
			if item.URL != "" {
				urls = append(urls, item.URL)
			}
		}
		for _, item := range result.Data {
			if item.URL != "" {
				urls = append(urls, item.URL)
			}
		}
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("未从 Grok2API 获取到有效图片链接")
	}

	return urls, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
