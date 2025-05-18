# LLM Agent CLI

一个支持 LLM API 调用的 CLI 工具，提供流式响应和配置管理功能。

## 功能特性

- ✅ 支持 LLM API 流式响应
- ⚙️ 配置文件自动管理
- 📝 支持多种输入方式：
  - 命令行参数
  - 管道输入
  - 文件内容读取
- 🔒 安全的配置文件权限 (0600)

## 安装依赖

1. 确保已安装 Go 1.16+：
```bash
go version
```

2. 安装 Cobra CLI 库：
```bash
go get -u github.com/spf13/cobra@latest
```

## 构建项目

```bash
# 克隆项目
git clone https://github.com/GoldBro233/GPTAgentCLI-Go.git
cd GPTAgentCLI-Go

# 构建可执行文件
go build -o chatgpt-agent
```

## 使用方法

### 首次配置
首次运行会在同目录自动创建配置文件：
```bash
./chatgpt-agent
```

### 查询方式
1. 命令行参数：
```bash
./chatgpt-agent "你的问题"
```

2. 使用 `-q` 标志：
```bash
./chatgpt-agent -q "你的问题"
```

3. 管道输入：
```bash
echo "你的问题" | ./chatgpt-agent
cat file.txt | ./chatgpt-agent
```

## 配置文件
配置文件位于可执行文件同目录下的 `config.json`，包含以下字段：
```json
{
  "provider": "deepseek",
  "model": "deepseek-chat",
  "apiKey": "your_api_key",
  "systemPrompt": "自定义系统提示"
}
```

## 支持的提供商
- DeepSeek

## 注意事项
1. 首次配置完成后程序会自动退出，需重新运行
2. 确保 API 密钥保密
3. 流式响应会实时显示内容并统计 token 用量

## 依赖
- Go 1.16+
- Cobra CLI 库

