package orchestrator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func sendCodeToWorker(workerUrl string, code string) (string, error) {
	payload, _ := json.Marshal(map[string]string{"code": code})
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
		return "", fmt.Errorf("python execution error: %s", workerResp.Error)
	}
	return workerResp.Output, nil
}
