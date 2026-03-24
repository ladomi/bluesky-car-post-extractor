package main

import (
	"bytes"
	"context"
	"encoding/json"
	"syscall/js"

	"github.com/bluesky-social/indigo/myapp/extractor"
)

const previewLimit = 24

type response struct {
	Did          string           `json:"did,omitempty"`
	PostCount    int              `json:"postCount"`
	RangeStart   string           `json:"rangeStart,omitempty"`
	RangeEnd     string           `json:"rangeEnd,omitempty"`
	Preview      []extractor.Post `json:"preview,omitempty"`
	DownloadJSON string           `json:"downloadJson,omitempty"`
	DownloadCSV  string           `json:"downloadCsv,omitempty"`
	Error        string           `json:"error,omitempty"`
}

func main() {
	js.Global().Set("extractBlueskyPostsFromCar", js.FuncOf(extractBlueskyPostsFromCar))
	js.Global().Set("blueskyCarExtractorReady", true)
	select {}
}

func extractBlueskyPostsFromCar(_ js.Value, args []js.Value) any {
	if len(args) != 1 {
		return marshalResponse(response{
			Error: "expected a Uint8Array argument",
		})
	}

	input := args[0]
	size := input.Get("length").Int()
	if size == 0 {
		return marshalResponse(response{
			Error: "input file is empty",
		})
	}

	carBytes := make([]byte, size)
	if copied := js.CopyBytesToGo(carBytes, input); copied != size {
		return marshalResponse(response{
			Error: "failed to copy input bytes from JavaScript",
		})
	}

	result, err := extractor.Extract(context.Background(), bytes.NewReader(carBytes))
	if err != nil {
		return marshalResponse(response{
			Error: err.Error(),
		})
	}

	downloadJSON, err := extractor.MarshalPosts(result.Posts)
	if err != nil {
		return marshalResponse(response{
			Error: err.Error(),
		})
	}

	downloadCSV, err := extractor.MarshalPostsCSV(result.Posts)
	if err != nil {
		return marshalResponse(response{
			Error: err.Error(),
		})
	}

	resp := response{
		Did:          result.Did,
		PostCount:    len(result.Posts),
		Preview:      previewPosts(result.Posts),
		DownloadJSON: string(downloadJSON),
		DownloadCSV:  string(downloadCSV),
	}
	if len(result.Posts) > 0 {
		resp.RangeStart = result.Posts[0].CreatedAt
		resp.RangeEnd = result.Posts[len(result.Posts)-1].CreatedAt
	}

	return marshalResponse(resp)
}

func previewPosts(posts []extractor.Post) []extractor.Post {
	if len(posts) <= previewLimit {
		return posts
	}
	return posts[:previewLimit]
}

func marshalResponse(resp response) string {
	payload, err := json.Marshal(resp)
	if err != nil {
		return `{"error":"failed to marshal response"}`
	}
	return string(payload)
}
