package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (a *App) funcs() template.FuncMap {
	return template.FuncMap{
		"short": func(date string) string {
			if len(date) != 10 {
				return date
			}
			m, _ := strconv.Atoi(date[5:7])
			d, _ := strconv.Atoi(date[8:])
			return fmt.Sprintf("%d/%d", m, d)
		},
		"hm": func(min int) string {
			h, m := min/60, min%60
			switch {
			case h == 0:
				return fmt.Sprintf("%d分", m)
			case m == 0:
				return fmt.Sprintf("%d時間", h)
			default:
				return fmt.Sprintf("%d時間%d分", h, m)
			}
		},
		"add": func(x, y int) int { return x + y },
		"sub": func(x, y int) int { return x - y },
		"pct": func(n, d int) int {
			if d == 0 {
				return 0
			}
			return n * 100 / d
		},
		"field": func(no int) Field { f, _ := fieldByNo(no); return f },
	}
}

// ---- day / today ----

type TaskView struct {
	Task
	Done   bool
	IsNext bool
}

type DayStats struct{ Done, Total, TotalMin, LeftMin int }

func (s DayStats) Complete() bool { return s.Total > 0 && s.Done == s.Total }

func dayStats(d Day, done map[string]bool, custom [2]int) DayStats {
	var s DayStats
	for _, tk := range d.Tasks {
		s.Total++
		s.TotalMin += tk.Min
		if done[tk.ID] {
			s.Done++
		} else {
			s.LeftMin += tk.Min
		}
	}
	s.Done += custom[0]
	s.Total += custom[1]
	return s
}

type Chip struct {
	Day      Day
	Stats    DayStats
	IsToday  bool
	Selected bool
}

func (a *App) today(w http.ResponseWriter, r *http.Request) {
	date := a.todayStr()
	if date < days[0].Date {
		date = days[0].Date
	}
	if last := days[len(days)-1].Date; date > last {
		date = last
	}
	a.renderDay(w, date)
}

func (a *App) dayPage(w http.ResponseWriter, r *http.Request) {
	date := r.PathValue("date")
	if _, _, ok := dayByDate(date); !ok {
		http.NotFound(w, r)
		return
	}
	a.renderDay(w, date)
}

func (a *App) renderDay(w http.ResponseWriter, date string) {
	day, idx, _ := dayByDate(date)
	done, err := a.store.DoneSet()
	if err != nil {
		a.fail(w, err)
		return
	}
	counts, err := a.store.CustomCounts()
	if err != nil {
		a.fail(w, err)
		return
	}
	custom, err := a.store.CustomTasks(date)
	if err != nil {
		a.fail(w, err)
		return
	}
	mockScores, err := a.store.MockScores()
	if err != nil {
		a.fail(w, err)
		return
	}
	today := a.todayStr()
	var chips []Chip
	for _, d := range days {
		chips = append(chips, Chip{Day: d, Stats: dayStats(d, done, counts[d.Date]), IsToday: d.Date == today, Selected: d.Date == date})
	}
	var tasks []TaskView
	nextFound := false
	for _, tk := range day.Tasks {
		v := TaskView{Task: tk, Done: done[tk.ID]}
		if !v.Done && !nextFound {
			v.IsNext, nextFound = true, true
		}
		tasks = append(tasks, v)
	}
	relation := "今日"
	switch {
	case date < today:
		relation = "過去の日"
	case date > today:
		relation = "これからの日"
	}
	data := map[string]any{
		"Nav": "today", "Day": day, "Index": idx + 1, "NumDays": len(days), "Chips": chips, "Tasks": tasks,
		"Custom": custom, "Stats": dayStats(day, done, counts[date]), "Relation": relation, "Back": a.base + "/day/" + date,
	}
	if day.Mock != "" {
		if sc, ok := mockScores[day.Mock]; ok {
			data["MockScore"], data["HasMock"] = sc, true
		}
	}
	a.render(w, "day", data)
}

func (a *App) toggleTask(w http.ResponseWriter, r *http.Request) {
	if !limitForm(w, r) {
		return
	}
	id := r.PathValue("id")
	if !taskExists(id) {
		http.NotFound(w, r)
		return
	}
	if _, err := a.store.ToggleTask(id); err != nil {
		a.fail(w, err)
		return
	}
	a.back(w, r, a.base+"/")
}

func (a *App) addCustom(w http.ResponseWriter, r *http.Request) {
	if !limitForm(w, r) {
		return
	}
	date := r.PostFormValue("date")
	title := strings.TrimSpace(r.PostFormValue("title"))
	min, ok := formInt(r, "minutes", 0, 600)
	if _, _, found := dayByDate(date); !found || title == "" || utf8.RuneCountInString(title) > 100 || !ok {
		http.Error(w, "入力内容を確認してください。", http.StatusBadRequest)
		return
	}
	if err := a.store.AddCustomTask(date, title, min); err != nil {
		a.fail(w, err)
		return
	}
	a.back(w, r, a.base+"/day/"+date)
}

