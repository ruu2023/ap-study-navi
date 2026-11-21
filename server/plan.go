package main

import "fmt"

// Exam is one past AP exam session (IPA PDF links are built from Path).
type Exam struct {
	Key   string
	Label string
	Path  string
}

const ipaBase = "https://www.ipa.go.jp/shiken/mondai-kaiotu/"

var exams = []Exam{
	{"r7a", "令和7年度秋期", "nl10bi0000009lh8-att/2025r07a"},
	{"r7h", "令和7年度春期", "nl10bi0000009lh8-att/2025r07h"},
	{"r6a", "令和6年度秋期", "m42obm000000afqx-att/2024r06a"},
	{"r6h", "令和6年度春期", "m42obm000000afqx-att/2024r06h"},
	{"r5a", "令和5年度秋期", "ps6vr70000010d6y-att/2023r05a"},
	{"r5h", "令和5年度春期", "ps6vr70000010d6y-att/2023r05h"},
	{"r4a", "令和4年度秋期", "gmcbt80000008smf-att/2022r04a"},
	{"r4h", "令和4年度春期", "gmcbt80000009sgk-att/2022r04h"},
}

func examByKey(key string) (Exam, bool) {
	for _, e := range exams {
		if e.Key == key {
			return e, true
		}
	}
	return Exam{}, false
}

func (e Exam) Links() [][2]string {
	if e.Key == "" {
		return nil
	}
	base := ipaBase + e.Path
	return [][2]string{{base + "_ap_pm_qs.pdf", "問題"}, {base + "_ap_pm_ans.pdf", "解答例"}, {base + "_ap_pm_cmnt.pdf", "採点講評"}}
}

// Field is one question slot of 科目B (問1〜問11).
type Field struct {
	No   int
	Code string
	Name string
}

var fields = []Field{
	{1, "sec", "情報セキュリティ"},
	{2, "st", "経営戦略"},
	{3, "pg", "プログラミング"},
	{4, "sa", "システムアーキテクチャ"},
	{5, "nw", "ネットワーク"},
	{6, "db", "データベース"},
	{7, "es", "組込みシステム開発"},
	{8, "dev", "情報システム開発"},
	{9, "pm", "プロジェクトマネジメント"},
	{10, "sm", "サービスマネジメント"},
	{11, "au", "システム監査"},
}

func fieldByNo(no int) (Field, bool) {
	if no >= 1 && no <= len(fields) {
		return fields[no-1], true
	}
	return Field{}, false
}

func fieldByCode(code string) (Field, bool) {
	for _, f := range fields {
		if f.Code == code {
			return f, true
		}
	}
	return Field{}, false
}

func (f Field) Label() string { return fmt.Sprintf("問%d %s", f.No, f.Name) }

// Task is one planned item of a day.
type Task struct {
	ID     string
	Title  string
	Min    int
	Detail string
	Exam   string // exam key for PDF links (optional)
	Links  [][2]string
}

func (t Task) ExamLinks() [][2]string {
	if e, ok := examByKey(t.Exam); ok {
		return e.Links()
	}
	return nil
}

type Day struct {
	Date, Dow, Phase, Title, Goal, Finish, Point string
	Mock                                         string
	ExamDay                                      bool
	Tasks                                        []Task
}

const solve = "25分を使い切って解き、20分で復習する。答えの根拠になる本文の行に印を付けてから書く。"

func q(id, exam string, no int) Task {
	e, _ := examByKey(exam)
	f, _ := fieldByNo(no)
	return Task{ID: id, Title: e.Label + " " + f.Label(), Min: 45, Detail: solve, Exam: exam}
}

func t(id, title string, min int, detail string) Task {
	return Task{ID: id, Title: title, Min: min, Detail: detail}
}

func te(id, title string, min int, detail, exam string) Task {
	return Task{ID: id, Title: title, Min: min, Detail: detail, Exam: exam}
}

const (
	tutorialURL = "https://ijuken.com/engine/api/public/demo_exam_start/02f50984f6a37331f9af54a7b401e41f/"
	guideURL    = "https://cbt-s.com/examinee/files/examination/17907371471.pdf"
)

