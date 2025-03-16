# LLM Forwarder Framework

*[中文](#chinese) | [English](#english)*

*Note: All code in this project is AI-generated using Claude 3.5 Sonnet*  
*注意：本项目所有代码均由 Claude 3.5 Sonnet 人工智能生成*

<a name="english"></a>
## English

A framework for forwarding, managing and monitoring Large Language Model (LLM) services. It provides a unified interface for handling model requests, collecting metrics, and exposing API endpoints to access these functionalities.

### Features

- Model request handling and forwarding
- Request history and storage
- Performance metrics collection
- Server health monitoring
- RESTful API endpoints
- Real-time streaming response
- Detailed request statistics
- Web interface monitoring

### Configuration

Example `config.yaml` file:

```yaml
server:
  host: "0.0.0.0"  # Server listen address
  port: 8080       # Server listen port

ollama:
  url: "http://localhost:11434"  # Ollama service address

storage:
  type: "file"     # Storage type (currently supports file storage)
  path: "./data"   # Storage path
```

### Project Structure

```
llm-fw/
├── config/       # Configuration related code
├── handlers/     # Request handlers
├── metrics/      # Metrics collector
├── proxy/        # Proxy related code
├── routes/       # Route definitions
├── storage/      # Storage implementation
├── templates/    # Web interface templates
├── config.yaml   # Configuration file
└── main.go      # Main program entry
```

### Installation and Running

1. Ensure Go 1.16 or higher is installed
2. Clone the project and enter the project directory
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run the server:
   ```bash
   go run main.go
   ```

### API Endpoints

- `GET /` - Web monitoring interface
- `GET /api/models` - Get list of available models
- `POST /api/generate` - Generate text
- `POST /api/chat` - Chat interface
- `GET /api/models/:model/stats` - Get model statistics
- `GET /api/health` - Health check endpoint

### Web Interface Features

- Model list display
- Real-time text generation
- Request history
- Model statistics
- Copy response text
- Streaming response display

### Contributing

Issues and Pull Requests are welcome!

### License

MIT License

---

<a name="chinese"></a>
## 中文

一个用于转发、管理和监控大型语言模型（LLM）服务的框架。它提供了统一的接口来处理模型请求、收集指标，并提供 API 端点来访问这些功能。

### 特点

- 模型请求处理和转发
- 请求历史记录和存储
- 性能指标收集
- 服务器健康监控
- RESTful API 端点
- 实时流式响应支持
- 详细的请求统计
- Web 界面监控

### 配置

示例 `config.yaml` 文件：

```yaml
server:
  host: "0.0.0.0"  # 服务器监听地址
  port: 8080       # 服务器监听端口

ollama:
  url: "http://localhost:11434"  # Ollama 服务地址

storage:
  type: "file"     # 存储类型（当前支持文件存储）
  path: "./data"   # 存储路径
```

### 项目结构

```
llm-fw/
├── config/       # 配置相关代码
├── handlers/     # 请求处理器
├── metrics/      # 指标收集器
├── proxy/        # 代理相关代码
├── routes/       # 路由定义
├── storage/      # 存储实现
├── templates/    # Web 界面模板
├── config.yaml   # 配置文件
└── main.go      # 主程序入口
```

### 安装和运行

1. 确保已安装 Go 1.16 或更高版本
2. 克隆项目并进入项目目录
3. 安装依赖：
   ```bash
   go mod download
   ```
4. 运行服务器：
   ```bash
   go run main.go
   ```

### API 端点

- `GET /` - Web 监控界面
- `GET /api/models` - 获取可用模型列表
- `POST /api/generate` - 生成文本
- `POST /api/chat` - 聊天接口
- `GET /api/models/:model/stats` - 获取模型统计信息
- `GET /api/health` - 健康检查端点

### Web 界面功能

- 模型列表显示
- 实时文本生成
- 请求历史记录
- 模型统计信息
- 复制响应文本
- 流式响应显示

### 贡献

欢迎提出问题和提交 Pull Request！

### 许可证

MIT License 