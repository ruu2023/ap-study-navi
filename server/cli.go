package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
)

// runCLI provides admin commands that work on the same DB as the server,
// e.g. `docker exec ap-study /ap-study session-add -exam r5h -q 10 ...`.
func runCLI(args []string) int {
	store, err := openStore(cfg("DB_PATH", "data/ap-study.db"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "open db:", err)
		return 1
	}
	defer store.db.Close()

	switch args[0] {
	case "session-add":
		fs := flag.NewFlagSet("session-add", flag.ContinueOnError)
		var x Session
		var min int
		fs.StringVar(&x.Date, "date", "", "YYYY-MM-DD")
		fs.StringVar(&x.TimeRange, "time", "", "e.g. 10:42〜10:50")
		fs.StringVar(&x.Exam, "exam", "", "r4h..r7a")
		fs.IntVar(&x.QNo, "q", 0, "question number 1-11")
		fs.IntVar(&min, "min", -1, "minutes (omit if unknown)")
		fs.IntVar(&x.Correct, "correct", -1, "correct items")
		fs.IntVar(&x.Partial, "partial", 0, "partial (△) items")
		fs.IntVar(&x.Total, "total", 0, "total items")
		fs.StringVar(&x.Memo, "memo", "", "memo")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		_, examOK := examByKey(x.Exam)
		if !validDate(x.Date) || !examOK || x.QNo < 1 || x.QNo > 11 || x.Correct < 0 || x.Total < 1 || x.Partial < 0 || x.Correct+x.Partial > x.Total {
			fmt.Fprintln(os.Stderr, "invalid arguments")
			return 2
		}
		if min >= 0 {
			x.Minutes = sql.NullInt64{Int64: int64(min), Valid: true}
		}
		if err := store.AddSession(x); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Println("added")
	case "sessions":
		list, err := store.Sessions()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		for _, s := range list {
			fmt.Printf("%d\t%s\t%s\t%s %s\t%d/%d\n", s.ID, s.Date, s.TimeRange, s.ExamLabel(), s.Field().Label(), s.Correct, s.Total)
		}
	case "task-done":
		fs := flag.NewFlagSet("task-done", flag.ContinueOnError)
		undo := fs.Bool("undo", false, "mark as not done instead")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		done, err := store.DoneSet()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		for _, id := range fs.Args() {
			if !taskExists(id) {
				fmt.Fprintln(os.Stderr, "unknown task:", id)
				return 2
			}
			if done[id] == !*undo {
				continue // already in the requested state
			}
			if _, err := store.ToggleTask(id); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
		}
		fmt.Println("ok")
	case "fields-set":
		fs := flag.NewFlagSet("fields-set", flag.ContinueOnError)
		main := fs.String("main", "", "comma-separated codes, e.g. st,pm,nw,au,db")
		ins := fs.String("insurance", "", "comma-separated codes")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		split := func(v string) ([]string, bool) {
			var out []string
			for _, c := range strings.Split(v, ",") {
				if c = strings.TrimSpace(c); c == "" {
					continue
				}
				if f, ok := fieldByCode(c); !ok || f.No == 1 {
					return nil, false
				}
				out = append(out, c)
			}
			return out, true
		}
		m, ok1 := split(*main)
		i, ok2 := split(*ins)
		if !ok1 || !ok2 || len(m) > maxMain {
			fmt.Fprintln(os.Stderr, "invalid field codes")
			return 2
		}
		if err := store.SetFields(m, i); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Println("saved")
	default:
		fmt.Fprintln(os.Stderr, "usage: ap-study [session-add|sessions|task-done|fields-set] ...")
		return 2
	}
	return 0
}
