package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	v2BaseURL = "https://api-production.data.gov.sg/v2/public/api"
	ckanURL   = "https://data.gov.sg/api/action/datastore_search"
	userAgent = "datagovsg-mcp/1.0"
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

func NewDatasetClientWithConfig(v2Base, ckanURL string, timeout time.Duration) *DatasetClient {
	return &DatasetClient{
		http:    &http.Client{Timeout: timeout},
		v2Base:  v2Base,
		ckanURL: ckanURL,
	}
}