func (a *App) customID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	if !limitForm(w, r) {
		return 0, false
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return 0, false
	}
	return id, true
}

func (a *App) toggleCustom(w http.ResponseWriter, r *http.Request) {
	id, ok := a.customID(w, r)
	if !ok {
		return
	}
	date, err := a.store.ToggleCustomTask(id)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		a.fail(w, err)
		return
	}
	a.back(w, r, a.base+"/day/"+date)
}

func (a *App) deleteCustom(w http.ResponseWriter, r *http.Request) {
	id, ok := a.customID(w, r)
	if !ok {
		return
	}
	date, err := a.store.DeleteCustomTask(id)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		a.fail(w, err)
		return
	}
	a.back(w, r, a.base+"/day/"+date)
}

// ---- plan ----

type PlanRow struct {
	Day     Day
	Stats   DayStats
	IsToday bool
}

type MockView struct {
	ID, Label string
	Score     int
	Has       bool
}

func (a *App) mockViews() ([]MockView, error) {
	scores, err := a.store.MockScores()
	if err != nil {
		return nil, err
	}
	var out []MockView
	for _, m := range mocks {
		sc, ok := scores[m[0]]
		out = append(out, MockView{ID: m[0], Label: m[1], Score: sc, Has: ok})
	}
	return out, nil
}

func (a *App) planPage(w http.ResponseWriter, r *http.Request) {
	done, err := a.store.DoneSet()
	if err != nil {
		a.fail(w, err)
		return
	}
	counts, err := a.store.CustomCounts()
	if err != nil {
		a.fail(w, err)
		return
	}
	mv, err := a.mockViews()
	if err != nil {
		a.fail(w, err)
		return
	}
	today := a.todayStr()
	type Phase struct {
		Name string
		Rows []PlanRow
	}
	var ps []Phase
	totalMin := 0
	for _, name := range phases {
		p := Phase{Name: name}
		for _, d := range days {
			if d.Phase == name {
				st := dayStats(d, done, counts[d.Date])
				totalMin += st.TotalMin
				p.Rows = append(p.Rows, PlanRow{Day: d, Stats: st, IsToday: d.Date == today})
			}
		}
		ps = append(ps, p)
	}
	main, ins, err := a.store.Fields()
	if err != nil {
		a.fail(w, err)
		return
	}
	a.render(w, "plan", map[string]any{"Nav": "plan", "Phases": ps, "TotalHours": (totalMin + 30) / 60,
		"Mocks": mv, "Main": codesToFields(main), "Insurance": codesToFields(ins), "Back": a.base + "/plan"})
}

func (a *App) saveMock(w http.ResponseWriter, r *http.Request) {
	if !limitForm(w, r) {
		return
	}
	id := r.PathValue("id")
	if !slices.ContainsFunc(mocks, func(m [2]string) bool { return m[0] == id }) {
		http.NotFound(w, r)
		return
	}
	raw := strings.TrimSpace(r.PostFormValue("score"))
	var score *int
	if raw != "" {
		n, ok := formInt(r, "score", 0, 100)
		if !ok {
			http.Error(w, "得点は0〜100で入力してください。", http.StatusBadRequest)
			return
		}
		score = &n
	}
	if err := a.store.SetMockScore(id, score); err != nil {
		a.fail(w, err)
		return
	}
	a.back(w, r, a.base+"/plan")
}

// ---- log ----

type FieldStat struct {
	Field                            Field
	Count, Correct, Total, Min, MinN int
	Role                             string
}

func (f FieldStat) Pct() int {
	if f.Total == 0 {
		return 0
	}
	return f.Correct * 100 / f.Total
}

func (f FieldStat) AvgMin() int {
	if f.MinN == 0 {
		return 0
	}
	return f.Min / f.MinN
}

func (a *App) fieldStats(sessions []Session) ([]FieldStat, error) {
	main, ins, err := a.store.Fields()
	if err != nil {
		return nil, err
	}
	out := make([]FieldStat, len(fields))
	for i, f := range fields {
		out[i].Field = f
		switch {
		case f.No == 1:
			out[i].Role = "必須"
		case slices.Contains(main, f.Code):
			out[i].Role = "メイン"
		case slices.Contains(ins, f.Code):
			out[i].Role = "保険"
		}
	}
	for _, s := range sessions {
		if s.QNo < 1 || s.QNo > len(out) {
			continue
		}
		st := &out[s.QNo-1]
		st.Count++
		st.Correct += s.Correct
		st.Total += s.Total
		if s.Minutes.Valid {
			st.Min += int(s.Minutes.Int64)
			st.MinN++
		}
	}
	return out, nil
}