// The task IDs of 11/19〜11/21 match the old GitHub Pages app so the seeded
// progress lines up.
var days = []Day{
	{
		Date: "2026-11-19", Dow: "木", Phase: "診断", Title: "操作に慣れて、現在地を知る",
		Goal: "CBTの画面を一度触り、主力分野の実力を測ります。", Finish: "操作に迷わない。得点を記録できた。",
		Point: "点が低くても問題ありません。今日は弱点を見つける日です。",
		Tasks: []Task{
			{ID: "d19-tutorial", Title: "公式チュートリアルで操作確認", Min: 15, Detail: "問題選択、「選択」ボタン、メモ、マーカー、電卓を触る。PCで開く。",
				Links: [][2]string{{tutorialURL, "チュートリアル"}, {guideURL, "操作説明PDF"}}},
			q("d19-q1", "r5a", 1), q("d19-q9", "r5a", 9), q("d19-q11", "r5a", 11),
		},
	},
	{
		Date: "2026-11-20", Dow: "金", Phase: "診断", Title: "解く分野の候補を試す",
		Goal: "残りの候補を試し、得意・不得意を確かめます。", Finish: "技術系の候補を2問以上解いた。",
		Point: "迷ったら、得点の高さより「25分で解き切れたか」で選びます。",
		Tasks: []Task{
			q("d20-q10", "r5a", 10),
			te("d20-tech1", "令和5年度秋期 技術候補1問目", 45, "問3・問5・問6から得意そうな分野を解く。", "r5a"),
			te("d20-tech2", "令和5年度秋期 技術候補2問目", 45, "問3・問5・問6から別の分野を解く。", "r5a"),
			t("d20-pick", "解く分野を決める", 15, "「分野」タブで選ぶ。"),
		},
	},
	{
		Date: "2026-11-21", Dow: "土", Phase: "分野別演習", Title: "必須と主力分野を固める",
		Goal: "問1と主力分野を2年度分解きます。", Finish: "すべての間違いを原因別に分類できた。",
		Point: "間違いは「知識不足・読み落とし・表現ミス」に分類します。",
		Tasks: []Task{
			q("d21-a1", "r6h", 1), q("d21-a9", "r6h", 9), q("d21-a10", "r6h", 10), q("d21-a11", "r6h", 11),
			q("d21-b1", "r5h", 1), q("d21-b9", "r5h", 9), q("d21-b10", "r5h", 10),
			q("d21-b2", "r5h", 2), // moved from 11/22 (solved on 11/21)
			q("d21-b11", "r5h", 11),
		},
	},
	{
		Date: "2026-11-22", Dow: "日", Phase: "分野別演習", Title: "新しい主力分野を試し、型をつかむ",
		Goal: "経営戦略とネットワークを2問ずつ解き、保険の監査と組込みも1問ずつ試します。", Finish: "メイン4分野と保険2分野をすべて1回以上解いた。",
		Point: "今日は25分を必ず使い切ります。早く終わったら、根拠の行を見直します。",
		Tasks: []Task{
			q("d22-st-a", "r6h", 2), q("d22-st-b2", "r4a", 2),
			q("d22-nw-a", "r6h", 5), q("d22-nw-b", "r5h", 5),
			q("d22-au", "r5a", 11), q("d22-es", "r6h", 7),
			t("d22-note", "決まり文句ノートを更新する", 60, "今日の解答例から、分野ごとの言い回しを「決まり文句」タブに足す。"),
		},
	},
	{
		Date: "2026-11-23", Dow: "月祝", Phase: "模試と補強", Title: "150分模試で実力を測る", Mock: "m1",
		Goal: "本番と同じ150分で5問を解き、60点との差を把握します。", Finish: "得点を記録し、不足分の原因が分かった。",
		Point: "1問25分を超えたら、途中でも次へ進みます。",
		Tasks: []Task{
			te("d23-mock", "令和6年度秋期 150分模試", 150, "問1＋メインから4分野を通しで解く。途中で解答を見ない。", "r6a"),
			te("d23-review", "模試の採点と復習", 150, "解答例と採点講評を読み、失点の原因を分類する。", "r6a"),
			t("d23-fix", "弱点を補強する", 60, "失点の大きい分野の知識を確認し、ノートに追記する。"),
		},
	},
	{
		Date: "2026-11-24", Dow: "火", Phase: "模試と補強", Title: "模試の失点を取り返す",
		Goal: "模試で失点が大きかった問題を解き直し、経営戦略とDBを1問ずつ解きます。", Finish: "同じ原因で間違えなくなった。",
		Point: "新しい問題より、間違えた問題の解き直しを優先します。",
		Tasks: []Task{
			te("d24-a", "失点が最も大きい問題を解き直す", 45, "解答の根拠となる本文に印を付ける。", "r6a"),
			te("d24-b", "失点が2番目に大きい問題を解き直す", 45, "解答の根拠となる本文に印を付ける。", "r6a"),
			q("d24-st", "r5a", 2), q("d24-db", "r6h", 6),
			t("d24-note", "弱点ノートを更新する", 15, "今日分かったことを1行ずつ追記する。"),
		},
	},
	{
		Date: "2026-11-25", Dow: "水", Phase: "模試と補強", Title: "問1と監査・PMを安定させる",
		Goal: "必須の問1を25分で解き切り、点の低い監査とPMを補強します。", Finish: "問1を25分以内に解き切れた。",
		Point: "監査は抜き出し問題です。空欄の前後と同じ主語の文を本文から探します。",
		Tasks: []Task{
			q("d25-sec-a", "r4a", 1), q("d25-sec-b", "r4h", 1),
			q("d25-au", "r4a", 11), q("d25-pm", "r5a", 9),
			t("d25-note", "弱点ノートを更新する", 15, "セキュリティ用語の覚え漏れを追記する。"),
		},
	},
	{
		Date: "2026-11-26", Dow: "木", Phase: "模試と補強", Title: "2回目の150分模試", Mock: "m2",
		Goal: "時間内に5問を解き切ることを最優先にします。", Finish: "150分で5問すべてに解答できた。",
		Point: "今日は解くだけで大丈夫です。復習は明日に回します。",
		Tasks: []Task{te("d26-mock", "令和7年度春期 150分模試", 150, "残り時間の目安を見ながら通しで解く。", "r7h")},
	},
	{
		Date: "2026-11-27", Dow: "金", Phase: "模試と補強", Title: "模試を復習し、弱点を潰す",
		Goal: "2回目の模試の失点を、原因から修正します。", Finish: "間違えた問題の解答根拠を説明できる。",
		Point: "1回目の模試と同じ間違いがないか確認します。",
		Tasks: []Task{
			te("d27-review", "模試の採点と復習", 90, "解答例と採点講評で失点の原因を確認する。", "r7h"),
			te("d27-redo", "失点が最も大きい問題を解き直す", 45, "時間を計り、PCで解答を打つ。", "r7h"),
			t("d27-note", "弱点ノートを更新する", 15, "本番前に見返す内容だけを残す。"),
		},
	},
	{
		Date: "2026-11-28", Dow: "土", Phase: "仕上げ", Title: "本番と同じ時刻で最終模試", Mock: "m3",
		Goal: "17:00開始のリハーサルで、本番の手順を確定します。", Finish: "本番の手順に迷いがなくなった。",
		Point: "復習が終わったら新しい問題は解かず、早めに休みます。",
		Tasks: []Task{
			t("d28-note", "弱点ノートと決まり文句を見直す", 60, "昼のうちに確認する。"),
			te("d28-mock", "17:00 令和7年度秋期 最終模試", 150, "19:30まで本番と同じ時間帯で解く。", "r7a"),
			te("d28-review", "模試の採点と復習", 90, "致命的なミスだけを確認し、深追いしない。", "r7a"),
		},
	},
	{
		Date: "2026-11-29", Dow: "日", Phase: "仕上げ", Title: "本番当日", ExamDay: true,
		Goal: "午前に軽く見直し、落ち着いて17:00の試験に臨みます。", Finish: "試験を最後まで受け切る。",
		Point: "開始直後に5問を選び、「選択」ボタンを押します。",
		Tasks: []Task{
			t("d29-note", "弱点ノートと決まり文句を見直す", 45, "午前中だけにして、新しい問題は解かない。"),
			t("d29-flow", "本番の手順を確認する", 15, "「本番」タブの時間配分と注意点を読む。"),
			t("d29-id", "本人確認書類を持って出発する", 0, "有効な本人確認書類がないと受験できない。"),
			t("d29-arrive", "16:30の受付開始を目安に到着する", 0, "集合時刻は試験開始15分前の16:45。"),
			t("d29-exam", "17:00 試験開始", 150, "19:30まで。全力を出し切りましょう。"),
		},
	},
}

