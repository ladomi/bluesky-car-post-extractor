# Notices

This directory contains code adapted from the Bluesky Social `indigo` project.

- Upstream repository: `github.com/bluesky-social/indigo`
- Referenced source when building this app: `cmd/gosky/car.go`
- Upstream license model: dual-licensed under MIT or Apache License 2.0, at the recipient's option

Material adaptation in this directory:

- `extractor/extractor.go`

Changes made in this project include:

- adapting the CAR unpack flow for browser/WASM execution
- filtering the export down to `app.bsky.feed.post`
- extracting only `createdAt` and `text`
- adding JSON and CSV serialization for download
- adding the static site UI in `myapp/`

The upstream license texts are redistributed in:

- `LICENSE-MIT`
- `LICENSE-APACHE`
