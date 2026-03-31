package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type PluginSettings struct {
	Path    string                `json:"path"`
	Secrets *SecretPluginSettings `json:"-"`
}

type SecretPluginSettings struct {
	UserOCID string `json:"UserOCID"`
	TenancyOCID string `json:"TenancyOCID"`
	Fingerprint string `json:"Fingerprint"`
	Region string `json:"Region"`
	PrivateKey string `json:"PrivateKey"`
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	settings := PluginSettings{}
	err := json.Unmarshal(source.JSONData, &settings)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal PluginSettings json: %w", err)
	}

	settings.Secrets = loadSecretPluginSettings(source.DecryptedSecureJSONData)

	return &settings, nil
}

func loadSecretPluginSettings(source map[string]string) *SecretPluginSettings {
	return &SecretPluginSettings{
		UserOCID: source["UserOCID"],
		TenancyOCID: source["TenancyOCID"],
		Fingerprint: source["Fingerprint"],
		Region: source["Region"],
		PrivateKey: source["PrivateKey"],
	}
}
