
# ChatGPT Agent CLI 工具

这是一个命令行界面 (CLI) 工具，用于与 ChatGPT 或 Grok API 交互，发送查询并获取响应。工具支持多种输入方式，包括命令行参数、管道输入和标志选项。以下是详细的文档，包括部署、使用的指南和开发建议。

## 1. 本地部署与构建

本节指导您在本地环境中部署和构建该工具。您需要具备基本的Go开发环境，因为该程序使用Go语言编写。

### 安装依赖
1. **安装Go环境**：确保您的系统已安装Go 1.16 或更高版本。您可以通过以下命令检查：
   ```
   go version
   ```
   如果未安装，请从[Go官网](https://golang.org/dl/)下载并安装。

2. **克隆仓库**：将代码仓库克隆到本地。
   ```
   git clone https://github.com/your-username/chatgpt-agent.git  # 替换为实际仓库地址
   cd chatgpt-agent
   ```

3. **安装依赖包**：该程序依赖于外部库，如 `github.com/spf13/cobra`。使用Go模块安装：
   ```
   go mod init chatgpt-agent  # 如果还没有go.mod文件
   go mod tidy  # 下载并整理依赖
   ```

### 构建程序
1. **构建二进制文件**：使用Go构建工具编译程序。
   ```
   go build -o chat-agent main.go
   ```
   这会生成一个可执行文件 `chat-agent`（在Windows上为 `chat-agent.exe`）。

2. **运行程序**：在构建后，您可以直接运行二进制文件进行测试：
   ```
   ./chat-agent --help  # 查看帮助信息
   ```

### 注意事项
- **API密钥**：构建后，您需要在运行时提供API密钥（使用 `--key` 标志）。请确保从相关提供者（如x.ai for Grok）获取有效的密钥。
- **环境变量**：如果您不想每次都输入密钥，可以在代码中添加环境变量支持（作为扩展开发）。
- **平台兼容**：该程序在Linux、macOS和Windows上均可运行，但管道输入在某些Windows环境可能需要额外工具（如Git Bash）。

## 2. 如何使用

本节解释如何运行该工具，包括命令参数、标志选项和示例。工具允许您发送查询到ChatGPT或Grok API，并支持从命令行、管道输入获取内容。

### 基本命令格式
```
chat-agent [query] [flags]
```
- `[query]`：可选的位置参数，用于直接输入查询内容。
- `[flags]`：命令行标志，用于配置API密钥、提供者、模型和查询。

### 可用标志
- `--key` 或 `-k`：字符串，API密钥（必填）。
- `--provider` 或 `-p`：字符串，API提供者，支持 "chatgpt" 或 "grok"（默认："chatgpt"）。
- `--model` 或 `-m`：字符串，指定模型名称，例如 "grok-3-beta" 或 "gpt-3.5-turbo"（默认："grok" 如果提供者为 "grok"）。
- `--ques` 或 `-q`：字符串，指定查询内容（优先级高）。如果结合管道输入，该内容会与管道内容合并。

### 使用示例
1. **基本查询**：
   ```
   ./chat-agent "你好，世界" --provider grok --model grok-3-beta --key 您的API密钥
   ```
   - 效果：直接使用位置参数作为查询，发送到Grok API。

2. **使用 -q 标志**：
   ```
   ./chat-agent --provider grok --model grok-3-beta --key 您的API密钥 -q "请分析这个文件"
   ```
   - 效果：使用指定的查询内容。如果有管道输入，它会追加到查询末尾。

3. **从管道输入读取内容**（例如读取文件）：
   ```
   cat yourfile.txt | ./chat-agent --provider grok --model grok-3-beta --key 您的API密钥
   ```
   - 效果：程序会从标准输入（管道）读取文件内容作为查询。

4. **结合 -q 和管道输入**：
   ```
   cat yourfile.txt | ./chat-agent --provider grok --model grok-3-beta --key 您的API密钥 -q "请解释以下内容"
   ```
   - 效果：查询会合并为 "请解释以下内容\n文件内容..."，发送到API。

### 注意事项
- **查询优先级**：-q 标志 > 位置参数 > 管道输入。如果 -q 提供，管道内容会附加到它后面。
- **错误处理**：如果查询为空，程序会输出 "错误：查询不能为空"。
- **API响应**：程序会打印API的响应内容到控制台。
- **安全性**：不要在命令行中直接暴露API密钥；考虑使用环境变量或配置文件。

## 3. 如何开发

本节针对开发者，提供代码结构概述、修改建议和扩展方法。如果您想扩展该工具（如添加新提供者或功能），可以基于此进行开发。

### 代码结构
- **main.go**：入口文件，包含主函数和Cobra命令定义。
  - `var` 声明：定义了全局变量，如 `apiKey`、`provider`、`model` 和 `question`。
  - `main()` 函数：设置Cobra命令，处理输入逻辑和API调用。
  - API调用逻辑：位于 `Run` 函数中，处理Grok API请求（当前仅支持Grok，可扩展）。
- **依赖**：使用 `github.com/spf13/cobra` 进行CLI管理，`net/http` 和 `encoding/json` 处理API交互。

### 开发步骤
1. **设置开发环境**：确保安装Go和依赖（如上文所述）。使用VS Code或其他IDE打开项目。

2. **添加新功能**：
   - **支持新提供者**：在 `Run` 函数中扩展 `if provider == "grok"` 块，例如添加 `else if provider == "chatgpt"`，并修改URL和Payload。
     ```go
     if provider == "grok" {
         // 当前逻辑
     } else if provider == "chatgpt" {
         url := "https://api.openai.com/v1/chat/completions"
         // 相应修改
     }
     ```
   - **修改查询逻辑**：`Run` 函数中的查询获取部分可以进一步自定义，例如添加更多输入源。
   - **错误处理和日志**：添加更多日志输出或错误检查，使用 `log` 包。

3. **测试建议**：
   - **单元测试**：编写测试函数，例如测试API调用逻辑（使用Go的 `testing` 包）。
   - **运行测试**：使用 `go test ./...`。
   - **调试**：在命令行运行程序时，使用 `--help` 查看标志，或添加print语句调试输入。

4. **最佳实践**：
   - **代码风格**：遵循Go的规范，使用 `go fmt` 格式化代码。
   - **版本控制**：使用Git管理代码变更。
   - **安全**：避免硬编码敏感信息，如API密钥。
   - **扩展**：如果需要更多功能（如支持文件上传），可以添加新命令或库（如 `os` 处理文件）。

如果您对代码有任何修改或扩展需求，请参考Go官方文档或Cobra文档。欢迎通过Pull Request贡献代码！