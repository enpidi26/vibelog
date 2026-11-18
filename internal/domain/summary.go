package domain

// CountItem adalah satu baris agregat (model, tool, proyek, atau tag).
type CountItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Summary merangkum entri user, mengikuti filter proyek bila ada.
type Summary struct {
	Total         int         `json:"total"`
	Success       int         `json:"success"`
	Partial       int         `json:"partial"`
	Failed        int         `json:"failed"`
	InProgress    int         `json:"in_progress"`
	AverageRating *float64    `json:"average_rating"`
	RatedCount    int         `json:"rated_count"`
	ByModel       []CountItem `json:"by_model"`
	ByTool        []CountItem `json:"by_tool"`
	ByProject     []CountItem `json:"by_project"`
	TopTags       []CountItem `json:"top_tags"`
}
