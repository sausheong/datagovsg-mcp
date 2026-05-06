# data.gov.sg MCP Server Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go MCP server with 8 tools covering data.gov.sg historical datasets and real-time feeds, served over STDIO.

**Architecture:** Two HTTP clients (`DatasetClient` for v2+CKAN APIs, `RealtimeClient` for v1 real-time APIs) wrapped by three tool registration functions, wired together in `main.go` using `mark3labs/mcp-go`.

**Tech Stack:** Go 1.22+, `github.com/mark3labs/mcp-go`, stdlib `net/http`, `encoding/json`, `net/http/httptest` for tests.

---

## File Map

| File | Responsibility |
|---|---|
| `main.go` | Create MCP server, instantiate clients, register all tools, serve STDIO |
| `internal/client/dataset.go` | `DatasetClient` — v2 API (search, metadata, collections) + CKAN (query) |
| `internal/client/dataset_test.go` | Tests for all `DatasetClient` methods |
| `internal/client/realtime.go` | `RealtimeClient` — environment, forecasts, transport |
| `internal/client/realtime_test.go` | Tests for all `RealtimeClient` methods |
| `internal/tools/datasets.go` | MCP tool handlers: `search_datasets`, `get_dataset_metadata`, `list_collections`, `get_collection_info` |
| `internal/tools/datasets_test.go` | Tests for dataset tool handlers |
| `internal/tools/query.go` | MCP tool handler: `query_dataset` |
| `internal/tools/query_test.go` | Tests for query tool handler |
| `internal/tools/realtime.go` | MCP tool handlers: `get_environment_reading`, `get_weather_forecast`, `get_transport_info` |
| `internal/tools/realtime_test.go` | Tests for realtime tool handlers |

---

## Task 1: Project scaffolding

**Files:**
- Create: `go.mod`
- Create: `internal/client/.gitkeep`
- Create: `internal/tools/.gitkeep`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/sausheong/projects/datagovsg_mcp
go mod init github.com/sausheong/datagovsg-mcp
mkdir -p internal/client internal/tools
```

- [ ] **Step 2: Add mcp-go dependency**

```bash
go get github.com/mark3labs/mcp-go@latest
```

- [ ] **Step 3: Verify go.mod looks correct**

```bash
cat go.mod
```

Expected: module line is `module github.com/sausheong/datagovsg-mcp`, go version `1.22` or higher, and `require` block contains `github.com/mark3labs/mcp-go`.

- [ ] **Step 4: Commit scaffolding**

```bash
git add go.mod go.sum
git commit -m "feat: initialize Go module with mcp-go dependency"
```

---

## Task 2: DatasetClient — types and struct

**Files:**
- Create: `internal/client/dataset.go`

- [ ] **Step 1: Write `internal/client/dataset.go` with all types and the client constructor**

```go
package client

import (
	"fmt"
	"net/http"
	"time"
)

const (
	v2BaseURL  = "https://api-production.data.gov.sg/v2/public/api"
	ckanURL    = "https://data.gov.sg/api/action/datastore_search"
	userAgent  = "datagovsg-mcp/1.0"
)

type DatasetClient struct {
	http    *http.Client
	v2Base  string
	ckanURL string
}

func NewDatasetClient() *DatasetClient {
	return &DatasetClient{
		http:    &http.Client{Timeout: 15 * time.Second},
		v2Base:  v2BaseURL,
		ckanURL: ckanURL,
	}
}

// Dataset list types

type Dataset struct {
	DatasetId           string `json:"datasetId"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	Format              string `json:"format"`
	Status              string `json:"status"`
	ManagedByAgencyName string `json:"managedByAgencyName"`
	CoverageStart       string `json:"coverageStart"`
	CoverageEnd         string `json:"coverageEnd"`
	LastUpdatedAt       string `json:"lastUpdatedAt"`
}

type DatasetsResult struct {
	Datasets      []Dataset `json:"datasets"`
	Pages         int       `json:"pages"`
	TotalRowCount int       `json:"totalRowCount"`
}

type datasetsResponse struct {
	Code     int            `json:"code"`
	Data     DatasetsResult `json:"data"`
	ErrorMsg string         `json:"errorMsg"`
}

// Dataset metadata types

type ColumnMeta struct {
	Name          string `json:"name"`
	ColumnTitle   string `json:"columnTitle"`
	DataType      string `json:"dataType"`
	UnitOfMeasure string `json:"unitOfMeasure,omitempty"`
	IsCategorical bool   `json:"isCategorical"`
}

type ColumnMetadata struct {
	Order       []string              `json:"order"`
	Map         map[string]string     `json:"map"`
	MetaMapping map[string]ColumnMeta `json:"metaMapping"`
}

type DatasetMetadata struct {
	DatasetId      string         `json:"datasetId"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Format         string         `json:"format"`
	ManagedBy      string         `json:"managedBy"`
	CoverageStart  string         `json:"coverageStart"`
	CoverageEnd    string         `json:"coverageEnd"`
	ColumnMetadata ColumnMetadata `json:"columnMetadata"`
}

type metadataResponse struct {
	Code     int             `json:"code"`
	Data     DatasetMetadata `json:"data"`
	ErrorMsg string          `json:"errorMsg"`
}

// Collections types

