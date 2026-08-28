package mayfly

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed docs/custom-objective-functions.md
var customObjectiveGuide string

//go:embed cmd/custom-objective/main.go
var customObjectiveExample string

func TestCustomObjectiveGuideEmbedsRunnableExample(t *testing.T) {
	const openingFence = "```go\n"

	start := strings.Index(customObjectiveGuide, openingFence)
	if start < 0 {
		t.Fatal("custom-objective guide has no Go example")
	}

	start += len(openingFence)

	end := strings.Index(customObjectiveGuide[start:], "\n```")
	if end < 0 {
		t.Fatal("custom-objective guide has an unclosed Go example")
	}

	got := strings.TrimSpace(customObjectiveGuide[start : start+end])
	want := strings.TrimSpace(customObjectiveExample)

	if got != want {
		t.Fatal("first custom-objective Go block differs from cmd/custom-objective/main.go")
	}
}
