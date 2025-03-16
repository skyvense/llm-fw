# LLM Framework

A framework for managing and monitoring Large Language Model (LLM) services. It provides a unified interface for handling model requests, collecting metrics, and exposing API endpoints to access these functionalities.

## Features

- Model request handling and proxying
- Request history and storage
- Performance metrics collection
- Server health monitoring
- RESTful API endpoints

## Configuration

Example `config.yaml` file:

```yaml
server:
  host: "0.0.0.0"  # Server listen address
  port: 8080       # Server listen port

storage:
  type: "file"     # Storage type (currently supports file storage)
  path: "./data"   # Storage path
```

## Project Structure

```
llm-fw/
├── config/       # Configuration related code
├── handlers/     # Request handlers
├── metrics/      # Metrics collector
├── proxy/        # Proxy related code
├── routes/       # Route definitions
├── storage/      # Storage implementation
├── config.yaml   # Configuration file
└── main.go      # Main program entry
```

## Installation and Running

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

## API Endpoints

- `GET /api/models` - Get list of available models
- `POST /api/generate` - Generate text
- `GET /api/stats` - Get model statistics
- `GET /api/health` - Health check endpoint

## Contributing

Issues and Pull Requests are welcome!

## License

MIT License 