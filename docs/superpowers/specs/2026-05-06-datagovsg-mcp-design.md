# data.gov.sg MCP Server — Design Spec

**Date:** 2026-05-06  
**Status:** Approved

## Overview

A Go MCP server that gives Claude (and any MCP-compatible client) access to Singapore's public data via [data.gov.sg](https://data.gov.sg). Covers ~4,500 historical datasets and 14 real-time feeds across environment, weather, and transport.

---

## Architecture

### Runtime
- Single Go binary, STDIO transport (JSON-RPC over stdin/stdout)
- No port, no config file, no database
- MCP library: `github.com/mark3labs/mcp-go`

### API Backends

| Client | Base URL | Purpose |
|---|---|---|
| `DatasetClient` | `https://api-production.data.gov.sg/v2/public/api/` | Dataset search, metadata, collections |
| `DatasetClient` | `https://data.gov.sg/api/action/datastore_search` | Record queries (CKAN) |
| `RealtimeClient` | `https://api.data.gov.sg/v1/` | Environment readings, forecasts, transport |

No authentication required. All APIs are public.

### File Layout

```
datagovsg_mcp/
├── main.go                  # wires server, registers all tools
├── go.mod
├── go.sum
└── internal/
    ├── client/
    │   ├── dataset.go       # DatasetClient — v2 API + CKAN
    │   └── realtime.go      # RealtimeClient — environment + transport
    └── tools/
        ├── datasets.go      # search_datasets, get_dataset_metadata,
        │                    #   list_collections, get_collection_info
        ├── query.go         # query_dataset
        └── realtime.go      # get_environment_reading, get_weather_forecast,
                             #   get_transport_info
```

### HTTP Client Config
- Timeout: 15 seconds
- User-Agent: `datagovsg-mcp/1.0`
- Retry: 2 attempts with exponential backoff on 429/503

---

## Tools

### 1. `search_datasets`
Find datasets by keyword.

**Parameters:**
- `query` (string, required) — keyword to search
- `page` (int, optional, default 1)
- `limit` (int, optional, default 10, max 100)

**Returns:** Array of `{ datasetId, name, description, format, managedByAgencyName, coverageStart, coverageEnd, lastUpdatedAt }`

**Upstream:** `GET /v2/public/api/datasets?query=&page=&resultPerPage=`

---

### 2. `get_dataset_metadata`
Get the full schema for a dataset — column names, types, units, and categorical flags.

**Parameters:**
- `dataset_id` (string, required)

**Returns:** `{ datasetId, name, description, format, managedBy, coverageStart, coverageEnd, columnMetadata: { order, map, metaMapping } }`

**Upstream:** `GET /v2/public/api/datasets/{id}/metadata`

---

### 3. `list_collections`
Browse thematic collections of datasets.

**Parameters:**
- `page` (int, optional, default 1)
- `limit` (int, optional, default 10, max 100)

**Returns:** Array of `{ collectionId, name, description, managedByAgencyName, frequency, childDatasets[] }`

**Upstream:** `GET /v2/public/api/collections?page=&resultPerPage=`

---

### 4. `get_collection_info`
Get full detail for a specific collection, including its child dataset IDs.

**Parameters:**
- `collection_id` (string, required)

**Returns:** `{ collectionId, name, description, managedByAgencyName, frequency, sources[], childDatasets[] }`

**Upstream:** `GET /v2/public/api/collections/{id}/metadata`

---

### 5. `query_dataset`
Pull records from a dataset with optional column-level filters.

**Parameters:**
- `dataset_id` (string, required)
- `filters` (object, optional) — key/value pairs where key is column name, value is exact match string. Example: `{ "residential_status": "overall" }`
- `limit` (int, optional, default 20, max 100)
- `offset` (int, optional, default 0)

**Returns:** `{ total, fields[], records[] }`

**Note:** Full-text search (`q` param) is not supported by all datasets. Use `filters` for column-specific matching.

**Upstream:** `GET data.gov.sg/api/action/datastore_search?resource_id=&limit=&offset=&filters=`

---

### 6. `get_environment_reading`
Get current sensor readings from NEA stations across Singapore.

**Parameters:**
- `type` (string, required) — one of:
  - `air-temperature` — °C per station
  - `rainfall` — mm per station
  - `relative-humidity` — % per station
  - `wind-speed` — km/h per station
  - `wind-direction` — degrees per station
  - `uv-index` — UV index (single reading)
  - `psi` — Pollutant Standards Index by region
  - `pm25` — PM2.5 by region

**Returns:** `{ timestamp, reading_unit, stations[], readings[] }` — each reading has `station_id` and `value`; stations carry name and GPS coordinates.

**Upstream:** `GET api.data.gov.sg/v1/environment/{type}`

---

### 7. `get_weather_forecast`
Get weather forecasts for Singapore.

**Parameters:**
- `type` (string, required) — one of:
  - `2-hour` — area-level nowcast (next 2 hours)
  - `24-hour` — region-level forecast (today)
  - `4-day` — island-wide 4-day outlook

**Returns:** Forecast periods, area/region breakdowns, weather condition descriptions (e.g., "Partly Cloudy", "Thundery Showers"), temperature ranges where applicable.

**Upstream:** `GET api.data.gov.sg/v1/environment/{type}-weather-forecast`

---

### 8. `get_transport_info`
Get live transport data.

**Parameters:**
- `type` (string, required) — one of:
  - `taxi` — current taxi GPS locations across Singapore
  - `carpark` — HDB carpark lot availability (2,001 carparks, lots available vs total)
  - `traffic-cameras` — live traffic camera image URLs + GPS location per camera

**Returns:**
- `taxi`: `{ timestamp, coordinates[] }` — array of `[longitude, latitude]`
- `carpark`: `{ timestamp, carparks[] }` — each with `carpark_number`, `lot_type`, `lots_available`, `total_lots`
- `traffic-cameras`: `{ timestamp, cameras[] }` — each with `camera_id`, `location`, `image_url`, `image_metadata`

**Upstream:** `GET api.data.gov.sg/v1/transport/{taxi-availability|carpark-availability|traffic-images}`

---

## Data Flow

```
Claude (MCP client)
    │  STDIO (JSON-RPC)
    ▼
main.go  →  mcp-go server
    │
    ├── tools/datasets.go  ──►  DatasetClient  ──►  api-production.data.gov.sg/v2/…
    │
    ├── tools/query.go     ──►  DatasetClient  ──►  data.gov.sg/api/action/datastore_search
    │
    └── tools/realtime.go  ──►  RealtimeClient ──►  api.data.gov.sg/v1/…
```

---

## Error Handling

- All API errors are returned as MCP tool errors with the upstream message included
- The LLM receives readable error text and can explain or retry
- No panics — all paths return `(result, error)`
- Invalid `type` values (e.g. unknown environment type) return a validation error before making any HTTP call

---

## Out of Scope

- Authentication / API keys (all APIs are public)
- Response caching
- Non-CSV/non-queryable dataset formats (GeoJSON, XLSX downloads)
- Blocked endpoints: `traffic-speed-bands`, `traffic-incidents`, `vms`, `estimated-travel-times`, `nowcasting` (all return 403)
