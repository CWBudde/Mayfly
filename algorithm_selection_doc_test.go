package mayfly

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed docs/algorithm-selection.md
var algorithmSelectionGuide string

//go:embed cmd/algorithm-selection/main.go
var algorithmSelectionExample string

func TestAlgorithmSelectionGuideEmbedsRunnableExample(t *testing.T) {
	const openingFence = "```go\n"

	start := strings.Index(algorithmSelectionGuide, openingFence)
	if start < 0 {
		t.Fatal("algorithm-selection guide has no Go example")
	}

	start += len(openingFence)

	end := strings.Index(algorithmSelectionGuide[start:], "\n```")
	if end < 0 {
		t.Fatal("algorithm-selection guide has an unclosed Go example")
	}

	got := strings.TrimSpace(algorithmSelectionGuide[start : start+end])
	want := strings.TrimSpace(algorithmSelectionExample)

	if got != want {
		t.Fatal("first algorithm-selection Go block differs from cmd/algorithm-selection/main.go")
	}
}
