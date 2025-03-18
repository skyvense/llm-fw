# LLM Framework

A lightweight framework for building LLM applications with Ollama integration.

一个轻量级的 LLM 应用框架，集成了 Ollama.

## Features 特性

- 🚀 Easy integration with Ollama
- 📊 Real-time metrics monitoring
- 💾 Request history storage
- 🌐 Web UI interface
- 🔄 Streaming response support
- 📝 Configurable through YAML

- 🚀 轻松集成 Ollama
- 📊 实时指标监控
- 💾 请求历史存储
- 🌐 Web 界面
- 🔄 流式响应支持
- 📝 通过 YAML 配置

## Prerequisites 前置条件

- Go 1.20 or later
- Ollama installed and running
- Git

## Installation 安装

1. Clone the repository 克隆仓库:
```bash
git clone https://github.com/yourusername/llm-fw.git
cd llm-fw
```

2. Install dependencies 安装依赖:
```bash
go mod download
```

3. Configure the application 配置应用:
Create a `config.yaml` file in the root directory 在根目录创建 `config.yaml` 文件:
```yaml
server:
  host: "localhost"
  port: 8080

ollama:
  url: "http://localhost:11434"

storage:
  type: "file"
  path: "data"
```

## Usage 使用方法

1. Start the server 启动服务器:
```bash
go run main.go
```

2. Open your browser and navigate to 打开浏览器访问:
```
http://localhost:8080
```

## Configuration 配置说明

The application can be configured through `config.yaml` in the root directory:

应用可以通过根目录下的 `config.yaml` 进行配置：

```yaml
server:
  host: "localhost"    # Server host 服务器主机
  port: 8080          # Server port 服务器端口

ollama:
  url: "http://localhost:11434"  # Ollama server URL Ollama 服务器地址

storage:
  type: "file"        # Storage type (file/memory) 存储类型（文件/内存）
  path: "data"        # Storage path for file storage 文件存储路径
```

## API Endpoints API 接口

- `GET /api/models` - List available models 列出可用模型
- `POST /api/generate` - Generate text 生成文本
- `GET /api/history` - Get request history 获取请求历史
- `GET /api/metrics` - Get metrics 获取指标

## Development 开发

### Project Structure 项目结构

```
llm-fw/
├── api/          # API types and interfaces API 类型和接口
├── config/       # Configuration handling 配置处理
├── handlers/     # Request handlers 请求处理器
├── metrics/      # Metrics collection 指标收集
├── routes/       # Route setup 路由设置
├── storage/      # Storage implementations 存储实现
├── templates/    # HTML templates HTML 模板
├── config.yaml   # Application configuration 应用配置
└── main.go       # Application entry point 应用入口
```

### Adding New Features 添加新功能

1. Create new types in `api` package 在 `api` 包中创建新类型
2. Implement handlers in `handlers` package 在 `handlers` 包中实现处理器
3. Add routes in `routes` package 在 `routes` 包中添加路由
4. Update templates if needed 如果需要，更新模板

## Contributing 贡献

Contributions are welcome! Please feel free to submit a Pull Request.

欢迎贡献！请随时提交 Pull Request。

## License 许可证

This project is licensed under the MIT License - see the LICENSE file for details.

本项目采用 MIT 许可证 - 详见 LICENSE 文件。 