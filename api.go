package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// APIClient 结构体用于封装API调用逻辑
type APIClient struct {
	APIKey string
	URL    string
}

// NewAPIClient 创建并返回一个新的 APIClient 实例
func NewAPIClient(apiKey string) *APIClient {
	return &APIClient{
		APIKey: apiKey,
		URL:    "https://api.deepseek.com/chat/completions",
	}
}

// SendRequest 发送API请求并处理响应
func (ac *APIClient) SendRequest(query, model, systemPrompt string) error {
	effectiveModel := model
	if effectiveModel == "" {
		effectiveModel = "deepseek-chat" // 默认模型
	}

	// 设置系统提示，如果为空则使用默认值
	systemContent := "You are a helpful assistant"
	if systemPrompt != "" {
		systemContent = systemPrompt
	}

	// 构建请求体
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{
				"content": systemContent,
				"role":    "system",
			},
			{
				"content": query,
				"role":    "user",
			},
		},
		"model":             effectiveModel,
		"frequency_penalty": 0,
		"max_tokens":        8192,
		"presence_penalty":  0,
		"response_format": map[string]string{
			"type": "text",
		},
		"stop":   nil,
		"stream": true,
		"stream_options": map[string]bool{
			"include_usage": true,
		},
		"temperature":  1,
		"top_p":        1,
		"tools":        nil,
		"tool_choice":  "none",
		"logprobs":     false,
		"top_logprobs": nil,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("JSON编码错误: %v", err)
	}

	client := &http.Client{
		Timeout: 60 * time.Second, // 设置超时时间，避免无限等待
	}
	req, err := http.NewRequest("POST", ac.URL, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return fmt.Errorf("请求创建错误: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+ac.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("API调用失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("收到响应，状态码: %d\n", resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("响应读取错误: %v", err)
		}
		fmt.Printf("响应内容: %s\n", string(body))
		return nil
	}

	// 处理流式响应
	return ac.handleStreamResponse(resp)
}

// handleStreamResponse 处理流式响应
func (ac *APIClient) handleStreamResponse(resp *http.Response) error {
	scanner := bufio.NewScanner(resp.Body)
	var fullContent strings.Builder
	var totalTokens int

	fmt.Println("回复内容：")
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// 去掉 "data: " 前缀
		jsonData := strings.TrimPrefix(line, "data: ")
		if jsonData == "[DONE]" {
			break
		}

		var streamResp map[string]interface{}
		err := json.Unmarshal([]byte(jsonData), &streamResp)
		if err != nil {
			fmt.Printf("流式JSON解码错误: %v\n", err)
			continue
		}

		// 检查是否有使用量信息
		if usage, ok := streamResp["usage"].(map[string]interface{}); ok {
			if tokens, ok := usage["total_tokens"].(float64); ok {
				totalTokens = int(tokens)
			}
		}

		// 提取内容增量
		if choices, ok := streamResp["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if content, ok := delta["content"].(string); ok {
						fmt.Print(content) // 实时输出内容
						fullContent.WriteString(content)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("流式读取错误: %v", err)
	}

	// 输出总token数
	fmt.Printf("\n\n消耗Token: %d\n", totalTokens)
	return nil
}