var phases = []string{"診断", "分野別演習", "模試と補強", "仕上げ"}

var mocks = [][2]string{{"m1", "11/23 令和6年度秋期"}, {"m2", "11/26 令和7年度春期"}, {"m3", "11/28 令和7年度秋期"}}

func dayByDate(date string) (Day, int, bool) {
	for i, d := range days {
		if d.Date == date {
			return d, i, true
		}
	}
	return Day{}, -1, false
}

func taskExists(id string) bool {
	for _, d := range days {
		for _, tk := range d.Tasks {
			if tk.ID == id {
				return true
			}
		}
	}
	return false
}

// Field choice settled on 11/21: 4 メイン (solved on exam day with 問1) and
// 2 保険. メイン may temporarily hold up to 5 while deciding.
const maxMain = 5

var (
	defaultMain      = []string{"db", "st", "nw", "pm"}
	defaultInsurance = []string{"au", "es"}
)

// Phrase groups for the 決まり文句 page.
type PhraseGroup struct {
	Field string
	Items [][2]string // 問われ方, 答えの型
}

var phraseGroups = []PhraseGroup{
	{"書き方の共通ルール", [][2]string{
		{"何について書くか（主語・対象）", "本文の言葉をそのまま使う"},
		{"それがどうなのか（結論）", "分野の決まり文句で言い切る"},
		{"終了条件・移行条件", "過程ではなく到達した状態（収束した、引継ぎが完了した）"},
		{"〇〇にとってのメリット", "主語を〇〇に置いて書く"},
		{"問題点を問われた", "結果（忙しい）より原因・手段（電話は時間が掛かる）"},
		{"記号か番号か", "設問文で答え方の形式を確認する"},
	}},
	{"プロジェクトマネジメント", [][2]string{
		{"PoC・試行をする理由", "実現に不確かな要素が多いから"},
		{"スコープ変更で報告するもの", "コスト・スケジュール（品質・コスト・納期）"},
		{"事前に合意しておく理由", "後で認識の違いによる手戻りやトラブルを防ぐため"},
		{"外部と情報共有する前に結ぶもの", "秘密保持契約（守秘契約）"},
		{"関係者を巻き込む理由", "利用者の視点で要件の漏れを防ぐため"},
		{"完了時点で評価できない理由", "効果が出るまで時間がかかる（例：収穫するまで評価できない）"},
		{"リスクへの対応", "回避・軽減・転嫁・受容"},
		{"人員を増やす", "誰の何を増やすかを本文の言葉で（例：エスカレーション対応の開発課リソースの拡充）"},
	}},
	{"サービスマネジメント", [][2]string{
		{"下線の理由", "顧客と約束したSLAを守るため、SLAに影響する結果まで書く"},
		{"取り決めの文書", "社外（外部供給者）は契約書、社内（内部供給者）は合意文書"},
		{"問合せが増えると", "サービスデスク要員の体制が確保できない"},
		{"FAQ・セルフサービスの効果", "サポート時間帯以外でも利用者が自分で解決できる"},
		{"展開管理の視点", "適用のタイミングをコントロールできる"},
		{"窓口一本化のメリット", "利用者が問合せ先を判断しなくてよい"},
	}},
	{"システム監査", [][2]string{
		{"空欄の語句", "空欄の前後と同じ主語・同じ処理の文を探し、本文から抜き出す"},
		{"〜する前に確かめる作業", "時系列で、その処理より前の記述から探す"},
		{"発見事項と監査要点", "どのデータがどのデータから作られるかでつなぐ"},
	}},
	{"情報セキュリティ", [][2]string{
		{"感染時の初動", "ネットワークから切り離す"},
		{"ランサムウェア対策のバックアップ", "PCやネットワークから切り離して（オフラインで）保管する"},
		{"全従業員が担う役割", "検知・連絡受付（トリアージ以降はCSIRT）"},
		{"「全て選べ」", "問われている目的に必要かで1つずつ絞る"},
	}},
	{"ネットワーク", [][2]string{
		{"MXレコード", "そのドメイン宛てのメールを受け取るサーバを書く"},
		{"OP25B", "ISPが外向き25番ポートを遮断してスパム送信を防ぐ"},
		{"OP25B環境で外へ送る", "ISPのメールサーバを中継サーバとして送信する"},
		{"経路・フィルタの宛先", "ネットワークアドレス（末尾が0）"},
		{"社内から参照する設定", "内部のIPアドレスを書く"},
	}},
}
