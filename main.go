package main

import (
	"fmt"
	"os"
)

func main() {
	// 初始化配置管理器
	configManager := NewConfigManager()
	config, err := configManager.SetupConfig()
	if err != nil {
		fmt.Printf("错误：%v\n", err)
		os.Exit(1)
	}

	// 初始化API客户端
	apiClient := NewAPIClient(config.APIKey)

	// 初始化命令行处理器
	cmdHandler := NewCmdHandler()

	// 定义查询处理回调函数
	queryHandler := func(query string) {
		if query == "" {
			return // 查询为空，直接返回
		}

		if config.Provider == "deepseek" {
			err := apiClient.SendRequest(query, config.Model, config.SystemPrompt)
			if err != nil {
				fmt.Printf("错误：%v\n", err)
			}
		} else {
			fmt.Println("错误：不支持的提供者。请在配置文件中设置有效的提供者")
		}
	}

	// 执行命令行解析
	if err := cmdHandler.ExecuteCmd(queryHandler); err != nil {
		fmt.Printf("错误：%v\n", err)
		os.Exit(1)
	}
}