type Collection struct {
	CollectionId        string   `json:"collectionId"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	ManagedByAgencyName string   `json:"managedByAgencyName"`
	ManagedBy           string   `json:"managedBy"`
	Frequency           string   `json:"frequency"`
	Sources             []string `json:"sources"`
	ChildDatasets       []string `json:"childDatasets"`
}

type CollectionsResult struct {
	Collections []Collection `json:"collections"`
}

type collectionsResponse struct {
	Code     int               `json:"code"`
	Data     CollectionsResult `json:"data"`
	ErrorMsg string            `json:"errorMsg"`
}

type collectionMetaResponse struct {
	Code int `json:"code"`
	Data struct {
		CollectionMetadata Collection `json:"collectionMetadata"`
	} `json:"data"`
	ErrorMsg string `json:"errorMsg"`
}

// CKAN query types

type CKANField struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type CKANResult struct {
	Fields  []CKANField              `json:"fields"`
	Records []map[string]interface{} `json:"records"`
	Total   int                      `json:"total"`
}

type ckanResponse struct {
	Success bool       `json:"success"`
	Result  CKANResult `json:"result"`
	Error   *struct {
		Type    string   `json:"__type"`
		Message []string `json:"message"`
	} `json:"error"`
}

func apiError(code int, msg string) error {
	return fmt.Errorf("upstream error %d: %s", code, msg)
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/client/...
```

Expected: no output (clean build).

- [ ] **Step 3: Commit**

```bash
git add internal/client/dataset.go
git commit -m "feat: add DatasetClient types and struct"
```

---

## Task 3: DatasetClient — HTTP helper + SearchDatasets

**Files:**
- Modify: `internal/client/dataset.go`
- Create: `internal/client/dataset_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/client/dataset_test.go`:

```go
package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestDatasetClient(serverURL string) *DatasetClient {
	return &DatasetClient{
		http:    &http.Client{Timeout: 5 * time.Second},
		v2Base:  serverURL + "/v2/public/api",
		ckanURL: serverURL + "/api/action/datastore_search",
	}
}

func TestSearchDatasets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/public/api/datasets" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("User-Agent") != userAgent {
			t.Errorf("expected User-Agent %q, got %q", userAgent, r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(datasetsResponse{
			Code: 0,
			Data: DatasetsResult{
				Datasets:      []Dataset{{DatasetId: "d_abc", Name: "Test Dataset", Format: "CSV"}},
				Pages:         1,
				TotalRowCount: 1,
			},
		})
	}))
	defer srv.Close()

	c := newTestDatasetClient(srv.URL)
	result, err := c.SearchDatasets("test", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Datasets) != 1 {
		t.Fatalf("expected 1 dataset, got %d", len(result.Datasets))
	}
	if result.Datasets[0].DatasetId != "d_abc" {
		t.Errorf("expected d_abc, got %s", result.Datasets[0].DatasetId)
	}
}

func TestSearchDatasets_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestDatasetClient(srv.URL)
	_, err := c.SearchDatasets("test", 1, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/client/... -run TestSearchDatasets -v
```

Expected: FAIL — `c.SearchDatasets undefined`

- [ ] **Step 3: Add doGet helper and SearchDatasets to `internal/client/dataset.go`**

First update the import block at the top of `internal/client/dataset.go` to:

```go
import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)
```

Then append to the end of the file:

```go
func (c *DatasetClient) doGet(url string) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
		req, reqErr := http.NewRequest(http.MethodGet, url, nil)
		if reqErr != nil {
			return nil, reqErr
		}
		req.Header.Set("User-Agent", userAgent)
		resp, err = c.http.Do(req)
		if err != nil {
			continue
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			resp.Body.Close()
			continue
		}
		return resp, nil
	}
	return resp, err
}

func (c *DatasetClient) SearchDatasets(query string, page, limit int) (*DatasetsResult, error) {
	url := fmt.Sprintf("%s/datasets?query=%s&page=%d&resultPerPage=%d", c.v2Base, query, page, limit)
	resp, err := c.doGet(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp.StatusCode, "search_datasets failed")
	}

	var body datasetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if body.ErrorMsg != "" {
		return nil, fmt.Errorf("api error: %s", body.ErrorMsg)
	}
	return &body.Data, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/client/... -run TestSearchDatasets -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/client/dataset.go internal/client/dataset_test.go
git commit -m "feat: add DatasetClient.SearchDatasets with retry"
```

---

## Task 4: DatasetClient — GetDatasetMetadata

**Files:**
- Modify: `internal/client/dataset.go`
- Modify: `internal/client/dataset_test.go`

- [ ] **Step 1: Write the failing test**

Append to `internal/client/dataset_test.go`:

```go
func TestGetDatasetMetadata(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/public/api/datasets/d_abc/metadata" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metadataResponse{
			Code: 0,
			Data: DatasetMetadata{
				DatasetId: "d_abc",
				Name:      "Test Dataset",
				ColumnMetadata: ColumnMetadata{
					Order: []string{"c_1"},
					Map:   map[string]string{"c_1": "period"},
					MetaMapping: map[string]ColumnMeta{
						"c_1": {Name: "period", ColumnTitle: "Period", DataType: "Text"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	c := newTestDatasetClient(srv.URL)
	meta, err := c.GetDatasetMetadata("d_abc")
	if err != nil {
		t.Fatal(err)
	}
	if meta.DatasetId != "d_abc" {
		t.Errorf("expected d_abc, got %s", meta.DatasetId)
	}
	if len(meta.ColumnMetadata.Order) != 1 {
		t.Errorf("expected 1 column, got %d", len(meta.ColumnMetadata.Order))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/client/... -run TestGetDatasetMetadata -v
```

Expected: FAIL — `c.GetDatasetMetadata undefined`

- [ ] **Step 3: Implement GetDatasetMetadata in `internal/client/dataset.go`**

Append to `internal/client/dataset.go`:

```go
func (c *DatasetClient) GetDatasetMetadata(datasetID string) (*DatasetMetadata, error) {
	url := fmt.Sprintf("%s/datasets/%s/metadata", c.v2Base, datasetID)
	resp, err := c.doGet(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp.StatusCode, "get_dataset_metadata failed")
	}

	var body metadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if body.ErrorMsg != "" {
		return nil, fmt.Errorf("api error: %s", body.ErrorMsg)
	}
	return &body.Data, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/client/... -run TestGetDatasetMetadata -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/client/dataset.go internal/client/dataset_test.go
git commit -m "feat: add DatasetClient.GetDatasetMetadata"
```

---

## Task 5: DatasetClient — ListCollections + GetCollectionInfo

**Files:**
- Modify: `internal/client/dataset.go`
- Modify: `internal/client/dataset_test.go`

- [ ] **Step 1: Write the failing tests**

Append to `internal/client/dataset_test.go`:

```go
func TestListCollections(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/public/api/collections" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(collectionsResponse{
			Code: 0,
			Data: CollectionsResult{
				Collections: []Collection{
					{CollectionId: "471", Name: "TradeNet Service Centres", ChildDatasets: []string{"d_ea81"}},
				},
			},
		})
	}))
	defer srv.Close()

	c := newTestDatasetClient(srv.URL)
	result, err := c.ListCollections(1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Collections) != 1 {
		t.Fatalf("expected 1 collection, got %d", len(result.Collections))
	}
	if result.Collections[0].CollectionId != "471" {
		t.Errorf("expected 471, got %s", result.Collections[0].CollectionId)
	}
}

func TestGetCollectionInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/public/api/collections/471/metadata" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(collectionMetaResponse{
			Code: 0,
			Data: struct {
				CollectionMetadata Collection `json:"collectionMetadata"`
			}{
				CollectionMetadata: Collection{
					CollectionId:  "471",
					Name:          "TradeNet Service Centres",
					ChildDatasets: []string{"d_ea81"},
				},
			},
		})
	}))
	defer srv.Close()

	c := newTestDatasetClient(srv.URL)
	info, err := c.GetCollectionInfo("471")
	if err != nil {
		t.Fatal(err)
	}
	if info.CollectionId != "471" {
		t.Errorf("expected 471, got %s", info.CollectionId)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/client/... -run "TestListCollections|TestGetCollectionInfo" -v
```

Expected: FAIL — methods undefined

- [ ] **Step 3: Implement both methods in `internal/client/dataset.go`**

Append to `internal/client/dataset.go`:

```go
func (c *DatasetClient) ListCollections(page, limit int) (*CollectionsResult, error) {
	url := fmt.Sprintf("%s/collections?page=%d&resultPerPage=%d", c.v2Base, page, limit)
	resp, err := c.doGet(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp.StatusCode, "list_collections failed")
	}

	var body collectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if body.ErrorMsg != "" {
		return nil, fmt.Errorf("api error: %s", body.ErrorMsg)
	}
	return &body.Data, nil
}

func (c *DatasetClient) GetCollectionInfo(collectionID string) (*Collection, error) {
	url := fmt.Sprintf("%s/collections/%s/metadata", c.v2Base, collectionID)
	resp, err := c.doGet(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp.StatusCode, "get_collection_info failed")
	}

	var body collectionMetaResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if body.ErrorMsg != "" {
		return nil, fmt.Errorf("api error: %s", body.ErrorMsg)
	}
	col := body.Data.CollectionMetadata
	return &col, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/client/... -run "TestListCollections|TestGetCollectionInfo" -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/client/dataset.go internal/client/dataset_test.go
git commit -m "feat: add DatasetClient.ListCollections and GetCollectionInfo"
```

---

## Task 6: DatasetClient — QueryDataset (CKAN)

**Files:**
- Modify: `internal/client/dataset.go`
- Modify: `internal/client/dataset_test.go`

- [ ] **Step 1: Write the failing test**

Append to `internal/client/dataset_test.go`:

```go
func TestQueryDataset(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/action/datastore_search" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("resource_id") != "d_abc" {
			t.Errorf("unexpected resource_id: %s", r.URL.Query().Get("resource_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ckanResponse{
			Success: true,
			Result: CKANResult{
				Fields:  []CKANField{{ID: "period", Type: "text"}, {ID: "_id", Type: "int4"}},
				Records: []map[string]interface{}{{"_id": 1, "period": "1992-03"}},
				Total:   1,
			},
		})
	}))
	defer srv.Close()

	c := newTestDatasetClient(srv.URL)
	result, err := c.QueryDataset("d_abc", nil, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
	if len(result.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(result.Records))
	}
}

func TestQueryDataset_WithFilters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filters := r.URL.Query().Get("filters")
		if filters == "" {
			t.Error("expected filters param")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ckanResponse{
			Success: true,
			Result:  CKANResult{Fields: []CKANField{}, Records: []map[string]interface{}{}, Total: 0},
		})
	}))
	defer srv.Close()

	c := newTestDatasetClient(srv.URL)
	_, err := c.QueryDataset("d_abc", map[string]string{"residential_status": "overall"}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestQueryDataset_CKANError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ckanResponse{
			Success: false,
			Error: &struct {
				Type    string   `json:"__type"`
				Message []string `json:"message"`
			}{Type: "Validation Error", Message: []string{"invalid resource_id"}},
		})
	}))
	defer srv.Close()

	c := newTestDatasetClient(srv.URL)
	_, err := c.QueryDataset("bad_id", nil, 10, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/client/... -run "TestQueryDataset" -v
```

Expected: FAIL — `c.QueryDataset undefined`

- [ ] **Step 3: Add remaining imports and QueryDataset to `internal/client/dataset.go`**

Update the import block at the top of `internal/client/dataset.go` to add `"net/url"`, `"strconv"`, and `"strings"` (encoding/json was already added in Task 3):

```go
import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)
```

Then append:

```go
func (c *DatasetClient) QueryDataset(datasetID string, filters map[string]string, limit, offset int) (*CKANResult, error) {
	params := url.Values{}
	params.Set("resource_id", datasetID)
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))
	if len(filters) > 0 {
		filterJSON, err := json.Marshal(filters)
		if err != nil {
			return nil, err
		}
		params.Set("filters", string(filterJSON))
	}

	fullURL := c.ckanURL + "?" + params.Encode()
	resp, err := c.doGet(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp.StatusCode, "query_dataset failed")
	}

	var body ckanResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if !body.Success {
		msg := "ckan error"
		if body.Error != nil && len(body.Error.Message) > 0 {
			msg = strings.Join(body.Error.Message, "; ")
		}
		return nil, fmt.Errorf("query failed: %s", msg)
	}
	return &body.Result, nil
}
```

Update the import block at the top of `internal/client/dataset.go` to:

```go
import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)
```

- [ ] **Step 4: Run all client tests**

```bash
go test ./internal/client/... -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/client/dataset.go internal/client/dataset_test.go
git commit -m "feat: add DatasetClient.QueryDataset with CKAN filters"
```

---

## Task 7: RealtimeClient — struct + GetEnvironmentReading

**Files:**
- Create: `internal/client/realtime.go`
- Create: `internal/client/realtime_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/client/realtime_test.go`:

```go
package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestRealtimeClient(serverURL string) *RealtimeClient {
	return &RealtimeClient{
		http:    &http.Client{Timeout: 5 * time.Second},
		baseURL: serverURL + "/v1",
	}
}

func TestGetEnvironmentReading_AirTemperature(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/environment/air-temperature" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"metadata": map[string]interface{}{
				"stations": []map[string]interface{}{
					{"id": "S107", "name": "East Coast Parkway", "location": map[string]float64{"latitude": 1.31, "longitude": 103.96}},
				},
				"reading_type": "DBT 1M F",
				"reading_unit": "deg C",
			},
			"items": []map[string]interface{}{
				{"timestamp": "2026-05-06T14:00:00+08:00", "readings": []map[string]interface{}{
					{"station_id": "S107", "value": 29.5},
				}},
			},
		})
	}))
	defer srv.Close()

	c := newTestRealtimeClient(srv.URL)
	result, err := c.GetEnvironmentReading("air-temperature")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "S107") {
		t.Errorf("expected S107 in result, got: %s", result)
	}
}

func TestGetEnvironmentReading_InvalidType(t *testing.T) {
	c := &RealtimeClient{http: &http.Client{}, baseURL: "http://unused"}
	_, err := c.GetEnvironmentReading("invalid-type")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/client/... -run "TestGetEnvironmentReading" -v
```

Expected: FAIL — `RealtimeClient undefined`

- [ ] **Step 3: Create `internal/client/realtime.go`**

```go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const realtimeBaseURL = "https://api.data.gov.sg/v1"

var validEnvironmentTypes = map[string]bool{
	"air-temperature":   true,
	"rainfall":          true,
	"relative-humidity": true,
	"wind-speed":        true,
	"wind-direction":    true,
	"uv-index":          true,
	"psi":               true,
	"pm25":              true,
}

var validForecastTypes = map[string]string{
	"2-hour":  "2-hour-weather-forecast",
	"24-hour": "24-hour-weather-forecast",
	"4-day":   "4-day-weather-forecast",
}

var validTransportTypes = map[string]string{
	"taxi":            "taxi-availability",
	"carpark":         "carpark-availability",
	"traffic-cameras": "traffic-images",
}

type RealtimeClient struct {
	http    *http.Client
	baseURL string
}

func NewRealtimeClient() *RealtimeClient {
	return &RealtimeClient{
		http:    &http.Client{Timeout: 15 * time.Second},
		baseURL: realtimeBaseURL,
	}
}

func (c *RealtimeClient) doGet(url string) ([]byte, error) {
	var (
		body []byte
		err  error
	)
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
		req, reqErr := http.NewRequest(http.MethodGet, url, nil)
		if reqErr != nil {
			return nil, reqErr
		}
		req.Header.Set("User-Agent", userAgent)

		resp, doErr := c.http.Do(req)
		if doErr != nil {
			err = doErr
			continue
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			resp.Body.Close()
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, apiError(resp.StatusCode, url)
		}

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		body = buf.Bytes()
		break
	}
	return body, err
}

func prettyJSON(data []byte) (string, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return "", err
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (c *RealtimeClient) GetEnvironmentReading(readingType string) (string, error) {
	if !validEnvironmentTypes[readingType] {
		valid := make([]string, 0, len(validEnvironmentTypes))
		for k := range validEnvironmentTypes {
			valid = append(valid, k)
		}
		return "", fmt.Errorf("invalid environment type %q; valid types: %v", readingType, valid)
	}
	url := fmt.Sprintf("%s/environment/%s", c.baseURL, readingType)
	data, err := c.doGet(url)
	if err != nil {
		return "", err
	}
	return prettyJSON(data)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/client/... -run "TestGetEnvironmentReading" -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/client/realtime.go internal/client/realtime_test.go
git commit -m "feat: add RealtimeClient with GetEnvironmentReading"
```

---

## Task 8: RealtimeClient — GetWeatherForecast + GetTransportInfo

**Files:**
- Modify: `internal/client/realtime.go`
- Modify: `internal/client/realtime_test.go`

- [ ] **Step 1: Write the failing tests**

Append to `internal/client/realtime_test.go`:

```go
func TestGetWeatherForecast_2Hour(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/environment/2-hour-weather-forecast" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"area_metadata": []map[string]interface{}{
				{"name": "Ang Mo Kio", "label_location": map[string]float64{"latitude": 1.375, "longitude": 103.839}},
			},
			"items": []map[string]interface{}{
				{"update_timestamp": "2026-05-06T14:00:00+08:00", "forecasts": []map[string]interface{}{
					{"area": "Ang Mo Kio", "forecast": "Partly Cloudy"},
				}},
			},
		})
	}))
	defer srv.Close()

	c := newTestRealtimeClient(srv.URL)
	result, err := c.GetWeatherForecast("2-hour")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "Ang Mo Kio") {
		t.Errorf("expected Ang Mo Kio in result, got: %s", result)
	}
}

func TestGetWeatherForecast_InvalidType(t *testing.T) {
	c := &RealtimeClient{http: &http.Client{}, baseURL: "http://unused"}
	_, err := c.GetWeatherForecast("bad-type")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestGetTransportInfo_Carpark(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/transport/carpark-availability" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"timestamp": "2026-05-06T14:00:00+08:00", "carpark_data": []map[string]interface{}{
					{"carpark_number": "HE12", "carpark_info": []map[string]interface{}{
						{"lot_type": "C", "lots_available": "101", "total_lots": "105"},
					}},
				}},
			},
		})
	}))
	defer srv.Close()

	c := newTestRealtimeClient(srv.URL)
	result, err := c.GetTransportInfo("carpark")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "HE12") {
		t.Errorf("expected HE12 in result, got: %s", result)
	}
}

func TestGetTransportInfo_InvalidType(t *testing.T) {
	c := &RealtimeClient{http: &http.Client{}, baseURL: "http://unused"}
	_, err := c.GetTransportInfo("bad-type")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/client/... -run "TestGetWeatherForecast|TestGetTransportInfo" -v
```

Expected: FAIL — methods undefined

- [ ] **Step 3: Append GetWeatherForecast and GetTransportInfo to `internal/client/realtime.go`**

```go
func (c *RealtimeClient) GetWeatherForecast(forecastType string) (string, error) {
	endpoint, ok := validForecastTypes[forecastType]
	if !ok {
		return "", fmt.Errorf("invalid forecast type %q; valid types: 2-hour, 24-hour, 4-day", forecastType)
	}
	url := fmt.Sprintf("%s/environment/%s", c.baseURL, endpoint)
	data, err := c.doGet(url)
	if err != nil {
		return "", err
	}
	return prettyJSON(data)
}

func (c *RealtimeClient) GetTransportInfo(infoType string) (string, error) {
	endpoint, ok := validTransportTypes[infoType]
	if !ok {
		return "", fmt.Errorf("invalid transport type %q; valid types: taxi, carpark, traffic-cameras", infoType)
	}
	url := fmt.Sprintf("%s/transport/%s", c.baseURL, endpoint)
	data, err := c.doGet(url)
	if err != nil {
		return "", err
	}
	return prettyJSON(data)
}
```

- [ ] **Step 4: Run all client tests**

```bash
go test ./internal/client/... -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/client/realtime.go internal/client/realtime_test.go
git commit -m "feat: add RealtimeClient.GetWeatherForecast and GetTransportInfo"
```

---

## Task 9: Dataset MCP tools

**Files:**
- Create: `internal/tools/datasets.go`
- Create: `internal/tools/datasets_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/tools/datasets_test.go`:

```go
package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sausheong/datagovsg-mcp/internal/client"
)

func newTestDatasetClientForTools(serverURL string) *client.DatasetClient {
	return client.NewDatasetClientWithConfig(serverURL+"/v2/public/api", serverURL+"/api/action/datastore_search", 5*time.Second)
}

func callTool(s *server.MCPServer, name string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	return s.CallTool(context.Background(), mcp.CallToolRequest{
		Params: struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			Name:      name,
			Arguments: args,
		},
	})
}

func TestSearchDatasetsTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"datasets": []map[string]interface{}{
					{"datasetId": "d_abc", "name": "Test Dataset", "format": "CSV"},
				},
				"pages": 1, "totalRowCount": 1,
			},
			"errorMsg": "",
		})
	}))
	defer srv.Close()

	s := server.NewMCPServer("test", "1.0")
	RegisterDatasetTools(s, newTestDatasetClientForTools(srv.URL))

	result, err := callTool(s, "search_datasets", map[string]interface{}{"query": "test"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "d_abc") {
		t.Errorf("expected d_abc in output, got: %s", text)
	}
}

func TestGetDatasetMetadataTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"datasetId": "d_abc", "name": "Test Dataset",
				"columnMetadata": map[string]interface{}{
					"order": []string{"c_1"},
					"map":   map[string]string{"c_1": "period"},
					"metaMapping": map[string]interface{}{
						"c_1": map[string]interface{}{"name": "period", "columnTitle": "Period", "dataType": "Text", "isCategorical": false},
					},
				},
			},
			"errorMsg": "",
		})
	}))
	defer srv.Close()

	s := server.NewMCPServer("test", "1.0")
	RegisterDatasetTools(s, newTestDatasetClientForTools(srv.URL))

	result, err := callTool(s, "get_dataset_metadata", map[string]interface{}{"dataset_id": "d_abc"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "period") {
		t.Errorf("expected period column in output, got: %s", text)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/tools/... -run "TestSearchDatasetsTool|TestGetDatasetMetadataTool" -v
```

Expected: FAIL — `RegisterDatasetTools undefined`, `NewDatasetClientWithConfig undefined`

- [ ] **Step 3: Add `NewDatasetClientWithConfig` to `internal/client/dataset.go`**

Append to `internal/client/dataset.go`:

```go
func NewDatasetClientWithConfig(v2Base, ckanURL string, timeout time.Duration) *DatasetClient {
	return &DatasetClient{
		http:    &http.Client{Timeout: timeout},
		v2Base:  v2Base,
		ckanURL: ckanURL,
	}
}
```

- [ ] **Step 4: Create `internal/tools/datasets.go`**

```go
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sausheong/datagovsg-mcp/internal/client"
)

func RegisterDatasetTools(s *server.MCPServer, c *client.DatasetClient) {
	s.AddTool(mcp.NewTool("search_datasets",
		mcp.WithDescription("Search data.gov.sg datasets by keyword. Returns dataset IDs, names, descriptions, formats, and coverage dates. Use this first to find relevant datasets before querying their data."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Keyword or topic to search for, e.g. 'unemployment', 'rainfall', 'HDB'")),
		mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
		mcp.WithNumber("limit", mcp.Description("Results per page, max 100 (default: 10)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, ok := req.Params.Arguments["query"].(string)
		if !ok || query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}
		page := 1
		if p, ok := req.Params.Arguments["page"].(float64); ok && p > 0 {
			page = int(p)
		}
		limit := 10
		if l, ok := req.Params.Arguments["limit"].(float64); ok && l > 0 {
			limit = int(l)
			if limit > 100 {
				limit = 100
			}
		}
		result, err := c.SearchDatasets(query, page, limit)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})

	s.AddTool(mcp.NewTool("get_dataset_metadata",
		mcp.WithDescription("Get the full schema for a dataset: column names, data types, units, and categorical flags. Call this after search_datasets to understand a dataset's structure before querying it."),
		mcp.WithString("dataset_id", mcp.Required(), mcp.Description("Dataset ID from search_datasets, e.g. 'd_ca32584c91ee07d091a4ce75fa868414'")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, ok := req.Params.Arguments["dataset_id"].(string)
		if !ok || id == "" {
			return mcp.NewToolResultError("dataset_id is required"), nil
		}
		meta, err := c.GetDatasetMetadata(id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("metadata fetch failed: %v", err)), nil
		}
		out, _ := json.MarshalIndent(meta, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})

	s.AddTool(mcp.NewTool("list_collections",
		mcp.WithDescription("Browse thematic collections of related datasets on data.gov.sg. Each collection groups datasets from the same agency or topic. Returns collection IDs and their child dataset IDs."),
		mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
		mcp.WithNumber("limit", mcp.Description("Results per page, max 100 (default: 10)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		page := 1
		if p, ok := req.Params.Arguments["page"].(float64); ok && p > 0 {
			page = int(p)
		}
		limit := 10
		if l, ok := req.Params.Arguments["limit"].(float64); ok && l > 0 {
			limit = int(l)
			if limit > 100 {
				limit = 100
			}
		}
		result, err := c.ListCollections(page, limit)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_collections failed: %v", err)), nil
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})

	s.AddTool(mcp.NewTool("get_collection_info",
		mcp.WithDescription("Get full details for a specific collection, including its child dataset IDs. Use after list_collections to see which datasets belong to a collection."),
		mcp.WithString("collection_id", mcp.Required(), mcp.Description("Collection ID from list_collections, e.g. '471'")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, ok := req.Params.Arguments["collection_id"].(string)
		if !ok || id == "" {
			return mcp.NewToolResultError("collection_id is required"), nil
		}
		info, err := c.GetCollectionInfo(id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_collection_info failed: %v", err)), nil
		}
		out, _ := json.MarshalIndent(info, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/tools/... -run "TestSearchDatasetsTool|TestGetDatasetMetadataTool" -v
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/client/dataset.go internal/tools/datasets.go internal/tools/datasets_test.go
git commit -m "feat: add dataset MCP tools (search, metadata, collections)"
```

---

## Task 10: Query MCP tool

**Files:**
- Create: `internal/tools/query.go`
- Create: `internal/tools/query_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/tools/query_test.go`:

```go
package tools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sausheong/datagovsg-mcp/internal/client"
)

func TestQueryDatasetTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"result": map[string]interface{}{
				"resource_id": "d_abc",
				"fields":      []map[string]interface{}{{"id": "period", "type": "text"}},
				"records":     []map[string]interface{}{{"period": "1992-03", "_id": 1}},
				"total":       1,
			},
		})
	}))
	defer srv.Close()

	s := server.NewMCPServer("test", "1.0")
	RegisterQueryTool(s, newTestDatasetClientForTools(srv.URL))

	result, err := callTool(s, "query_dataset", map[string]interface{}{
		"dataset_id": "d_abc",
		"limit":      float64(5),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "1992-03") {
		t.Errorf("expected 1992-03 in output, got: %s", text)
	}
}

func TestQueryDatasetTool_WithFilters(t *testing.T) {
	var receivedFilters string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedFilters = r.URL.Query().Get("filters")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"fields": []interface{}{}, "records": []interface{}{}, "total": 0},
		})
	}))
	defer srv.Close()

	s := server.NewMCPServer("test", "1.0")
	RegisterQueryTool(s, newTestDatasetClientForTools(srv.URL))

	_, err := callTool(s, "query_dataset", map[string]interface{}{
		"dataset_id": "d_abc",
		"filters":    map[string]interface{}{"residential_status": "overall"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if receivedFilters == "" {
		t.Error("expected filters to be passed to upstream API")
	}
}

func TestQueryDatasetTool_MissingDatasetID(t *testing.T) {
	s := server.NewMCPServer("test", "1.0")
	RegisterQueryTool(s, newTestDatasetClientForTools("http://unused"))

	result, err := callTool(s, "query_dataset", map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for missing dataset_id")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/tools/... -run "TestQueryDatasetTool" -v
```

Expected: FAIL — `RegisterQueryTool undefined`

- [ ] **Step 3: Create `internal/tools/query.go`**

```go
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sausheong/datagovsg-mcp/internal/client"
)

func RegisterQueryTool(s *server.MCPServer, c *client.DatasetClient) {
	s.AddTool(mcp.NewTool("query_dataset",
		mcp.WithDescription("Query actual records from a data.gov.sg dataset. Use get_dataset_metadata first to learn column names for filters. Supports column-level exact-match filters. Returns records, field types, and total count."),
		mcp.WithString("dataset_id", mcp.Required(), mcp.Description("Dataset ID, e.g. 'd_ca32584c91ee07d091a4ce75fa868414'")),
		mcp.WithObject("filters", mcp.Description("Optional column filters as key/value pairs. Key is column name, value is exact match string. Example: {\"residential_status\": \"overall\"}")),
		mcp.WithNumber("limit", mcp.Description("Max records to return, max 100 (default: 20)")),
		mcp.WithNumber("offset", mcp.Description("Number of records to skip for pagination (default: 0)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		datasetID, ok := req.Params.Arguments["dataset_id"].(string)
		if !ok || datasetID == "" {
			return mcp.NewToolResultError("dataset_id is required"), nil
		}

		limit := 20
		if l, ok := req.Params.Arguments["limit"].(float64); ok && l > 0 {
			limit = int(l)
			if limit > 100 {
				limit = 100
			}
		}

		offset := 0
		if o, ok := req.Params.Arguments["offset"].(float64); ok && o >= 0 {
			offset = int(o)
		}

		var filters map[string]string
		if f, ok := req.Params.Arguments["filters"].(map[string]interface{}); ok {
			filters = make(map[string]string, len(f))
			for k, v := range f {
				if s, ok := v.(string); ok {
					filters[k] = s
				}
			}
		}

		result, err := c.QueryDataset(datasetID, filters, limit, offset)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("query failed: %v", err)), nil
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/tools/... -run "TestQueryDatasetTool" -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tools/query.go internal/tools/query_test.go
git commit -m "feat: add query_dataset MCP tool"
```

---

## Task 11: Realtime MCP tools

**Files:**
- Create: `internal/tools/realtime.go`
- Create: `internal/tools/realtime_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/tools/realtime_test.go`:

```go
package tools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sausheong/datagovsg-mcp/internal/client"
)

func newTestRealtimeClientForTools(serverURL string) *client.RealtimeClient {
	return client.NewRealtimeClientWithConfig(serverURL+"/v1", 5*time.Second)
}

func TestGetEnvironmentReadingTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"metadata": map[string]interface{}{"reading_unit": "deg C", "stations": []interface{}{}},
			"items":    []map[string]interface{}{{"timestamp": "2026-05-06T14:00:00+08:00", "readings": []interface{}{}}},
		})
	}))
	defer srv.Close()

	s := server.NewMCPServer("test", "1.0")
	RegisterRealtimeTools(s, newTestRealtimeClientForTools(srv.URL))

	result, err := callTool(s, "get_environment_reading", map[string]interface{}{"type": "air-temperature"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %v", result.Content)
	}
}

func TestGetEnvironmentReadingTool_InvalidType(t *testing.T) {
	s := server.NewMCPServer("test", "1.0")
	RegisterRealtimeTools(s, newTestRealtimeClientForTools("http://unused"))

	result, err := callTool(s, "get_environment_reading", map[string]interface{}{"type": "bad-type"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for invalid type")
	}
}

func TestGetWeatherForecastTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"update_timestamp": "2026-05-06T14:00:00+08:00", "forecasts": []map[string]interface{}{
					{"area": "Ang Mo Kio", "forecast": "Partly Cloudy"},
				}},
			},
		})
	}))
	defer srv.Close()

	s := server.NewMCPServer("test", "1.0")
	RegisterRealtimeTools(s, newTestRealtimeClientForTools(srv.URL))

	result, err := callTool(s, "get_weather_forecast", map[string]interface{}{"type": "2-hour"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "Partly Cloudy") {
		t.Errorf("expected forecast in output, got: %s", text)
	}
}

func TestGetTransportInfoTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"timestamp": "2026-05-06T14:00:00+08:00", "carpark_data": []map[string]interface{}{
					{"carpark_number": "HE12"},
				}},
			},
		})
	}))
	defer srv.Close()

	s := server.NewMCPServer("test", "1.0")
	RegisterRealtimeTools(s, newTestRealtimeClientForTools(srv.URL))

	result, err := callTool(s, "get_transport_info", map[string]interface{}{"type": "carpark"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "HE12") {
		t.Errorf("expected HE12 in output, got: %s", text)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/tools/... -run "TestGetEnvironmentReadingTool|TestGetWeatherForecastTool|TestGetTransportInfoTool" -v
```

Expected: FAIL — `RegisterRealtimeTools undefined`, `NewRealtimeClientWithConfig undefined`

- [ ] **Step 3: Add `NewRealtimeClientWithConfig` to `internal/client/realtime.go`**

Append to `internal/client/realtime.go`:

```go
func NewRealtimeClientWithConfig(baseURL string, timeout time.Duration) *RealtimeClient {
	return &RealtimeClient{
		http:    &http.Client{Timeout: timeout},
		baseURL: baseURL,
	}
}
```

- [ ] **Step 4: Create `internal/tools/realtime.go`**

```go
package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sausheong/datagovsg-mcp/internal/client"
)

func RegisterRealtimeTools(s *server.MCPServer, c *client.RealtimeClient) {
	s.AddTool(mcp.NewTool("get_environment_reading",
		mcp.WithDescription("Get current environmental sensor readings from NEA stations across Singapore. Each reading includes timestamp, station names, GPS coordinates, and measured values."),
		mcp.WithString("type", mcp.Required(), mcp.Description(
			"Type of reading. One of: air-temperature (°C), rainfall (mm), relative-humidity (%), wind-speed (km/h), wind-direction (degrees), uv-index, psi (Pollutant Standards Index), pm25 (PM2.5 by region)",
		)),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		readingType, ok := req.Params.Arguments["type"].(string)
		if !ok || readingType == "" {
			return mcp.NewToolResultError("type is required"), nil
		}
		result, err := c.GetEnvironmentReading(readingType)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("environment reading failed: %v", err)), nil
		}
		return mcp.NewToolResultText(result), nil
	})

	s.AddTool(mcp.NewTool("get_weather_forecast",
		mcp.WithDescription("Get weather forecasts for Singapore. The 2-hour forecast gives area-level nowcasts; 24-hour gives regional conditions for today; 4-day gives an island-wide outlook with temperature ranges."),
		mcp.WithString("type", mcp.Required(), mcp.Description(
			"Forecast window. One of: 2-hour (area-level nowcast), 24-hour (regional forecast for today), 4-day (island-wide outlook)",
		)),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		forecastType, ok := req.Params.Arguments["type"].(string)
		if !ok || forecastType == "" {
			return mcp.NewToolResultError("type is required"), nil
		}
		result, err := c.GetWeatherForecast(forecastType)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("weather forecast failed: %v", err)), nil
		}
		return mcp.NewToolResultText(result), nil
	})

	s.AddTool(mcp.NewTool("get_transport_info",
		mcp.WithDescription("Get live transport data from Singapore. Taxi returns GPS coordinates of available taxis. Carpark returns lot availability for 2,001 HDB carparks. Traffic-cameras returns image URLs and locations for traffic cameras."),
		mcp.WithString("type", mcp.Required(), mcp.Description(
			"Transport data type. One of: taxi (live taxi GPS locations), carpark (HDB carpark lot availability), traffic-cameras (live traffic camera image URLs and locations)",
		)),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		infoType, ok := req.Params.Arguments["type"].(string)
		if !ok || infoType == "" {
			return mcp.NewToolResultError("type is required"), nil
		}
		result, err := c.GetTransportInfo(infoType)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("transport info failed: %v", err)), nil
		}
		return mcp.NewToolResultText(result), nil
	})
}
```

- [ ] **Step 5: Run all tests**

```bash
go test ./... -v
```

Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add internal/client/realtime.go internal/tools/realtime.go internal/tools/realtime_test.go
git commit -m "feat: add realtime MCP tools (environment, forecast, transport)"
```

---

## Task 12: Main server

**Files:**
- Create: `main.go`

- [ ] **Step 1: Create `main.go`**

```go
package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/sausheong/datagovsg-mcp/internal/client"
	"github.com/sausheong/datagovsg-mcp/internal/tools"
)

