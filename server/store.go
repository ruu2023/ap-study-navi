package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }

const schema = `
CREATE TABLE IF NOT EXISTS task_done (
  task_id TEXT PRIMARY KEY,
  done_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS custom_tasks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date TEXT NOT NULL,
  title TEXT NOT NULL,
  minutes INTEGER NOT NULL DEFAULT 0,
  done INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS sessions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date TEXT NOT NULL,
  time_range TEXT NOT NULL DEFAULT '',
  exam TEXT NOT NULL,
  qno INTEGER NOT NULL,
  minutes INTEGER,
  correct INTEGER NOT NULL,
  partial INTEGER NOT NULL DEFAULT 0,
  total INTEGER NOT NULL,
  memo TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS mock_scores (
  mock_id TEXT PRIMARY KEY,
  score INTEGER NOT NULL
);
`

func openStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // one writer; the app is single-user
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	s := &Store{db: db}
	if err := s.seedOnce(); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}
	return s, nil
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

// seedOnce imports the progress and practice log recorded before this app
// existed (GitHub Pages app + notes/study-log.md). It runs only on a new DB.
func (s *Store) seedOnce() error {
	var seeded string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key='seeded'`).Scan(&seeded)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	done := []string{
		"d19-tutorial", "d19-q1", "d19-q9", "d19-q11",
		"d20-q10", "d20-tech1", "d20-tech2", "d20-pick",
		"d21-a1", "d21-a9", "d21-a10", "d21-a11", "d21-b1", "d21-b9",
	}
	for _, id := range done {
		if _, err := tx.Exec(`INSERT OR IGNORE INTO task_done(task_id, done_at) VALUES(?, ?)`, id, now()); err != nil {
			return err
		}
	}
	type seed struct {
		date, tr, exam string
		qno            int
		min            any
		c, p, tot      int
		memo           string
	}
	seeds := []seed{
		{"2026-11-19", "21:25〜21:46", "r5a", 10, 21, 1, 0, 6, "社外は契約書・社内は合意文書。理由は「SLAを守るため」につなげ、結果まで書く。"},
		{"2026-11-20", "", "r5a", 5, 18, 3, 0, 9, "OP25B、MXレコード、ネットワークアドレス（末尾0）を落とした。社内から参照する設定は内部IP。"},
		{"2026-11-20", "", "r5a", 6, 9, 10, 0, 10, "全問正解。"},
		{"2026-11-20", "", "r6h", 1, 7, 8, 0, 8, "全問正解。解答例の「・」は別解のことがある。"},
		{"2026-11-21", "〜9:33", "r6h", 10, nil, 1, 2, 7, "展開管理＝適用タイミングの制御。終了条件は「到達した状態」で書く。"},
		{"2026-11-21", "10:42〜10:50", "r6h", 11, 8, 1, 0, 8, "監査は抜き出し型。8分で推測で埋めた。25分使って本文から拾う。"},
		{"2026-11-21", "〜11:45", "r5h", 1, 10, 5, 0, 10, "表どうしの対応づけミス。バックアップはオフライン保管。「全て選べ」は目的で絞る。"},
		{"2026-11-21", "", "r6h", 9, 13, 3, 0, 8, "PoCの理由＝不確かな要素が多い。スコープ変更の影響＝コスト・スケジュール。"},
		{"2026-11-21", "", "r5h", 9, 9, 1, 0, 9, "本文の言葉をそのまま使えば取れた問題が多い。10分弱で終えている。"},
	}
	for _, sd := range seeds {
		if _, err := tx.Exec(`INSERT INTO sessions(date, time_range, exam, qno, minutes, correct, partial, total, memo, created_at) VALUES(?,?,?,?,?,?,?,?,?,?)`,
			sd.date, sd.tr, sd.exam, sd.qno, sd.min, sd.c, sd.p, sd.tot, sd.memo, now()); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`INSERT INTO settings(key, value) VALUES('fields_main', ?), ('fields_insurance', ?), ('seeded', ?)`,
		strings.Join(defaultMain, ","), strings.Join(defaultInsurance, ","), now()); err != nil {
		return err
	}
	return tx.Commit()
}

// ---- tasks ----

func (s *Store) DoneSet() (map[string]bool, error) {
	rows, err := s.db.Query(`SELECT task_id FROM task_done`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		m[id] = true
	}
	return m, rows.Err()
}

func (s *Store) ToggleTask(id string) (bool, error) {
	res, err := s.db.Exec(`DELETE FROM task_done WHERE task_id=?`, id)
	if err != nil {
		return false, err
	}
	if n, _ := res.RowsAffected(); n > 0 {
		return false, nil
	}
	_, err = s.db.Exec(`INSERT INTO task_done(task_id, done_at) VALUES(?, ?)`, id, now())
	return err == nil, err
}

type CustomTask struct {
	ID      int64
	Date    string
	Title   string
	Minutes int
	Done    bool
}

func (s *Store) CustomTasks(date string) ([]CustomTask, error) {
	rows, err := s.db.Query(`SELECT id, date, title, minutes, done FROM custom_tasks WHERE date=? ORDER BY id`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CustomTask
	for rows.Next() {
		var c CustomTask
		if err := rows.Scan(&c.ID, &c.Date, &c.Title, &c.Minutes, &c.Done); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) CustomCounts() (map[string][2]int, error) {
	rows, err := s.db.Query(`SELECT date, SUM(done), COUNT(*) FROM custom_tasks GROUP BY date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[string][2]int{}
	for rows.Next() {
		var d string
		var done, total int
		if err := rows.Scan(&d, &done, &total); err != nil {
			return nil, err
		}
		m[d] = [2]int{done, total}
	}
	return m, rows.Err()
}

