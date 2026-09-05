// types.go
package main

import (
	"context"
	"net/http"
)

type ImageProvider interface {
	Name() string
	Generate(ctx context.Context, req *ImageGenRequest) ([]string, error)
}

type ImageGenRequest struct {
	Prompt        string `json:"prompt"`
	NumImages     int    `json:"num_images"`
	Img2imgBase64 string `json:"img2img_base64,omitempty"`
	Transparent   bool   `json:"transparent"`
	ResponseFormat string `json:"response_format,omitempty"`
}
