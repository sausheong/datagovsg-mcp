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
