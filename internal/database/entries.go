package database

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/enpidi26/vibelog/internal/domain"
)

func scanEntry(row interface {
	Scan(dest ...any) error
}, userID int64) (domain.Entry, error) {
	var e domain.Entry
	var projectID, parentID, rating sql.NullInt64
	var projectName sql.NullString
	var tagsText, created, updated string
	err := row.Scan(&e.ID, &projectID, &projectName, &e.Title, &e.PromptText, &e.Model, &e.Tool,
		&e.Outcome, &rating, &e.ResultNotes, &tagsText, &e.IterationNo, &parentID, &created, &updated)
	if err != nil {
		return domain.Entry{}, err
	}
	e.UserID = userID
	e.ProjectID = nullableInt64(projectID)
	e.ParentID = nullableInt64(parentID)
	e.Rating = nullableInt(rating)
	e.ProjectName = projectName.String
	e.Tags = textToTags(tagsText)
	e.CreatedAt = parseTime(created)
	e.UpdatedAt = parseTime(updated)
	return e, nil
}

const entrySelect = `
	SELECT e.id, e.project_id, p.name, e.title, e.prompt_text, e.model, e.tool,
		e.outcome, e.rating, e.result_notes, e.tags, e.iteration_no, e.parent_id,
		e.created_at, e.updated_at
	FROM entries e LEFT JOIN projects p ON p.id = e.project_id AND p.user_id = e.user_id
	WHERE e.id = ? AND e.user_id = ?`

func GetEntry(db *sql.DB, userID, id int64) (domain.Entry, error) {
	e, err := scanEntry(db.QueryRow(entrySelect, id, userID), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Entry{}, domain.ErrNotFound
		}
		return domain.Entry{}, err
	}
	return e, nil
}

func CreateEntry(db *sql.DB, userID int64, in domain.EntryInput) (domain.Entry, error) {
	tags, _ := domain.ParseTagsInput(in.Tags)
	now := utcNow()
	var projectID any
	if in.ProjectID != nil {
		// Proyek harus milik user yang sama; kalau bukan, tolak.
		if _, err := GetProject(db, userID, *in.ProjectID); err != nil {
			return domain.Entry{}, domain.ErrNotFound
		}
		projectID = *in.ProjectID
	}
	res, err := db.Exec(`INSERT INTO entries
		(user_id, project_id, title, prompt_text, model, tool, outcome, rating, result_notes, tags, iteration_no, parent_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, NULL, ?, ?)`,
		userID, projectID, strings.TrimSpace(in.Title), in.PromptText, strings.TrimSpace(in.Model),
		strings.TrimSpace(in.Tool), in.Outcome, in.Rating, in.ResultNotes, tagsToText(tags), now, now)
	if err != nil {
		return domain.Entry{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Entry{}, err
	}
	return GetEntry(db, userID, id)
}

func UpdateEntry(db *sql.DB, userID, id int64, in domain.EntryInput) (domain.Entry, error) {
	tags, _ := domain.ParseTagsInput(in.Tags)
	var projectID any
	if in.ProjectID != nil {
		if _, err := GetProject(db, userID, *in.ProjectID); err != nil {
			return domain.Entry{}, domain.ErrNotFound
		}
		projectID = *in.ProjectID
	}
	res, err := db.Exec(`UPDATE entries SET project_id = ?, title = ?, prompt_text = ?, model = ?, tool = ?,
		outcome = ?, rating = ?, result_notes = ?, tags = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		projectID, strings.TrimSpace(in.Title), in.PromptText, strings.TrimSpace(in.Model),
		strings.TrimSpace(in.Tool), in.Outcome, in.Rating, in.ResultNotes, tagsToText(tags), utcNow(), id, userID)
	if err != nil {
		return domain.Entry{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.Entry{}, domain.ErrNotFound
	}
	return GetEntry(db, userID, id)
}

func DeleteEntry(db *sql.DB, userID, id int64) error {
	res, err := db.Exec(`DELETE FROM entries WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// DuplicateEntry membuat iterasi berikutnya dari entri lama: isi disalin,
// iteration_no naik 1, parent_id menunjuk entri asal, outcome kembali ke
// in_progress supaya iterasi baru jelas belum dinilai.
func DuplicateEntry(db *sql.DB, userID, id int64) (domain.Entry, error) {
	orig, err := GetEntry(db, userID, id)
	if err != nil {
		return domain.Entry{}, err
	}
	now := utcNow()
	var projectID any
	if orig.ProjectID != nil {
		projectID = *orig.ProjectID
	}
	res, err := db.Exec(`INSERT INTO entries
		(user_id, project_id, title, prompt_text, model, tool, outcome, rating, result_notes, tags, iteration_no, parent_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'in_progress', NULL, ?, ?, ?, ?, ?, ?)`,
		userID, projectID, orig.Title, orig.PromptText, orig.Model, orig.Tool,
		orig.ResultNotes, tagsToText(orig.Tags), orig.IterationNo+1, orig.ID, now, now)
	if err != nil {
		return domain.Entry{}, err
	}
	newID, err := res.LastInsertId()
	if err != nil {
		return domain.Entry{}, err
	}
	return GetEntry(db, userID, newID)
}

func ListEntries(db *sql.DB, userID int64, f domain.EntryFilter) ([]domain.Entry, error) {
	limit := f.Limit
	if !f.All && (limit < 1 || limit > 200) {
		limit = 100
	}
	query := `
		SELECT e.id, e.project_id, p.name, e.title, e.prompt_text, e.model, e.tool,
			e.outcome, e.rating, e.result_notes, e.tags, e.iteration_no, e.parent_id,
			e.created_at, e.updated_at
		FROM entries e LEFT JOIN projects p ON p.id = e.project_id AND p.user_id = e.user_id
		WHERE e.user_id = ?`
	args := []any{userID}
	if f.ProjectID != nil {
		if *f.ProjectID == 0 {
			// project_id=0 berarti entri tanpa proyek (project_id IS NULL).
			query += " AND e.project_id IS NULL"
		} else {
			query += " AND e.project_id = ?"
			args = append(args, *f.ProjectID)
		}
	}
	if f.Model != "" {
		query += " AND e.model = ?"
		args = append(args, f.Model)
	}
	if f.Tool != "" {
		query += " AND e.tool = ?"
		args = append(args, f.Tool)
	}
	if f.Outcome != "" {
		query += " AND e.outcome = ?"
		args = append(args, f.Outcome)
	}
	if f.Tag != "" {
		// Tag disimpan ", " separated dan dinormalisasi huruf kecil.
		query += ` AND (',' || REPLACE(e.tags, ', ', ',') || ',') LIKE ? ESCAPE '\'`
		args = append(args, "%,"+escapeLike(strings.ToLower(f.Tag))+",%")
	}
	if f.Query != "" {
		q := likePattern(f.Query)
		query += ` AND (LOWER(e.title) LIKE ? ESCAPE '\' OR LOWER(e.prompt_text) LIKE ? ESCAPE '\' OR LOWER(e.result_notes) LIKE ? ESCAPE '\' OR LOWER(e.tags) LIKE ? ESCAPE '\')`
		args = append(args, q, q, q, q)
	}
	query += " ORDER BY e.created_at DESC, e.id DESC"
	if !f.All {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Entry{}
	for rows.Next() {
		e, err := scanEntry(rows, userID)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
