// config.go
package main

import (
	"encoding/json"
	"os"
)

type Provider struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"apikey"`
	Model    string `json:"model"`
}

var Providers []ImageProvider

const configFile = "providers.json"

func LoadProviders() {
	// 环境变量
	addProviderFromEnv("GATEWAY_ENDPOINT", "GATEWAY_API_KEY")

	// JSON 文件
	data, err := os.ReadFile(configFile)
	if err == nil {
		var providers []Provider
		json.Unmarshal(data, &providers)
		for _, p := range providers {
			Providers = append(Providers, NewGatewayProvider(p.Name, p.Endpoint, p.APIKey, p.Model))
		}
	}
}

func addProviderFromEnv(keyEndpoint, keyKey string) {
	endpoint := os.Getenv(keyEndpoint)
	apikey := os.Getenv(keyKey)
	if endpoint != "" && apikey != "" {
		Providers = append(Providers, NewGatewayProvider("Grok2API", endpoint, apikey, ""))
	}
}

func SaveProviders() {
	data, _ := json.MarshalIndent(ProvidersToConfig(), "", "  ")
	os.WriteFile(configFile, data, 0644)
}

func ProvidersToConfig() []Provider {
	var list []Provider
	for _, p := range Providers {
		if gateway, ok := p.(*GatewayProvider); ok {
			list = append(list, Provider{
				Name:     gateway.name,
				Type:     "gateway",
				Endpoint: gateway.endpoint,
				APIKey:   gateway.apiKey,
				Model:    gateway.model,
			})
		}
	}
	return list
}