func (a *App) logPage(w http.ResponseWriter, r *http.Request) {
	sessions, err := a.store.Sessions()
	if err != nil {
		a.fail(w, err)
		return
	}
	stats, err := a.fieldStats(sessions)
	if err != nil {
		a.fail(w, err)
		return
	}
	var tried []FieldStat
	for _, s := range stats {
		if s.Count > 0 {
			tried = append(tried, s)
		}
	}
	a.render(w, "log", map[string]any{"Nav": "log", "Sessions": sessions, "Stats": tried, "Exams": exams, "Fields": fields,
		"Today": a.todayStr(), "Back": a.base + "/log"})
}

func (a *App) addSession(w http.ResponseWriter, r *http.Request) {
	if !limitForm(w, r) {
		return
	}
	var x Session
	x.Date = r.PostFormValue("date")
	x.Exam = r.PostFormValue("exam")
	x.TimeRange = strings.TrimSpace(r.PostFormValue("time_range"))
	x.Memo = strings.TrimSpace(r.PostFormValue("memo"))
	var ok1, ok2, ok3, ok4 bool
	x.QNo, ok1 = formInt(r, "qno", 1, 11)
	x.Correct, ok2 = formInt(r, "correct", 0, 50)
	x.Total, ok3 = formInt(r, "total", 1, 50)
	x.Partial, ok4 = 0, true
	if strings.TrimSpace(r.PostFormValue("partial")) != "" {
		x.Partial, ok4 = formInt(r, "partial", 0, 50)
	}
	if m := strings.TrimSpace(r.PostFormValue("minutes")); m != "" {
		n, ok := formInt(r, "minutes", 0, 300)
		if !ok {
			ok4 = false
		}
		x.Minutes = sql.NullInt64{Int64: int64(n), Valid: true}
	}
	_, examOK := examByKey(x.Exam)
	if !validDate(x.Date) || !examOK || !ok1 || !ok2 || !ok3 || !ok4 || x.Correct+x.Partial > x.Total ||
		utf8.RuneCountInString(x.TimeRange) > 30 || utf8.RuneCountInString(x.Memo) > 2000 {
		http.Error(w, "入力内容を確認してください（正解＋部分点は問題数以下）。", http.StatusBadRequest)
		return
	}
	if err := a.store.AddSession(x); err != nil {
		a.fail(w, err)
		return
	}
	a.back(w, r, a.base+"/log")
}

func (a *App) deleteSession(w http.ResponseWriter, r *http.Request) {
	if !limitForm(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := a.store.DeleteSession(id); err != nil {
		a.fail(w, err)
		return
	}
	a.back(w, r, a.base+"/log")
}

// ---- fields ----

func codesToFields(codes []string) []Field {
	var out []Field
	for _, c := range codes {
		if f, ok := fieldByCode(c); ok {
			out = append(out, f)
		}
	}
	return out
}

func (a *App) fieldsPage(w http.ResponseWriter, r *http.Request) {
	sessions, err := a.store.Sessions()
	if err != nil {
		a.fail(w, err)
		return
	}
	stats, err := a.fieldStats(sessions)
	if err != nil {
		a.fail(w, err)
		return
	}
	main, ins, err := a.store.Fields()
	if err != nil {
		a.fail(w, err)
		return
	}
	a.render(w, "fields", map[string]any{"Nav": "fields", "Stats": stats[1:], "MainN": len(main), "InsN": len(ins),
		"Saved": r.URL.Query().Get("saved") != "", "Back": a.base + "/fields"})
}

func (a *App) saveFields(w http.ResponseWriter, r *http.Request) {
	if !limitForm(w, r) {
		return
	}
	var main, ins []string
	for _, f := range fields[1:] {
		switch r.PostFormValue("role_" + f.Code) {
		case "main":
			main = append(main, f.Code)
		case "insurance":
			ins = append(ins, f.Code)
		}
	}
	if len(main) > maxMain {
		http.Error(w, "メインは5分野までです。", http.StatusBadRequest)
		return
	}
	if err := a.store.SetFields(main, ins); err != nil {
		a.fail(w, err)
		return
	}
	http.Redirect(w, r, a.base+"/fields?saved=1", http.StatusSeeOther)
}

// ---- static pages ----

func (a *App) phrasesPage(w http.ResponseWriter, r *http.Request) {
	a.render(w, "phrases", map[string]any{"Nav": "phrases", "Groups": phraseGroups})
}

func (a *App) examPage(w http.ResponseWriter, r *http.Request) {
	main, ins, err := a.store.Fields()
	if err != nil {
		a.fail(w, err)
		return
	}
	a.render(w, "exam", map[string]any{"Nav": "exam", "Main": codesToFields(main), "Insurance": codesToFields(ins)})
}
