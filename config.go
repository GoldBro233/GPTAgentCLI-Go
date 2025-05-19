package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Config 结构体定义配置信息
type Config struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	APIKey       string `json:"apiKey"`
	SystemPrompt string `json:"systemPrompt"` // 系统提示字段
}

// ConfigManager 结构体用于封装配置管理逻辑
type ConfigManager struct {
	Path string // 配置文件路径
}

// NewConfigManager 创建并返回一个新的 ConfigManager 实例
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		Path: getConfigPath(),
	}
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("错误：获取应用程序路径失败: %v\n", err)
		os.Exit(1)
	}
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "config.json") // 使用 .json 扩展名
}

// LoadConfig 加载配置
func (cm *ConfigManager) LoadConfig() (Config, error) {
	var config Config
	if _, err := os.Stat(cm.Path); os.IsNotExist(err) {
		return config, fmt.Errorf("配置文件不存在")
	}
	data, err := ioutil.ReadFile(cm.Path)
	if err != nil {
		return config, fmt.Errorf("读取配置文件失败: %v", err)
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("解码JSON失败: %v", err)
	}
	return config, nil
}

// SaveConfig 保存配置
func (cm *ConfigManager) SaveConfig(config Config) error {
	data, err := json.MarshalIndent(&config, "", "  ") // 使用 json.MarshalIndent 美化输出
	if err != nil {
		return fmt.Errorf("编码JSON失败: %v", err)
	}
	dir := filepath.Dir(cm.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}
	err = ioutil.WriteFile(cm.Path, data, 0600) // 使用更严格的权限
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}
	fmt.Println("配置已保存成功！")
	return nil
}

// PromptForConfig 提示用户输入配置信息
func (cm *ConfigManager) PromptForConfig() Config {
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

// SetupConfig 检查并设置配置，首次运行时提示用户输入
func (cm *ConfigManager) SetupConfig() (Config, error) {
	config, err := cm.LoadConfig()
	if err != nil {
		if err.Error() == "配置文件不存在" {
			fmt.Println("配置文件不存在。请输入您的配置信息：")
			config = cm.PromptForConfig()
			if err := cm.SaveConfig(config); err != nil {
				return config, fmt.Errorf("保存配置文件失败: %v", err)
			}
			fmt.Println("首次配置完成，程序将退出。请重新运行程序以使用。")
			os.Exit(0)
		}
		return config, fmt.Errorf("加载配置文件失败: %v", err)
	}
	return config, nil
}
