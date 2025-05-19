package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// CmdHandler 结构体用于封装命令行处理逻辑
type CmdHandler struct {
	Question string // 用于指定查询内容
}

// NewCmdHandler 创建并返回一个新的 CmdHandler 实例
func NewCmdHandler() *CmdHandler {
	return &CmdHandler{}
}

// isTerminal 检查文件是否是终端
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// BuildRootCmd 创建并返回 cobra 命令
func (ch *CmdHandler) BuildRootCmd(queryHandler func(string)) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "chatgpt-agent [query]",
		Short: "ChatGPT Agent CLI Tool",
		Long: `A command-line tool to interact with ChatGPT or other AI models like DeepSeek.
This tool allows users to send queries directly, via pipe, or interactively, with support for file content as context.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			query, err := ch.parseInput(args)
			if err != nil {
				return err
			}
			if query == "" {
				return fmt.Errorf("query cannot be empty")
			}
			queryHandler(query) // 将处理后的查询传递给回调函数
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// 绑定 --ques 或 -q 标志
	rootCmd.Flags().StringVarP(&ch.Question, "ques", "q", "", "Specify the query content")

	// 设置自定义帮助信息
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println("ChatGPT Agent CLI - A tool for interacting with AI models like ChatGPT or DeepSeek.")
		fmt.Println("\nUsage:")
		fmt.Printf("  %s [flags] [query]\n", cmd.Use)
		fmt.Println("\nFlags:")
		printFlagWithPadding("-q, --ques", "Specify the query content directly")
		printFlagWithPadding("-h, --help", "Display this help information")
		fmt.Println("\nExamples:")
		fmt.Println("  Direct query:")
		fmt.Println("    chatgpt-agent 'How are you?'")
		fmt.Println("  Using flag:")
		fmt.Println("    chatgpt-agent --ques 'Explain Go programming'")
		fmt.Println("  Pipe a file name to read its content:")
		fmt.Println("    echo 'config.json' | chatgpt-agent -q 'Analyze this file'")
		fmt.Println("\nConfiguration:")
		fmt.Println("  Configuration is loaded from config.json in the executable directory.")
		fmt.Println("  Run the tool for the first time to set up API keys and preferences.")
	})

	return rootCmd
}

// parseInput 解析输入内容，包括命令行参数、标志和管道输入
func (ch *CmdHandler) parseInput(args []string) (string, error) {
	var query string
	var pipeContent string

	// 第一优先：使用 --ques 或 -q 标志
	if ch.Question != "" {
		query = ch.Question
		// 仅在标准输入不是终端时尝试读取管道输入
		if !isTerminal(os.Stdin) {
			var err error
			pipeContent, err = readInputFromScanner(os.Stdin)
			if err != nil {
				return "", fmt.Errorf("failed to read pipe input: %v", err)
			}
			if pipeContent != "" {
				// 尝试读取文件内容（假设管道输入可能是文件名）
				if isSingleLine(pipeContent) {
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
		input, err := readInputFromScanner(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read standard input: %v", err)
		}
		query = input
	}

	return query, nil
}

// readInputFromScanner 从指定的输入源读取内容
func readInputFromScanner(input *os.File) (string, error) {
	scanner := bufio.NewScanner(input)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if len(lines) > 0 {
		return strings.Join(lines, "\n"), nil
	}
	return "", nil
}

// isSingleLine 检查输入是否为单行
func isSingleLine(input string) bool {
	return !strings.Contains(input, "\n")
}

// printFlagWithPadding 格式化输出标志和描述
func printFlagWithPadding(name, description string) {
	padding := 20
	fmt.Printf("  %-*s %s\n", padding, name, description)
}

// ExecuteCmd 执行命令行解析
func (ch *CmdHandler) ExecuteCmd(queryHandler func(string)) error {
	rootCmd := ch.BuildRootCmd(queryHandler)
	return rootCmd.Execute()
}