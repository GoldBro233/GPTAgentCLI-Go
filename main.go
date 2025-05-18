package main

import (
	"bufio"
	"encoding/json" // 用于JSON编码和解码
	"fmt"
	"io/ioutil" // 用于读取响应体
	"net/http"  // 用于发送HTTP请求
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var apiKey string   // API密钥变量
var provider string // 用于指定API提供者
var model string    // 用于指定模型名称
var question string // 用于指定查询内容

func main() {
	rootCmd := &cobra.Command{
		Use:   "chatgpt-agent [query]",
		Short: "ChatGPT或Grok Agent CLI工具",
		Run: func(cmd *cobra.Command, args []string) {
			var query string

			// 第一优先：使用 --ques 或 -q 标志
			if question != "" {
				query = question
				// 检查并追加管道输入
				scanner := bufio.NewScanner(os.Stdin)
				var pipeLines []string
				for scanner.Scan() {
					pipeLines = append(pipeLines, scanner.Text())
				}
				if err := scanner.Err(); err == nil && len(pipeLines) > 0 {
					query += "\n" + strings.Join(pipeLines, "\n") // 将管道内容追加到查询末尾
				}
			} else if len(args) > 0 {
				// 第二优先：使用位置参数
				query = strings.Join(args, " ")
			} else {
				// 第三优先：从标准输入（包括管道）读取所有内容
				scanner := bufio.NewScanner(os.Stdin)
				var lines []string
				for scanner.Scan() {
					lines = append(lines, scanner.Text())
				}
				if err := scanner.Err(); err != nil {
					fmt.Println("读取标准输入错误:", err)
					return
				}
				query = strings.Join(lines, "\n") // 使用换行符连接多行
			}

			if query == "" {
				fmt.Println("错误：查询不能为空")
				return
			}

			if provider == "grok" {
				// Grok API调用逻辑
				url := "https://api.x.ai/v1/chat/completions" // Grok API端点，需替换为实际值
				effectiveModel := model // 使用用户提供的模型，如果为空则设置默认值
				if effectiveModel == "" {
					effectiveModel = "grok" // 默认模型为"grok"
				}
				payload := map[string]interface{}{
					"model": effectiveModel, // 使用指定的模型
					"messages": []map[string]string{
						{"role": "user", "content": query},
					},
					// 可能需要更多字段，如max_tokens
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					fmt.Println("JSON编码错误:", err)
					return
				}

				client := &http.Client{}
				req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonPayload)))
				if err != nil {
					fmt.Println("请求创建错误:", err)
					return
				}
				req.Header.Set("Authorization", "Bearer "+apiKey) // 设置Authorization头部
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				if err != nil {
					fmt.Println("API调用失败:", err)
					return
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println("响应读取错误:", err)
					return
				}

				var result map[string]interface{}
				err = json.Unmarshal(body, &result)
				if err != nil {
					fmt.Println("JSON解码错误:", err)
					return
				}

				// 提取响应内容
				if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
					if message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string); ok {
						fmt.Println(message) // 输出响应内容
					}
				}
			} else {
				// 如果provider不是"grok"，可以添加更多逻辑或其他provider的处理
				fmt.Println("错误：不支持的提供者。请使用 --provider chatgpt 或 grok")
			}
		},
	}

	rootCmd.Flags().StringVarP(&apiKey, "key", "k", "", "API密钥")
	rootCmd.Flags().StringVarP(&provider, "provider", "p", "chatgpt", "API提供者: chatgpt 或 grok")
	rootCmd.Flags().StringVarP(&model, "model", "m", "", "指定模型名称，例如 gpt-3.5-turbo 或 grok")
	rootCmd.Flags().StringVarP(&question, "ques", "q", "", "指定查询内容")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}