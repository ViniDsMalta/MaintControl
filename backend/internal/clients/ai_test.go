package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"MaintControl/internal/models"
)

func TestAIClientPredict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/predict" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var request struct {
			MachineType string                    `json:"machine_type"`
			Readings    []models.TelemetryReading `json:"readings"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.MachineType != "motor" || len(request.Readings) != 30 {
			t.Fatalf("unexpected AI payload: %+v", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"prediction":"NORMAL","probabilities":{"NORMAL":0.9,"DEGRADATION":0.08,"FAILURE":0.02}}`))
	}))
	defer server.Close()

	client := NewAIClient(server.URL, time.Second)
	readings := make([]models.TelemetryReading, 30)
	prediction, err := client.Predict(context.Background(), "motor", readings)
	if err != nil {
		t.Fatalf("predict failed: %v", err)
	}
	if prediction.Prediction != models.StatusNormal || prediction.Probabilities[models.StatusFailure] != 0.02 {
		t.Fatalf("unexpected prediction: %+v", prediction)
	}
}
