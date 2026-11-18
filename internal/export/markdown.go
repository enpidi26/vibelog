// Package export menyusun prompt pack dari entri jurnal.
package export

import (
	"fmt"
	"strings"
	"time"

	"github.com/enpidi26/vibelog/internal/domain"
)

// Markdown menyusun prompt pack: prompt dan catatan hasil per entri,
// dikelompokkan per proyek, siap ditempel ke AI agent atau editor.
// Fence code ``` di dalam teks prompt diamankan supaya tidak menutup blok.
func Markdown(entries []domain.Entry) string {
	var b strings.Builder
	b.WriteString("# VibeLog Prompt Pack\n\n")
	b.WriteString(fmt.Sprintf("Exported %s. %d entries.\n\n", time.Now().Format("2006-01-02"), len(entries)))
	groups := map[string][]domain.Entry{}
	order := []string{}
	for _, e := range entries {
		name := e.ProjectName
		if name == "" {
			name = "No project"
		}
		if _, ok := groups[name]; !ok {
			order = append(order, name)
		}
		groups[name] = append(groups[name], e)
	}
	for _, name := range order {
		b.WriteString("## " + name + "\n\n")
		for _, e := range groups[name] {
			b.WriteString("### " + e.Title + "\n\n")
			b.WriteString(fmt.Sprintf("- Model: %s | Tool: %s | Hasil: %s | Iterasi: %d\n", e.Model, e.Tool, e.Outcome, e.IterationNo))
			if e.Rating != nil {
				fmt.Fprintf(&b, "- Rating: %d/5\n", *e.Rating)
			}
			if len(e.Tags) > 0 {
				b.WriteString("- Tag: " + strings.Join(e.Tags, ", ") + "\n")
			}
			b.WriteString("\nPrompt:\n\n```text\n")
			b.WriteString(strings.ReplaceAll(e.PromptText, "```", "` ` `"))
			b.WriteString("\n```\n\n")
			if strings.TrimSpace(e.ResultNotes) != "" {
				b.WriteString("Hasil / catatan:\n\n")
				b.WriteString(e.ResultNotes + "\n\n")
			}
			b.WriteString("---\n\n")
		}
	}
	return b.String()
}
