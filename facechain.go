package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

const (
	facechainAPI = "https://dashscope.aliyuncs.com/api/v1/services/aigc/album/gen_potrait"
	dashscopeTaskAPI = "https://dashscope.aliyuncs.com/api/v1/tasks/"
	dashscopeAPIKey  = "YOUR_DASHSCOPE_API_KEY"
)

type facechainReq struct {
	Model      string              `json:"model"`
	Parameters facechainParams     `json:"parameters"`
	Input      facechainInput      `json:"input"`
}

type facechainParams struct {
	Style       string `json:"style"`
	N           int    `json:"n"`
	SkinRetouch bool   `json:"skin_retouch"`
}

type facechainInput struct {
	TemplateURL string   `json:"template_url,omitempty"`
	UserURLs    []string `json:"user_urls"`
}

type facechainSubmitResp struct {
	Output struct {
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
	} `json:"output"`
	RequestID string `json:"request_id"`
}

type facechainTaskResp struct {
	Output struct {
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
		Results    []struct {
			URL string `json:"url"`
		} `json:"results"`
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"output"`
}

// callFaceChain submits a user photo COS URL to FaceChain (train-free mode) and returns the result image URL.
// gender: "male" or "female" — used to select the right template.
// If templateURL is empty, returns an error (templates must be configured).
func callFaceChain(userCOSURL string, gender string, templateURL string) (string, error) {
	if templateURL == "" {
		return "", fmt.Errorf("FaceChain模板未配置")
	}

	logger.Printf("FaceChain(train-free): user=%s, gender=%s, template=%s", userCOSURL, gender, templateURL)

	payload := facechainReq{
		Model: "facechain-generation",
		Parameters: facechainParams{
			Style:       "train_free_portrait_url_template",
			N:           1,
			SkinRetouch: true,
		},
		Input: facechainInput{
			TemplateURL: templateURL,
			UserURLs:    []string{userCOSURL},
		},
	}

	bodyBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", facechainAPI, bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+dashscopeAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DashScope-Async", "enable")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("FaceChain提交失败: %w", err)
	}
	defer resp.Body.Close()

	var submitResp facechainSubmitResp
	if err := json.NewDecoder(resp.Body).Decode(&submitResp); err != nil {
		return "", fmt.Errorf("解析FaceChain响应失败: %w", err)
	}

	taskID := submitResp.Output.TaskID
	if taskID == "" {
		return "", fmt.Errorf("FaceChain未返回task_id")
	}
	logger.Printf("FaceChain task: %s", taskID)

	// Poll for result (max 3 minutes)
	for i := 0; i < 90; i++ {
		time.Sleep(2 * time.Second)

		req, _ := http.NewRequest("GET", dashscopeTaskAPI+taskID, nil)
		req.Header.Set("Authorization", "Bearer "+dashscopeAPIKey)

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		var taskResp facechainTaskResp
		json.NewDecoder(resp.Body).Decode(&taskResp)
		resp.Body.Close()

		switch taskResp.Output.TaskStatus {
		case "SUCCEEDED":
			if len(taskResp.Output.Results) == 0 {
				return "", fmt.Errorf("FaceChain成功但无结果")
			}
			resultURL := taskResp.Output.Results[0].URL
			logger.Printf("FaceChain成功: %s", resultURL)
			return resultURL, nil
		case "FAILED":
			return "", fmt.Errorf("FaceChain失败: %s", taskResp.Output.Message)
		}
	}

	return "", fmt.Errorf("FaceChain超时")
}

// downloadImage downloads an image from a URL and returns the bytes.
func downloadImage(url string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("下载图片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载图片HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// uploadFileToCOS reads a multipart.File and uploads it to COS, returning the public URL.
func uploadFileToCOS(file multipart.File, filename string) (string, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}
	key := generateCOSKey("uploads", filename)
	mimeType := "image/jpeg"
	return cosUpload(data, key, mimeType)
}
