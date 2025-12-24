package orchestrator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

func sendCodeToWorker(workerURL string, code string) (string, error) {
	if err := waitForWorker(workerURL); err != nil {
		return "", fmt.Errorf("worker failed to become ready: %w", err)
	}

	payload, _ := json.Marshal(map[string]string{"code": code})
	resp, err := http.Post(workerURL, "application/json", bytes.NewBuffer(payload))
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

func waitForWorker(url string) error {
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return errors.New("timeout waiting for server to start")
		case <-ticker.C:
			resp, err := http.Get(url)
			if err == nil {
				resp.Body.Close()
				return nil
			}
		}
	}
}
