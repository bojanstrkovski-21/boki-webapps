package icon

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestLiveYouTubeIcon(t *testing.T) {
	if os.Getenv("BOKI_LIVE_ICON") != "1" {
		t.Skip()
	}
	t.Setenv("HOME", t.TempDir())
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	got, err := Lookup(ctx, "https://www.youtube.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("no icon")
	}
	for _, img := range got {
		t.Log(img.Source, img.Score, img.Path)
	}
}
