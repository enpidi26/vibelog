package main

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	PasswordHash    string     `json:"-"`
	EmailVerified   bool       `json:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type Project struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"-"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	EntryCount  int       `json:"entry_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Entry struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"-"`
	ProjectID   *int64    `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Title       string    `json:"title"`
	PromptText  string    `json:"prompt_text"`
	Model       string    `json:"model"`
	Tool        string    `json:"tool"`
	Outcome     string    `json:"outcome"`
	Rating      *int      `json:"rating"`
	ResultNotes string    `json:"result_notes"`
	Tags        []string  `json:"tags"`
	IterationNo int       `json:"iteration_no"`
	ParentID    *int64    `json:"parent_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type EntryFilter struct {
	Query     string
	ProjectID *int64
	Model     string
	Tool      string
	Outcome   string
	Tag       string
	Limit     int
}

type CountItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Summary struct {
	Total        int         `json:"total"`
	Success      int         `json:"success"`
	Partial      int         `json:"partial"`
	Failed       int         `json:"failed"`
	InProgress   int         `json:"in_progress"`
	AverageRating *float64   `json:"average_rating"`
	RatedCount   int         `json:"rated_count"`
	ByModel      []CountItem `json:"by_model"`
	ByTool       []CountItem `json:"by_tool"`
	ByProject    []CountItem `json:"by_project"`
	TopTags      []CountItem `json:"top_tags"`
}

var ErrNotFound = errors.New("tidak ditemukan")
var ErrEmailTaken = errors.New("email already registered")
var ErrInvalidOTP = errors.New("invalid or expired OTP code")

