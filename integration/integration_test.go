package integration

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShorten(t *testing.T) {
	httpClient := http.DefaultClient

	// create
	createResp, err := httpClient.Post(
		"http://localhost:8080/api/shorten",
		"application/json",
		strings.NewReader(`{"url": "https://www.google.com"}`),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	location := createResp.Header.Get("location")
	require.NotEmpty(t, location)

	// retrieve
	getResp, err := httpClient.Get("http://localhost:8080" + location)
	require.NoError(t, err)
	require.Contains(t, getResp.Header.Get("content-type"), "application/json")
	var data map[string]interface{}
	require.NoError(t, json.NewDecoder(getResp.Body).Decode(&data))
	require.Equal(t, "8ffdefb", data["hash"])
	require.Equal(t, "https://www.google.com", data["url"])

	// list
	getListResp, err := httpClient.Get("http://localhost:8080/api/shorten?limit=100000")
	require.NoError(t, err)
	require.Contains(t, getResp.Header.Get("content-type"), "application/json")
	var listData []map[string]interface{}
	require.NoError(t, json.NewDecoder(getListResp.Body).Decode(&listData))
	require.True(t, len(listData) > 0)
	var found map[string]interface{}
	for _, d := range listData {
		if d["id"] == data["id"] {
			found = d
		}
	}
	require.Equal(t, data, found, "newly created shorten not found in the list")

	// delete
	deleteReq, err := http.NewRequest(http.MethodDelete, "http://localhost:8080"+location, nil)
	require.NoError(t, err)
	deleteResp, err := httpClient.Do(deleteReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	// attempt to find removed entity
	reGetResp, err := httpClient.Get("http://localhost:8080" + location)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, reGetResp.StatusCode)
}

func TestRedirect(t *testing.T) {
	httpClient := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// create
	createResp, err := httpClient.Post(
		"http://localhost:8080/api/shorten",
		"application/json",
		strings.NewReader(`{"url": "https://www.google.com"}`),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	location := createResp.Header.Get("location")

	// retrieve
	getResp, err := httpClient.Get("http://localhost:8080" + location)
	require.NoError(t, err)
	var data map[string]interface{}
	require.NoError(t, json.NewDecoder(getResp.Body).Decode(&data))

	// verify redirection
	resp, err := httpClient.Get("http://localhost/" + data["hash"].(string))
	require.NoError(t, err)
	require.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	require.Equal(t, "https://www.google.com", resp.Header.Get("location"))
}
