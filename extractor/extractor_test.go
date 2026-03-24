package extractor

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestExtractMatchesReferencePosts(t *testing.T) {
	carFile, err := os.Open(filepath.Join("..", "ref", "repo.car"))
	if err != nil {
		t.Fatalf("open sample repo car: %v", err)
	}
	t.Cleanup(func() {
		_ = carFile.Close()
	})

	got, err := Extract(context.Background(), carFile)
	if err != nil {
		t.Fatalf("extract posts: %v", err)
	}

	if got.Did != "did:plc:iejbew3dkphs4lfkhhqd2ly6" {
		t.Fatalf("unexpected did: %s", got.Did)
	}

	want := loadReferencePosts(t)
	if !reflect.DeepEqual(got.Posts, want) {
		t.Fatalf("extracted posts do not match reference output")
	}
}

func TestMarshalPostsCSV(t *testing.T) {
	posts := []Post{
		{
			CreatedAt: "2025-03-06T20:08:56.020Z",
			Text:      "plain text",
		},
		{
			CreatedAt: "2025-03-07T00:00:00.000Z",
			Text:      "quote \"and\"\nnewline",
		},
	}

	raw, err := MarshalPostsCSV(posts)
	if err != nil {
		t.Fatalf("marshal csv: %v", err)
	}

	rows, err := csv.NewReader(bytes.NewReader(raw)).ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}

	want := [][]string{
		{"createdAt", "text"},
		{"2025-03-06T20:08:56.020Z", "plain text"},
		{"2025-03-07T00:00:00.000Z", "quote \"and\"\nnewline"},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("unexpected csv rows: %#v", rows)
	}
}

func loadReferencePosts(t *testing.T) []Post {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join("..", "ref", "app.bsky.feed.post"))
	if err != nil {
		t.Fatalf("read reference directory: %v", err)
	}

	posts := make([]Post, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		raw, err := os.ReadFile(filepath.Join("..", "ref", "app.bsky.feed.post", entry.Name()))
		if err != nil {
			t.Fatalf("read reference post %q: %v", entry.Name(), err)
		}

		var post Post
		if err := json.Unmarshal(raw, &post); err != nil {
			t.Fatalf("decode reference post %q: %v", entry.Name(), err)
		}

		posts = append(posts, post)
	}

	sort.Slice(posts, func(i, j int) bool {
		if posts[i].CreatedAt == posts[j].CreatedAt {
			return posts[i].Text < posts[j].Text
		}
		return posts[i].CreatedAt < posts[j].CreatedAt
	})

	return posts
}
