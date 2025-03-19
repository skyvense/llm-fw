# LLM Forwarder Framework

LLM Forwarder Framework 是一个用于转发和管理 LLM（大型语言模型）请求的框架，支持多种模型和存储方式。

## 功能特点

- 支持 Ollama API 的模型调用
- 提供模型使用统计和历史记录
- 支持多种存储方式（SQLite/文件存储）
- 实时请求监控和统计
- 美观的 Web 界面
- 支持中英文切换

## 配置说明

### 基础配置

在项目根目录创建 `config.yaml` 文件：

```yaml
server:
  host: "localhost"  # 服务器监听地址
  port: 8080        # 服务器监听端口

ollama:
  url: "http://localhost:11434"  # Ollama 服务器地址

storage:
  type: "sqlite"    # 存储类型：sqlite 或 file
  path: "data/llm-fw.db"  # SQLite 数据库文件路径
```

### 存储配置

框架支持两种存储方式，推荐使用 SQLite：

1. SQLite 存储（推荐）：
   ```yaml
   storage:
     type: "sqlite"
     path: "data/llm-fw.db"  # SQLite 数据库文件路径
   ```

2. 文件存储：
   ```yaml
   storage:
     type: "file"
     path: "data"  # 数据文件存储目录
   ```

### 环境变量

也可以通过环境变量覆盖配置文件中的设置：

- `SERVER_HOST`: 服务器监听地址
- `SERVER_PORT`: 服务器监听端口
- `OLLAMA_URL`: Ollama 服务器地址
- `STORAGE_TYPE`: 存储类型（sqlite/file）
- `STORAGE_PATH`: 存储路径

## 使用方法

1. 启动服务：
   ```bash
   go run .
   ```

2. 访问 Web 界面：
   打开浏览器访问 `http://localhost:8080`

3. API 接口：
   - 聊天接口：`POST /api/chat`
   - 生成接口：`POST /api/generate`
   - 模型列表：`GET /api/models`
   - 历史记录：`GET /api/history`

## API 示例

### 聊天接口

```bash
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2",
    "messages": [
      {
        "role": "user",
        "content": "你好"
      }
    ]
  }'
```

### 获取模型列表

```bash
# 获取所有模型
curl http://localhost:8080/api/models

# 只获取有统计数据的模型
curl http://localhost:8080/api/models?stats_only=true
```

## 开发说明

项目使用 Go 1.21+ 开发，主要依赖：

- gin: Web 框架
- modernc.org/sqlite: SQLite 数据库驱动（纯 Go 实现，无 CGO 依赖）
- yaml.v3: YAML 配置文件解析

## 注意事项

1. 使用 SQLite 存储时，数据库文件会自动创建
2. 确保配置文件中的路径具有正确的读写权限
3. 建议在生产环境中使用 SQLite 存储，以获得更好的并发性能和数据一致性

## Features 特性

- 🔄 Proxy requests to Ollama server
- 📊 Real-time metrics monitoring
- 💾 Request history storage
- 🌐 Web UI interface
- ⚡ Streaming response support
- 📝 Configurable through YAML

- 🔄 代理转发 Ollama 服务器请求
- 📊 实时指标监控
- 💾 请求历史存储
- 🌐 Web 界面
- ⚡ 流式响应支持
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
  url: "http://localhost:11434"  # Your Ollama server address 你的 Ollama 服务器地址

storage:
  type: "file"
  path: "data"
```

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