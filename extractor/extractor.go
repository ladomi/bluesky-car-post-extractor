// SPDX-License-Identifier: MIT OR Apache-2.0
//
// Adapted in part from github.com/bluesky-social/indigo/cmd/gosky/car.go.
// Modified in this project to extract only Bluesky post timestamps/text and
// serialize them for browser-side downloads.

package extractor

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/repo"
	"github.com/ipfs/go-cid"
)

const postCollectionPrefix = "app.bsky.feed.post/"

type Post struct {
	CreatedAt string `json:"createdAt"`
	Text      string `json:"text"`
}

type Result struct {
	Did   string `json:"did"`
	Posts []Post `json:"posts"`
}

func Extract(ctx context.Context, car io.Reader) (*Result, error) {
	repoData, err := repo.ReadRepoFromCar(ctx, car)
	if err != nil {
		return nil, fmt.Errorf("read repo from car: %w", err)
	}

	result := &Result{
		Did: repoData.RepoDid(),
	}

	err = repoData.ForEach(ctx, "", func(path string, _ cid.Cid) error {
		if !strings.HasPrefix(path, postCollectionPrefix) {
			return nil
		}

		_, rec, err := repoData.GetRecord(ctx, path)
		if err != nil {
			return fmt.Errorf("get record %q: %w", path, err)
		}

		post, ok := rec.(*bsky.FeedPost)
		if !ok {
			return fmt.Errorf("record %q is %T, expected *bsky.FeedPost", path, rec)
		}

		result.Posts = append(result.Posts, Post{
			CreatedAt: post.CreatedAt,
			Text:      post.Text,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(result.Posts, func(i, j int) bool {
		if result.Posts[i].CreatedAt == result.Posts[j].CreatedAt {
			return result.Posts[i].Text < result.Posts[j].Text
		}
		return result.Posts[i].CreatedAt < result.Posts[j].CreatedAt
	})

	return result, nil
}

func MarshalPosts(posts []Post) ([]byte, error) {
	return json.MarshalIndent(posts, "", "  ")
}

func MarshalPostsCSV(posts []Post) ([]byte, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	writer.UseCRLF = true

	if err := writer.Write([]string{"createdAt", "text"}); err != nil {
		return nil, err
	}

	for _, post := range posts {
		if err := writer.Write([]string{post.CreatedAt, post.Text}); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
