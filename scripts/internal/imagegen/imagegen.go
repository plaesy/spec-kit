// Package imagegen generates image assets via a configured provider
// (OpenAI or Gemini) and writes the decoded bytes to disk. It ports
// scripts/bash/create-image.sh using only the Go standard library.
package imagegen

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Options configures an image generation request.
type Options struct {
	Prompt   string
	Provider string // "openai" (default) or "gemini"
	Size     string // e.g. "1024x1024"
	Out      string // output file path
}

const defaultSize = "1024x1024"

// Generate calls the configured provider's image-generation API, decodes
// the returned base64 image, and writes it to opts.Out (creating parent
// directories as needed). It returns the output path on success.
func Generate(opts Options) (string, error) {
	if opts.Prompt == "" || opts.Out == "" {
		return "", fmt.Errorf("prompt and out are required")
	}

	provider := opts.Provider
	if provider == "" {
		provider = os.Getenv("PLAESY_IMAGE_PROVIDER")
	}
	if provider == "" {
		provider = "openai"
	}

	size := opts.Size
	if size == "" {
		size = defaultSize
	}

	if err := os.MkdirAll(filepath.Dir(opts.Out), 0o755); err != nil {
		return "", fmt.Errorf("creating output directory: %w", err)
	}

	var data []byte
	var err error

	switch provider {
	case "openai":
		data, err = generateOpenAI(opts.Prompt, size)
	case "gemini":
		data, err = generateGemini(opts.Prompt)
	default:
		return "", fmt.Errorf("unknown provider %q (supported: openai, gemini)", provider)
	}
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(opts.Out, data, 0o644); err != nil {
		return "", fmt.Errorf("writing output file: %w", err)
	}

	info, err := os.Stat(opts.Out)
	if err != nil || info.Size() == 0 {
		return "", fmt.Errorf("%s was not written or is empty", opts.Out)
	}

	return opts.Out, nil
}

type apiError struct {
	Message string `json:"message"`
}

func generateOpenAI(prompt, size string) ([]byte, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is not set. Export it, e.g.:\n  export OPENAI_API_KEY=sk-...")
	}

	reqBody, err := json.Marshal(map[string]any{
		"model":  "gpt-image-1",
		"prompt": prompt,
		"size":   size,
		"n":      1,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/images/generations", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	respBody, err := doRequest(req)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Error *apiError `json:"error"`
		Data  []struct {
			B64JSON string `json:"b64_json"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("parsing OpenAI response: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return nil, fmt.Errorf("OpenAI image API: %s", parsed.Error.Message)
	}
	if len(parsed.Data) == 0 || parsed.Data[0].B64JSON == "" {
		return nil, fmt.Errorf("OpenAI image API: no image data returned")
	}

	return base64.StdEncoding.DecodeString(parsed.Data[0].B64JSON)
}

func generateGemini(prompt string) ([]byte, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not set. Export it, e.g.:\n  export GEMINI_API_KEY=...")
	}

	reqBody, err := json.Marshal(map[string]any{
		"instances":  []map[string]any{{"prompt": prompt}},
		"parameters": map[string]any{"sampleCount": 1},
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/imagen-3.0-generate-002:predict?key=%s", apiKey)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	respBody, err := doRequest(req)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Error       *apiError `json:"error"`
		Predictions []struct {
			BytesBase64Encoded string `json:"bytesBase64Encoded"`
		} `json:"predictions"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("parsing Gemini response: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return nil, fmt.Errorf("Gemini image API: %s", parsed.Error.Message)
	}
	if len(parsed.Predictions) == 0 || parsed.Predictions[0].BytesBase64Encoded == "" {
		return nil, fmt.Errorf("Gemini image API: no image data returned")
	}

	return base64.StdEncoding.DecodeString(parsed.Predictions[0].BytesBase64Encoded)
}

func doRequest(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return body, nil
}
