package orchestrator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type WorkerPayload struct {
	Type string           `json:"type"`
	Args *json.RawMessage `json:"args"`
}

func sendToWorker(workerUrl string, job *Job) (string, error) {
	payload, _ := json.Marshal(WorkerPayload{
		Type: job.Type,
		Args: &job.Args,
	})
	resp, err := http.Post(workerUrl, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("worker unreachable: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var workerResp WorkerResponse
	if err = json.Unmarshal(bodyBytes, &workerResp); err != nil {
		return "", fmt.Errorf("invalid response from worker: %s", string(bodyBytes))
	}

	if workerResp.Error != "" {
		return "", fmt.Errorf("execution error: %s", workerResp.Error)
	}
	return workerResp.Output, nil
}
