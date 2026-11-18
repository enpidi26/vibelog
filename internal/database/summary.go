package database

import (
	"database/sql"

	"github.com/enpidi26/vibelog/internal/domain"
)

func EntrySummary(db *sql.DB, userID int64, projectID *int64) (domain.Summary, error) {
	var s domain.Summary
	s.ByModel = []domain.CountItem{}
	s.ByTool = []domain.CountItem{}
	s.ByProject = []domain.CountItem{}
	s.TopTags = []domain.CountItem{}
	projClause := ""
	args := []any{userID}
	if projectID != nil {
		if *projectID == 0 {
			projClause = " AND project_id IS NULL"
		} else {
			projClause = " AND project_id = ?"
			args = append(args, *projectID)
		}
	}
	row := db.QueryRow(`SELECT COUNT(*),
		COALESCE(SUM(CASE WHEN outcome='success' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN outcome='partial' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN outcome='failed' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN outcome='in_progress' THEN 1 ELSE 0 END), 0),
		AVG(rating), COUNT(rating)
		FROM entries WHERE user_id = ?`+projClause, args...)
	var avg sql.NullFloat64
	if err := row.Scan(&s.Total, &s.Success, &s.Partial, &s.Failed, &s.InProgress, &avg, &s.RatedCount); err != nil {
		return s, err
	}
	if avg.Valid {
		v := avg.Float64
		s.AverageRating = &v
	}
	countBy := func(col string) ([]domain.CountItem, error) {
		rows, err := db.Query(`SELECT `+col+`, COUNT(*) FROM entries WHERE user_id = ?`+projClause+` GROUP BY `+col+` ORDER BY COUNT(*) DESC, `+col+` ASC LIMIT 10`, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		items := []domain.CountItem{}
		for rows.Next() {
			var it domain.CountItem
			if err := rows.Scan(&it.Name, &it.Count); err != nil {
				return nil, err
			}
			items = append(items, it)
		}
		return items, rows.Err()
	}
	var err error
	if s.ByModel, err = countBy("model"); err != nil {
		return s, err
	}
	if s.ByTool, err = countBy("tool"); err != nil {
		return s, err
	}
	// By project: nama proyek, entri tanpa proyek digabung sebagai "No project".
	rows, err := db.Query(`SELECT COALESCE(p.name, 'No project'), COUNT(*)
		FROM entries e LEFT JOIN projects p ON p.id = e.project_id AND p.user_id = e.user_id
		WHERE e.user_id = ?`+projClause+` GROUP BY COALESCE(p.name, 'No project') ORDER BY COUNT(*) DESC LIMIT 10`, args...)
	if err != nil {
		return s, err
	}
	for rows.Next() {
		var it domain.CountItem
		if err := rows.Scan(&it.Name, &it.Count); err != nil {
			rows.Close()
			return s, err
		}
		s.ByProject = append(s.ByProject, it)
	}
	rows.Close()
	// Top tags dihitung di Go karena tag disimpan sebagai teks gabungan.
	tagRows, err := db.Query(`SELECT tags FROM entries WHERE user_id = ?`+projClause+` AND tags <> ''`, args...)
	if err != nil {
		return s, err
	}
	counts := map[string]int{}
	for tagRows.Next() {
		var t string
		if err := tagRows.Scan(&t); err != nil {
			tagRows.Close()
			return s, err
		}
		for _, tag := range textToTags(t) {
			counts[tag]++
		}
	}
	tagRows.Close()
	for name, count := range counts {
		s.TopTags = append(s.TopTags, domain.CountItem{Name: name, Count: count})
	}
	// Urutkan: jumlah turun, lalu nama naik. Insertion sort cukup (tag sedikit).
	for i := 1; i < len(s.TopTags); i++ {
		for j := i; j > 0 && (s.TopTags[j].Count > s.TopTags[j-1].Count || (s.TopTags[j].Count == s.TopTags[j-1].Count && s.TopTags[j].Name < s.TopTags[j-1].Name)); j-- {
			s.TopTags[j], s.TopTags[j-1] = s.TopTags[j-1], s.TopTags[j]
		}
	}
	if len(s.TopTags) > 10 {
		s.TopTags = s.TopTags[:10]
	}
	return s, nil
}