func main() {
	datasetClient := client.NewDatasetClient()
	realtimeClient := client.NewRealtimeClient()

	s := server.NewMCPServer(
		"datagovsg-mcp",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	tools.RegisterDatasetTools(s, datasetClient)
	tools.RegisterQueryTool(s, datasetClient)
	tools.RegisterRealtimeTools(s, realtimeClient)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Build the binary**

```bash
go build -o datagovsg-mcp .
```

Expected: binary `datagovsg-mcp` created in current directory, no errors.

- [ ] **Step 3: Run all tests one final time**

```bash
go test ./... -v
```

Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add main.go
git commit -m "feat: wire MCP server in main.go"
```

---

## Task 13: Wire into Claude Code as MCP server

**Files:**
- Create or modify: `~/.claude/claude_desktop_config.json` (or project `.claude/settings.json`)

- [ ] **Step 1: Get the binary path**

```bash
pwd
```

Note the full path. The binary is at `<pwd>/datagovsg-mcp`.

- [ ] **Step 2: Add to Claude Code MCP config**

Add the server to `.claude/settings.json` in this project directory:

```json
{
  "mcpServers": {
    "datagovsg": {
      "command": "/Users/sausheong/projects/datagovsg_mcp/datagovsg-mcp",
      "args": []
    }
  }
}
```

- [ ] **Step 3: Restart Claude Code and verify tools appear**

In Claude Code, run `/mcp` and confirm `datagovsg` appears with 8 tools listed:
- `search_datasets`
- `get_dataset_metadata`
- `list_collections`
- `get_collection_info`
- `query_dataset`
- `get_environment_reading`
- `get_weather_forecast`
- `get_transport_info`

- [ ] **Step 4: Smoke test each tool type**

Ask Claude:
1. "What datasets are available about Singapore's unemployment rate?" → should call `search_datasets`
2. "Show me the columns in dataset d_ca32584c91ee07d091a4ce75fa868414" → should call `get_dataset_metadata`
3. "What's the current air temperature in Singapore?" → should call `get_environment_reading`
4. "What's the weather forecast for the next 2 hours?" → should call `get_weather_forecast`
5. "How many carpark lots are available in Singapore right now?" → should call `get_transport_info`

- [ ] **Step 5: Final commit**

```bash
git add .claude/settings.json
git commit -m "chore: add MCP server config for Claude Code"
```
