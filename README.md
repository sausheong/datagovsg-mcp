# datagovsg-mcp

An MCP (Model Context Protocol) server that gives AI assistants access to Singapore Government public data via [data.gov.sg](https://data.gov.sg).

## Tools

### Dataset tools

| Tool | Description |
|------|-------------|
| `search_datasets` | Search for datasets by keyword. Returns dataset IDs, names, descriptions, formats, and coverage dates. |
| `get_dataset_metadata` | Get the full schema for a dataset — column names, data types, units, and categorical flags. |
| `list_collections` | Browse thematic collections of related datasets grouped by agency or topic. |
| `get_collection_info` | Get details for a specific collection including its child dataset IDs. |
| `query_dataset` | Fetch actual records from a dataset with optional column filters and pagination. |

### Realtime tools

| Tool | Description |
|------|-------------|
| `get_environment_reading` | Live NEA sensor readings: `air-temperature`, `rainfall`, `relative-humidity`, `wind-speed`, `wind-direction`, `uv-index`, `psi`, `pm25` |
| `get_weather_forecast` | Weather forecasts: `2-hour` (area nowcast), `24-hour` (regional), `4-day` (island-wide outlook) |
| `get_transport_info` | Live transport data: `taxi` (GPS locations), `carpark` (HDB lot availability), `traffic-cameras` (image URLs and locations) |

## Usage workflow

A typical dataset query follows three steps:

1. **Find a dataset** — `search_datasets` with a keyword (e.g. `"unemployment"`)
2. **Inspect its schema** — `get_dataset_metadata` with the returned dataset ID
3. **Query records** — `query_dataset` with column filters from the schema

## Installation

### Build from source

Requires Go 1.21+.

```bash
git clone https://github.com/sausheong/datagovsg-mcp.git
cd datagovsg-mcp
make build
```

The binary is output to `./datagovsg-mcp`.

## Running the server

### STDIO mode (default)

Used by MCP clients that launch the server as a subprocess (Claude Desktop, Claude Code, etc.):

```bash
./datagovsg-mcp
```

### HTTP mode

Starts a streamable HTTP server for remote or multi-client use:

```bash
./datagovsg-mcp --http :8080
```

The MCP endpoint is available at `http://localhost:8080/mcp`.

## MCP client configuration

### Claude Code

Add to `.claude/mcp.json` in your project:

```json
{
  "mcpServers": {
    "datagovsg": {
      "command": "/path/to/datagovsg-mcp"
    }
  }
}
```

### Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "datagovsg": {
      "command": "/path/to/datagovsg-mcp"
    }
  }
}
```

### HTTP transport

For clients that support streamable HTTP:

```json
{
  "mcpServers": {
    "datagovsg": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

## Development

```bash
make test      # run tests
make run       # build and run in STDIO mode
make run-http  # build and run HTTP server on :8080
make clean     # remove binary
```

## Data sources

- **Datasets API** — [data.gov.sg](https://data.gov.sg) (Singapore Government open data portal)
- **Realtime APIs** — [api.data.gov.sg](https://api.data.gov.sg) (NEA environment, weather forecasts, transport)

No API key is required. All data is publicly accessible.

## License

MIT
