package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var apiKey string   // API密钥变量（从配置文件加载）
var provider string // 用于指定API提供者（从配置文件加载）
var model string    // 用于指定模型名称（从配置文件加载）
var question string // 用于指定查询内容

type Config struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	APIKey       string `json:"apiKey"`
	SystemPrompt string `json:"systemPrompt"` // 新增系统提示字段
}

// isTerminal 检查文件是否是终端
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

func main() {
	configPath := getConfigPath()

	var config Config

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 创建目录（如果需要）
		dir := filepath.Dir(configPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("错误：创建配置目录失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("配置文件不存在。请输入您的配置信息：")
		config = promptForConfig()
		if err := saveConfig(configPath, config); err != nil {
			fmt.Printf("错误：保存配置文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("首次配置完成，程序将退出。请重新运行程序以使用。")
		os.Exit(0) // 首次创建配置文件后立即退出
	} else {
		var err error
		config, err = loadConfig(configPath)
		if err != nil {
			fmt.Printf("错误：加载配置文件失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 设置全局变量
	provider = config.Provider
	model = config.Model
	apiKey = config.APIKey

	rootCmd := &cobra.Command{
		Use:   "chatgpt-agent [query]",
		Short: "ChatGPT或Grok Agent CLI工具",
		Run: func(cmd *cobra.Command, args []string) {
			var query string
			var pipeContent string

			// 第一优先：使用 --ques 或 -q 标志
			if question != "" {
				query = question
				// 仅在标准输入不是终端时尝试读取管道输入
				if !isTerminal(os.Stdin) {
					scanner := bufio.NewScanner(os.Stdin)
					var pipeLines []string
					for scanner.Scan() {
						pipeLines = append(pipeLines, scanner.Text())
					}
					if err := scanner.Err(); err == nil && len(pipeLines) > 0 {
						pipeContent = strings.Join(pipeLines, "\n")
						// 尝试读取文件内容（假设管道输入可能是文件名）
						if len(pipeLines) == 1 {
							filePath := strings.TrimSpace(pipeContent)
							if fileContent, err := ioutil.ReadFile(filePath); err == nil {
								query += "\n文件内容如下:\n" + string(fileContent)
							} else {
								query += "\n管道输入内容:\n" + pipeContent
							}
						} else {
							query += "\n管道输入内容:\n" + pipeContent
						}
					}
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
					fmt.Printf("读取标准输入错误: %v\n", err)
					return
				}
				query = strings.Join(lines, "\n") // 使用换行符连接多行
			}

			if query == "" {
				fmt.Println("错误：查询不能为空")
				return
			}

			if provider == "deepseek" {
				// DeepSeek API调用逻辑
				url := "https://api.deepseek.com/chat/completions" // DeepSeek API 端点，与官方一致
				effectiveModel := model                            // 使用配置文件中的模型，如果为空则设置默认值
				if effectiveModel == "" {
					effectiveModel = "deepseek-chat" // 默认模型与官方一致
				}

				// 设置系统提示，如果配置文件中没有则使用默认值
				systemContent := "You are a helpful assistant"
				if config.SystemPrompt != "" {
					systemContent = config.SystemPrompt
				}

				// 按照官方提供的格式构建请求体
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
					"model":             effectiveModel, // 使用指定的模型
					"frequency_penalty": 0,
					"max_tokens":        2048,
					"presence_penalty":  0,
					"response_format": map[string]string{
						"type": "text",
					},
					"stop":           nil,
					"stream":         true, // 启用流式传输
					"stream_options": map[string]bool{
						"include_usage": true, // 包含使用量信息
					},
					"temperature":    1,
					"top_p":          1,
					"tools":          nil,
					"tool_choice":    "none",
					"logprobs":       false,
					"top_logprobs":   nil,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					fmt.Printf("JSON编码错误: %v\n", err)
					return
				}

				client := &http.Client{
					Timeout: 60 * time.Second, // 设置超时时间，避免无限等待
				}
				req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonPayload)))
				if err != nil {
					fmt.Printf("请求创建错误: %v\n", err)
					return
				}
				req.Header.Set("Authorization", "Bearer "+apiKey) // 设置Authorization头部
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept", "application/json") // 添加Accept头部，符合官方示例

				resp, err := client.Do(req)
				if err != nil {
					fmt.Printf("API调用失败: %v\n", err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != 200 {
					fmt.Printf("收到响应，状态码: %d\n", resp.StatusCode)
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						fmt.Printf("响应读取错误: %v\n", err)
						return
					}
					fmt.Printf("响应内容: %s\n", string(body))
					return
				}

				// 处理流式响应
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
					fmt.Printf("流式读取错误: %v\n", err)
					return
				}

				// 输出总token数
				fmt.Printf("\n\n消耗Token: %d\n", totalTokens)
			} else {
				fmt.Println("错误：不支持的提供者。请在配置文件中设置有效的提供者")
			}
		},
	}

	// 仅保留 --ques 或 -q 标志
	rootCmd.Flags().StringVarP(&question, "ques", "q", "", "指定查询内容")

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("命令执行错误: %v\n", err)
		os.Exit(1)
	}
}

// 获取配置文件路径
func getConfigPath() string {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("错误：获取应用程序路径失败: %v\n", err)
		os.Exit(1)
	}
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "config.json") // 使用 .json 扩展名
}

// 提示用户输入配置
func promptForConfig() Config {
	var config Config
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("输入提供者 (chatgpt 或 deepseek): ")
	providerInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("错误：读取提供者输入失败: %v\n", err)
		os.Exit(1)
	}
	config.Provider = strings.TrimSpace(providerInput)

	fmt.Print("输入模型名称 (例如 deepseek-chat): ")
	modelInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("错误：读取模型名称输入失败: %v\n", err)
		os.Exit(1)
	}
	config.Model = strings.TrimSpace(modelInput)

	fmt.Print("输入 API 密钥: ")
	apiKeyInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("错误：读取API密钥输入失败: %v\n", err)
		os.Exit(1)
	}
	config.APIKey = strings.TrimSpace(apiKeyInput)

	fmt.Print("输入系统提示 (按回车使用默认值 'You are a helpful assistant'): ")
	systemPromptInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("错误：读取系统提示输入失败: %v\n", err)
		os.Exit(1)
	}
	systemPrompt := strings.TrimSpace(systemPromptInput)
	if systemPrompt != "" {
		config.SystemPrompt = systemPrompt
	}

	return config
}

// 保存配置到文件
func saveConfig(path string, config Config) error {
	data, err := json.MarshalIndent(&config, "", "  ") // 使用 json.MarshalIndent 美化输出
	if err != nil {
		return fmt.Errorf("编码JSON失败: %v", err)
	}
	err = ioutil.WriteFile(path, data, 0600) // 使用更严格的权限
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}
	fmt.Println("配置已保存成功！")
	return nil
}

// 加载配置从文件
func loadConfig(path string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("读取配置文件失败: %v", err)
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("解码JSON失败: %v", err)
	}
	return config, nil
}