// OpenDB membuka SQLite, mengaktifkan WAL + foreign keys, dan migrasi.
func OpenDB(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=ON&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			email_verified INTEGER NOT NULL DEFAULT 0,
			email_verified_at TEXT,
			created_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			color TEXT NOT NULL DEFAULT 'teal',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			project_id INTEGER REFERENCES projects(id) ON DELETE SET NULL,
			title TEXT NOT NULL,
			prompt_text TEXT NOT NULL,
			model TEXT NOT NULL,
			tool TEXT NOT NULL,
			outcome TEXT NOT NULL,
			rating INTEGER,
			result_notes TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '',
			iteration_no INTEGER NOT NULL DEFAULT 1,
			parent_id INTEGER REFERENCES entries(id) ON DELETE SET NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS otp_codes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			code TEXT NOT NULL,
			purpose TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			used INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_projects_user ON projects(user_id, name);
		CREATE INDEX IF NOT EXISTS idx_entries_user_created ON entries(user_id, created_at DESC, id DESC);
		CREATE INDEX IF NOT EXISTS idx_entries_user_project ON entries(user_id, project_id);
		CREATE INDEX IF NOT EXISTS idx_entries_user_outcome ON entries(user_id, outcome);
		CREATE INDEX IF NOT EXISTS idx_otp_user_purpose ON otp_codes(user_id, purpose, expires_at);
	`)

	// Migrate pre-OTP schemas: add email verification columns to users if missing.
	var hasEmailVerified int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name = 'email_verified'`).Scan(&hasEmailVerified); err != nil {
		return err
	}
	if hasEmailVerified == 0 {
		if _, err := db.Exec(`ALTER TABLE users ADD COLUMN email_verified INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
		if _, err := db.Exec(`ALTER TABLE users ADD COLUMN email_verified_at TEXT`); err != nil {
			return err
		}
	}

	return err
}

func utcNow() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t.UTC()
}

func tagsToText(tags []string) string { return strings.Join(tags, ", ") }

func textToTags(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	tags, _ := NormalizeTags(strings.Split(s, ","))
	if tags == nil {
		return []string{}
	}
	return tags
}

func nullableInt64(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	n := v.Int64
	return &n
}

func nullableInt(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	n := int(v.Int64)
	return &n
}

// ---- Users ----

func CreateUser(db *sql.DB, name, email, passwordHash string) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	res, err := db.Exec(`INSERT INTO users (name, email, password_hash, email_verified, created_at) VALUES (?, ?, ?, 0, ?)`,
		strings.TrimSpace(name), email, passwordHash, utcNow())
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return User{}, ErrEmailTaken
		}
		return User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}
	return GetUserByID(db, id)
}

func GetUserByID(db *sql.DB, id int64) (User, error) {
	var u User
	var created string
	var verified int
	var verifiedAt sql.NullString
	err := db.QueryRow(`SELECT id, name, email, password_hash, email_verified, email_verified_at, created_at FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &verified, &verifiedAt, &created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	u.EmailVerified = verified == 1
	if verifiedAt.Valid {
		t := parseTime(verifiedAt.String)
		u.EmailVerifiedAt = &t
	}
	u.CreatedAt = parseTime(created)
	return u, nil
}

func GetUserByEmail(db *sql.DB, email string) (User, error) {
	var u User
	var created string
	var verified int
	var verifiedAt sql.NullString
	err := db.QueryRow(`SELECT id, name, email, password_hash, email_verified, email_verified_at, created_at FROM users WHERE email = ?`,
		strings.ToLower(strings.TrimSpace(email))).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &verified, &verifiedAt, &created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	u.EmailVerified = verified == 1
	if verifiedAt.Valid {
		t := parseTime(verifiedAt.String)
		u.EmailVerifiedAt = &t
	}
	u.CreatedAt = parseTime(created)
	return u, nil
}

// ---- Projects ----

func CreateProject(db *sql.DB, userID int64, in ProjectInput) (Project, error) {
	color := in.Color
	if color == "" {
		color = "teal"
	}
	now := utcNow()
	res, err := db.Exec(`INSERT INTO projects (user_id, name, description, color, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		userID, strings.TrimSpace(in.Name), strings.TrimSpace(in.Description), color, now, now)
	if err != nil {
		return Project{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Project{}, err
	}
	return GetProject(db, userID, id)
}

func GetProject(db *sql.DB, userID, id int64) (Project, error) {
	var p Project
	var created, updated string
	err := db.QueryRow(`
		SELECT p.id, p.name, p.description, p.color, p.created_at, p.updated_at,
			(SELECT COUNT(*) FROM entries e WHERE e.user_id = p.user_id AND e.project_id = p.id)
		FROM projects p WHERE p.id = ? AND p.user_id = ?`, id, userID).
		Scan(&p.ID, &p.Name, &p.Description, &p.Color, &created, &updated, &p.EntryCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, ErrNotFound
		}
		return Project{}, err
	}
	p.UserID = userID
	p.CreatedAt = parseTime(created)
	p.UpdatedAt = parseTime(updated)
	return p, nil
}

func ListProjects(db *sql.DB, userID int64) ([]Project, error) {
	rows, err := db.Query(`
		SELECT p.id, p.name, p.description, p.color, p.created_at, p.updated_at,
			(SELECT COUNT(*) FROM entries e WHERE e.user_id = p.user_id AND e.project_id = p.id)
		FROM projects p WHERE p.user_id = ? ORDER BY p.name COLLATE NOCASE ASC, p.id ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Project{}
	for rows.Next() {
		var p Project
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

func UpdateProject(db *sql.DB, userID, id int64, in ProjectInput) (Project, error) {
	color := in.Color
	if color == "" {
		color = "teal"
	}
	res, err := db.Exec(`UPDATE projects SET name = ?, description = ?, color = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		strings.TrimSpace(in.Name), strings.TrimSpace(in.Description), color, utcNow(), id, userID)
	if err != nil {
		return Project{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return Project{}, ErrNotFound
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
		return ErrNotFound
	}
	return nil
}

// ---- Entries ----

func scanEntry(row interface {
	Scan(dest ...any) error
}, userID int64) (Entry, error) {
	var e Entry
	var projectID, parentID, rating sql.NullInt64
	var projectName sql.NullString
	var tagsText, created, updated string
	err := row.Scan(&e.ID, &projectID, &projectName, &e.Title, &e.PromptText, &e.Model, &e.Tool,
		&e.Outcome, &rating, &e.ResultNotes, &tagsText, &e.IterationNo, &parentID, &created, &updated)
	if err != nil {
		return Entry{}, err
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

func GetEntry(db *sql.DB, userID, id int64) (Entry, error) {
	e, err := scanEntry(db.QueryRow(entrySelect, id, userID), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Entry{}, ErrNotFound
		}
		return Entry{}, err
	}
	return e, nil
}

func CreateEntry(db *sql.DB, userID int64, in EntryInput) (Entry, error) {
	tags, _ := ParseTagsInput(in.Tags)
	now := utcNow()
	var projectID any
	if in.ProjectID != nil {
		// Proyek harus milik user yang sama; kalau bukan, tolak.
		if _, err := GetProject(db, userID, *in.ProjectID); err != nil {
			return Entry{}, ErrNotFound
		}
		projectID = *in.ProjectID
	}
	res, err := db.Exec(`INSERT INTO entries
		(user_id, project_id, title, prompt_text, model, tool, outcome, rating, result_notes, tags, iteration_no, parent_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, NULL, ?, ?)`,
		userID, projectID, strings.TrimSpace(in.Title), in.PromptText, strings.TrimSpace(in.Model),
		strings.TrimSpace(in.Tool), in.Outcome, in.Rating, in.ResultNotes, tagsToText(tags), now, now)
	if err != nil {
		return Entry{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Entry{}, err
	}
	return GetEntry(db, userID, id)
}

func UpdateEntry(db *sql.DB, userID, id int64, in EntryInput) (Entry, error) {
	tags, _ := ParseTagsInput(in.Tags)
	var projectID any
	if in.ProjectID != nil {
		if _, err := GetProject(db, userID, *in.ProjectID); err != nil {
			return Entry{}, ErrNotFound
		}
		projectID = *in.ProjectID
	}
	res, err := db.Exec(`UPDATE entries SET project_id = ?, title = ?, prompt_text = ?, model = ?, tool = ?,
		outcome = ?, rating = ?, result_notes = ?, tags = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		projectID, strings.TrimSpace(in.Title), in.PromptText, strings.TrimSpace(in.Model),
		strings.TrimSpace(in.Tool), in.Outcome, in.Rating, in.ResultNotes, tagsToText(tags), utcNow(), id, userID)
	if err != nil {
		return Entry{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return Entry{}, ErrNotFound
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
		return ErrNotFound
	}
	return nil
}

// DuplicateEntry membuat iterasi berikutnya dari entri lama: isi disalin,
// iteration_no naik 1, parent_id menunjuk entri asal, outcome kembali ke
// in_progress supaya iterasi baru jelas belum dinilai.
func DuplicateEntry(db *sql.DB, userID, id int64) (Entry, error) {
	orig, err := GetEntry(db, userID, id)
	if err != nil {
		return Entry{}, err
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
		return Entry{}, err
	}
	newID, err := res.LastInsertId()
	if err != nil {
		return Entry{}, err
	}
	return GetEntry(db, userID, newID)
}

func ListEntries(db *sql.DB, userID int64, f EntryFilter) ([]Entry, error) {
	limit := f.Limit
	if limit < 1 || limit > 200 {
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
		query += " AND (',' || REPLACE(e.tags, ', ', ',') || ',' LIKE ?)"
		args = append(args, "%,"+strings.ToLower(f.Tag)+",%")
	}
	if f.Query != "" {
		q := "%" + strings.ToLower(f.Query) + "%"
		query += " AND (LOWER(e.title) LIKE ? OR LOWER(e.prompt_text) LIKE ? OR LOWER(e.result_notes) LIKE ? OR LOWER(e.tags) LIKE ?)"
		args = append(args, q, q, q, q)
	}
	query += " ORDER BY e.created_at DESC, e.id DESC LIMIT ?"
	args = append(args, limit)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Entry{}
	for rows.Next() {
		e, err := scanEntry(rows, userID)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func EntrySummary(db *sql.DB, userID int64, projectID *int64) (Summary, error) {
	var s Summary
	s.ByModel = []CountItem{}
	s.ByTool = []CountItem{}
	s.ByProject = []CountItem{}
	s.TopTags = []CountItem{}
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
	countBy := func(col string) ([]CountItem, error) {
		rows, err := db.Query(`SELECT `+col+`, COUNT(*) FROM entries WHERE user_id = ?`+projClause+` GROUP BY `+col+` ORDER BY COUNT(*) DESC, `+col+` ASC LIMIT 10`, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		items := []CountItem{}
		for rows.Next() {
			var it CountItem
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
		var it CountItem
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
		s.TopTags = append(s.TopTags, CountItem{Name: name, Count: count})
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

// ---- OTP Management ----

// CreateOTP membuat kode OTP 6 digit untuk user dan purpose tertentu.
func CreateOTP(db *sql.DB, userID int64, code, purpose string, expiryMinutes int) error {
	expiresAt := time.Now().UTC().Add(time.Duration(expiryMinutes) * time.Minute).Format(time.RFC3339Nano)
	_, err := db.Exec(`INSERT INTO otp_codes (user_id, code, purpose, expires_at, created_at) VALUES (?, ?, ?, ?, ?)`,
		userID, code, purpose, expiresAt, utcNow())
	return err
}

// VerifyOTP memvalidasi kode OTP untuk user dan purpose tertentu.
// Mengembalikan error jika tidak valid atau sudah kadaluarsa.
func VerifyOTP(db *sql.DB, userID int64, code, purpose string) error {
	var id int64
	var expiresAt, used string
	err := db.QueryRow(`SELECT id, expires_at, used FROM otp_codes 
		WHERE user_id = ? AND code = ? AND purpose = ? 
		ORDER BY created_at DESC LIMIT 1`, userID, code, purpose).
		Scan(&id, &expiresAt, &used)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidOTP
		}
		return err
	}
	if used == "1" {
		return ErrInvalidOTP
	}
	expiry := parseTime(expiresAt)
	if time.Now().UTC().After(expiry) {
		return ErrInvalidOTP
	}
	// Tandai sebagai sudah dipakai
	_, err = db.Exec(`UPDATE otp_codes SET used = 1 WHERE id = ?`, id)
	return err
}

// MarkEmailVerified menandai email user sebagai terverifikasi.
func MarkEmailVerified(db *sql.DB, userID int64) error {
	now := utcNow()
	_, err := db.Exec(`UPDATE users SET email_verified = 1, email_verified_at = ? WHERE id = ?`, now, userID)
	return err
}

// CleanupExpiredOTPs menghapus OTP yang sudah kadaluarsa (lebih dari 24 jam).
func CleanupExpiredOTPs(db *sql.DB) error {
	cutoff := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339Nano)
	_, err := db.Exec(`DELETE FROM otp_codes WHERE expires_at < ?`, cutoff)
	return err
}
