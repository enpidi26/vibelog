package database

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/enpidi26/vibelog/internal/domain"
)

func CreateProject(db *sql.DB, userID int64, in domain.ProjectInput) (domain.Project, error) {
	color := in.Color
	if color == "" {
		color = "teal"
	}
	now := utcNow()
	res, err := db.Exec(`INSERT INTO projects (user_id, name, description, color, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		userID, strings.TrimSpace(in.Name), strings.TrimSpace(in.Description), color, now, now)
	if err != nil {
		return domain.Project{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Project{}, err
	}
	return GetProject(db, userID, id)
}

func GetProject(db *sql.DB, userID, id int64) (domain.Project, error) {
	var p domain.Project
	var created, updated string
	err := db.QueryRow(`
		SELECT p.id, p.name, p.description, p.color, p.created_at, p.updated_at,
			(SELECT COUNT(*) FROM entries e WHERE e.user_id = p.user_id AND e.project_id = p.id)
		FROM projects p WHERE p.id = ? AND p.user_id = ?`, id, userID).
		Scan(&p.ID, &p.Name, &p.Description, &p.Color, &created, &updated, &p.EntryCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Project{}, domain.ErrNotFound
		}
		return domain.Project{}, err
	}
	p.UserID = userID
	p.CreatedAt = parseTime(created)
	p.UpdatedAt = parseTime(updated)
	return p, nil
}

func ListProjects(db *sql.DB, userID int64) ([]domain.Project, error) {
	rows, err := db.Query(`
		SELECT p.id, p.name, p.description, p.color, p.created_at, p.updated_at,
			(SELECT COUNT(*) FROM entries e WHERE e.user_id = p.user_id AND e.project_id = p.id)
		FROM projects p WHERE p.user_id = ? ORDER BY p.name COLLATE NOCASE ASC, p.id ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Project{}
	for rows.Next() {
		var p domain.Project
		var created, updated string
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Color, &created, &updated, &p.EntryCount); err != nil {
			return nil, err
		}
		p.UserID = userID
		p.CreatedAt = parseTime(created)
		p.UpdatedAt = parseTime(updated)
		out = append(out, p)
	}
	return out, rows.Err()
}

func UpdateProject(db *sql.DB, userID, id int64, in domain.ProjectInput) (domain.Project, error) {
	color := in.Color
	if color == "" {
		color = "teal"
	}
	res, err := db.Exec(`UPDATE projects SET name = ?, description = ?, color = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		strings.TrimSpace(in.Name), strings.TrimSpace(in.Description), color, utcNow(), id, userID)
	if err != nil {
		return domain.Project{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.Project{}, domain.ErrNotFound
	}
	return GetProject(db, userID, id)
}

// DeleteProject menghapus proyek tapi MEMPERTAHANKAN entri jurnalnya:
// project_id entri diubah jadi NULL (ON DELETE SET NULL), supaya riwayat
// prompt tidak hilang bersama proyeknya.
func DeleteProject(db *sql.DB, userID, id int64) error {
	res, err := db.Exec(`DELETE FROM projects WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