func (s *Store) AddCustomTask(date, title string, minutes int) error {
	_, err := s.db.Exec(`INSERT INTO custom_tasks(date, title, minutes, created_at) VALUES(?,?,?,?)`, date, title, minutes, now())
	return err
}

func (s *Store) ToggleCustomTask(id int64) (string, error) {
	var date string
	err := s.db.QueryRow(`UPDATE custom_tasks SET done = 1 - done WHERE id=? RETURNING date`, id).Scan(&date)
	return date, err
}

func (s *Store) DeleteCustomTask(id int64) (string, error) {
	var date string
	err := s.db.QueryRow(`DELETE FROM custom_tasks WHERE id=? RETURNING date`, id).Scan(&date)
	return date, err
}

// ---- sessions ----

type Session struct {
	ID        int64
	Date      string
	TimeRange string
	Exam      string
	QNo       int
	Minutes   sql.NullInt64
	Correct   int
	Partial   int
	Total     int
	Memo      string
}

func (s Session) Field() Field { f, _ := fieldByNo(s.QNo); return f }
func (s Session) ExamLabel() string {
	if e, ok := examByKey(s.Exam); ok {
		return e.Label
	}
	return s.Exam
}
func (s Session) Pct() int {
	if s.Total == 0 {
		return 0
	}
	return s.Correct * 100 / s.Total
}

func (s *Store) Sessions() ([]Session, error) {
	rows, err := s.db.Query(`SELECT id, date, time_range, exam, qno, minutes, correct, partial, total, memo FROM sessions ORDER BY date DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Session
	for rows.Next() {
		var x Session
		if err := rows.Scan(&x.ID, &x.Date, &x.TimeRange, &x.Exam, &x.QNo, &x.Minutes, &x.Correct, &x.Partial, &x.Total, &x.Memo); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *Store) AddSession(x Session) error {
	_, err := s.db.Exec(`INSERT INTO sessions(date, time_range, exam, qno, minutes, correct, partial, total, memo, created_at) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		x.Date, x.TimeRange, x.Exam, x.QNo, x.Minutes, x.Correct, x.Partial, x.Total, x.Memo, now())
	return err
}

func (s *Store) DeleteSession(id int64) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE id=?`, id)
	return err
}

// ---- settings ----

func (s *Store) Fields() (main, insurance []string, err error) {
	get := func(k string) ([]string, error) {
		var v string
		err := s.db.QueryRow(`SELECT value FROM settings WHERE key=?`, k).Scan(&v)
		if err == sql.ErrNoRows || v == "" {
			return nil, nil
		}
		return strings.Split(v, ","), err
	}
	if main, err = get("fields_main"); err != nil {
		return
	}
	insurance, err = get("fields_insurance")
	return
}

func (s *Store) SetFields(main, insurance []string) error {
	_, err := s.db.Exec(`INSERT INTO settings(key, value) VALUES('fields_main', ?), ('fields_insurance', ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value`, strings.Join(main, ","), strings.Join(insurance, ","))
	return err
}

func (s *Store) MockScores() (map[string]int, error) {
	rows, err := s.db.Query(`SELECT mock_id, score FROM mock_scores`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[string]int{}
	for rows.Next() {
		var id string
		var sc int
		if err := rows.Scan(&id, &sc); err != nil {
			return nil, err
		}
		m[id] = sc
	}
	return m, rows.Err()
}

func (s *Store) SetMockScore(id string, score *int) error {
	if score == nil {
		_, err := s.db.Exec(`DELETE FROM mock_scores WHERE mock_id=?`, id)
		return err
	}
	_, err := s.db.Exec(`INSERT INTO mock_scores(mock_id, score) VALUES(?, ?) ON CONFLICT(mock_id) DO UPDATE SET score=excluded.score`, id, *score)
	return err
}
