package mayfly

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed docs/getting-started.md
var gettingStartedGuide string

//go:embed cmd/getting-started/main.go
var gettingStartedExample string

func TestGettingStartedGuideEmbedsRunnableExample(t *testing.T) {
	const openingFence = "```go\n"

	start := strings.Index(gettingStartedGuide, openingFence)
	if start < 0 {
		t.Fatal("getting-started guide has no Go example")
	}

	start += len(openingFence)

	end := strings.Index(gettingStartedGuide[start:], "\n```")
	if end < 0 {
		t.Fatal("getting-started guide has an unclosed Go example")
	}

	got := strings.TrimSpace(gettingStartedGuide[start : start+end])
	want := strings.TrimSpace(gettingStartedExample)

	if got != want {
		t.Fatal("first getting-started Go block differs from cmd/getting-started/main.go")
	}
}
