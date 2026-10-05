package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"MaintControl/internal/models"
)

type AIClient struct {
	baseURL string
	http    *http.Client
}

func NewAIClient(baseURL string, timeout time.Duration) *AIClient {
	return &AIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *AIClient) Predict(ctx context.Context, machineType string, readings []models.TelemetryReading) (models.AIPrediction, error) {
	payload := struct {
		MachineType string                    `json:"machine_type"`
		Readings    []models.TelemetryReading `json:"readings"`
	}{MachineType: machineType, Readings: readings}

	body, err := json.Marshal(payload)
	if err != nil {
		return models.AIPrediction{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/predict", bytes.NewReader(body))
	if err != nil {
		return models.AIPrediction{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := c.http.Do(req)
	if err != nil {
		return models.AIPrediction{}, fmt.Errorf("ai service request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return models.AIPrediction{}, fmt.Errorf("ai service returned status %d", response.StatusCode)
	}

	var prediction models.AIPrediction
	decoder := json.NewDecoder(response.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&prediction); err != nil {
		return models.AIPrediction{}, fmt.Errorf("decode ai response: %w", err)
	}
	if prediction.Prediction == "" || prediction.Probabilities == nil {
		return models.AIPrediction{}, errors.New("incomplete ai response")
	}

	return prediction, nil
}
