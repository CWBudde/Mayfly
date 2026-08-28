package mayfly

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed docs/parameter-tuning.md
var parameterTuningGuide string

//go:embed cmd/parameter-tuning/main.go
var parameterTuningExample string

func TestParameterTuningGuideEmbedsRunnableExample(t *testing.T) {
	const openingFence = "```go\n"

	start := strings.Index(parameterTuningGuide, openingFence)
	if start < 0 {
		t.Fatal("parameter-tuning guide has no Go example")
	}

	start += len(openingFence)

	end := strings.Index(parameterTuningGuide[start:], "\n```")
	if end < 0 {
		t.Fatal("parameter-tuning guide has an unclosed Go example")
	}

	got := strings.TrimSpace(parameterTuningGuide[start : start+end])
	want := strings.TrimSpace(parameterTuningExample)

	if got != want {
		t.Fatal("first parameter-tuning Go block differs from cmd/parameter-tuning/main.go")
	}
}
