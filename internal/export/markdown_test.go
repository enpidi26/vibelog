package export

import (
	"strings"
	"testing"

	"github.com/enpidi26/vibelog/internal/domain"
)

func TestMarkdownEscapesFence(t *testing.T) {
	md := Markdown([]domain.Entry{{
		Title: "T", Model: "M", Tool: "X", Outcome: "success", IterationNo: 1,
		PromptText: "a ``` b",
	}})
	if strings.Contains(md, "a ``` b") {
		t.Fatal("fence di dalam prompt harus diamankan")
	}
	if !strings.Contains(md, "## No project") || !strings.Contains(md, "### T") {
		t.Fatalf("markdown missing sections: %s", md)
	}
}

func TestMarkdownGroupsByProject(t *testing.T) {
	md := Markdown([]domain.Entry{
		{Title: "A", ProjectName: "Alpha", PromptText: "p"},
		{Title: "B", ProjectName: "Beta", PromptText: "p"},
		{Title: "C", ProjectName: "Alpha", PromptText: "p"},
	})
	if strings.Index(md, "## Alpha") > strings.Index(md, "## Beta") {
		t.Fatal("project order should follow first appearance")
	}
	if strings.Count(md, "### A") != 1 || strings.Count(md, "### C") != 1 {
		t.Fatal("entries should appear once")
	}
}
