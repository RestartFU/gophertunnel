package minecraft

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestHandleResourcePacksInfoUsesConfiguredHTTPClient(t *testing.T) {
	id := uuid.New()
	packData := testResourcePackArchive(t, id)
	requested := false
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requested = true
		if request.URL.Host != "packs.example" {
			t.Fatalf("resource-pack request host = %q", request.URL.Host)
		}
		return &http.Response{
			StatusCode:    http.StatusOK,
			Status:        "200 OK",
			ContentLength: int64(len(packData)),
			Body:          io.NopCloser(bytes.NewReader(packData)),
			Header:        make(http.Header),
		}, nil
	})}

	conn := testConn()
	conn.ctx = context.Background()
	conn.resourcePackHTTPClient = client
	err := conn.handleResourcePacksInfo(&packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{{
		UUID:        id,
		Version:     "1.0.0",
		Size:        uint64(len(packData)),
		DownloadURL: "https://packs.example/resource.mcpack",
	}}})
	if err != nil {
		t.Fatalf("handleResourcePacksInfo(): %v", err)
	}
	if !requested {
		t.Fatal("configured resource-pack HTTP client was not used")
	}
	if len(conn.resourcePacks) != 1 {
		t.Fatalf("downloaded packs = %d, want 1", len(conn.resourcePacks))
	}
}

func TestHandleResourcePacksInfoRejectsTooManyPacks(t *testing.T) {
	packs := make([]protocol.TexturePackInfo, maxResourcePackCount+1)
	for i := range packs {
		packs[i] = protocol.TexturePackInfo{UUID: uuid.New(), Version: "1.0.0", Size: 1}
	}

	err := testConn().handleResourcePacksInfo(&packet.ResourcePacksInfo{TexturePacks: packs})
	if err == nil || !strings.Contains(err.Error(), "too many") {
		t.Fatalf("handleResourcePacksInfo() error = %v, want pack-count error", err)
	}
}

func TestHandleResourcePacksInfoRejectsOversizedPack(t *testing.T) {
	err := testConn().handleResourcePacksInfo(&packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{{
		UUID:    uuid.New(),
		Version: "1.0.0",
		Size:    maxResourcePackTotalSize + 1,
	}}})
	if err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("handleResourcePacksInfo() error = %v, want size-limit error", err)
	}
}

func TestHandleResourcePacksInfoRejectsOversizedAggregate(t *testing.T) {
	err := testConn().handleResourcePacksInfo(&packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{
		{UUID: uuid.New(), Version: "1.0.0", Size: maxResourcePackTotalSize},
		{UUID: uuid.New(), Version: "1.0.0", Size: 1},
	}})
	if err == nil || !strings.Contains(err.Error(), "aggregate") {
		t.Fatalf("handleResourcePacksInfo() error = %v, want aggregate-size error", err)
	}
}
