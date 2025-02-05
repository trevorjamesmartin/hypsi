package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/adrg/xdg"
	_ "github.com/mattn/go-sqlite3"
)

type StateFactory struct{}

func (sf *StateFactory) Create() AppState {
	var path, port string
	var exists bool
	// initialize application state
	has := HypsiAppState{}
	has.SetRewind(0)

	// store path
	path, exists = os.LookupEnv("HYPSI_PATH")
	if !exists {
		path = fmt.Sprintf("%s/hypsi", xdg.DataHome)
	}
	has.SetStorePath(path)

	// port
	port, exists = os.LookupEnv("PORT")
	if !exists {
		port = "3000"
	}

	if _, err := strconv.Atoi(port); err != nil {
		fmt.Printf("PORT: (%s)", port)
		log.Fatal(err)
	}

	has.SetPort(port)

	// interrupt if running already
	http.Get(fmt.Sprintf("http://localhost:%s/interrupt", port))

	return &has
}

type AppState interface {
	GetRewind() int
	GetMessage() string
	GetStorePath() string
	GetPort() string

	SetRewind(int)
	SetMessage(string)
	SetStorePath(string)
	SetPort(string)

	Load()
	Save()
}

type HypsiAppState struct {
	Rewind    int    `json:"rewind"`
	Message   string `json:"message,omitempty"`
	StorePath string `json:""`
	Port      string `json:""`
}

func (has *HypsiAppState) SetPort(value string) {
	has.Port = value
}

func (has *HypsiAppState) GetPort() string {
	return has.Port
}

func (has *HypsiAppState) GetStorePath() string {
	return has.StorePath
}

func (has *HypsiAppState) SetStorePath(path string) {
	var err error

	// create if needed
	if _, err = os.Stat(path); os.IsNotExist(err) {
		// create with 0755 permissions (read, write, and execute for owner, read and execute for group and others)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			log.Fatal(err) // Handle the error appropriately
		}
	}

	has.StorePath = path
}

func (has *HypsiAppState) GetRewind() int {
	return has.Rewind
}

func (has *HypsiAppState) GetMessage() string {
	return has.Message
}

func (has *HypsiAppState) SetRewind(value int) {
	has.Rewind = value
}

func (has *HypsiAppState) SetMessage(value string) {
	has.Message = value
}

func (has *HypsiAppState) Load() {
	var id, rewind int
	var message string

	sqlData := openDatabase()
	defer sqlData.Close()

	row := sqlData.QueryRow(`select * from state order by id desc limit 1`)

	if row.Scan(&id, &rewind, &message) != nil {
		has.SetRewind(0)
	} else {
		has.SetRewind(rewind)
		has.SetMessage(message)
	}
}

func (has *HypsiAppState) Save() {
	var id, rewind int
	var message, stmt string

	sqlData := openDatabase()
	defer sqlData.Close()

	row := sqlData.QueryRow(`select * from state order by id desc limit 1`)
	if row.Scan(&id, &rewind, &message) != nil {
		stmt = fmt.Sprintf(`insert into state(id, rewind, message) values(%d, %d, '%s');`, 0, has.Rewind, has.Message)
	} else {
		stmt = fmt.Sprintf(`update state set rewind=%d, message='%s' where id=%d;`, has.Rewind, has.Message, id)
	}
	_, err := sqlData.Exec(stmt)
	if err != nil {
		fmt.Printf("%q: %s\n", err, stmt)
	}

}

type HyprCtlActiveWorkspace struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	Monitor         string `json:"monitor"`
	MonitorID       int    `json:"monitorID"`
	Windows         int    `json:"windows"`
	HasFullScreen   bool   `json:"hasFullScreen"`
	LastWindow      string `json:"lastWindow"`
	LastWindowTitle string `json:"lastWindowTitle"`
}

type WorkspaceActor struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// monitor as defined by hyprctl
type HyprMonitor struct {
	Id               int            `json:"id,omitempty"`
	Name             string         `json:"name"`
	Description      string         `json:",omitempty"`
	Make             string         `json:",omitempty"`
	Model            string         `json:",omitempty"`
	Serial           string         `json:",omitempty"`
	Width            float64        `json:",omitempty"`
	Height           float64        `json:",omitempty"`
	RefreshRate      float64        `json:",omitempty"`
	X                int            `json:"x"`
	Y                int            `json:"y"`
	ActiveWorkspace  WorkspaceActor `json:"activeWorkspace"`
	SpecialWorkspace WorkspaceActor `json:"specialWorkspace"`
	Reserved         []int          `json:"reserved"`
	Scale            float64        `json:"scale"`
	Transform        int            `json:",omitempty"`
	Focused          bool           `json:",omitempty"`
	DpmsStatus       bool           `json:",omitempty"`
	Vrr              bool           `json:",omitempty"`
	Solitary         string         `json:",omitempty"`
	ActivelyTearing  bool           `json:",omitempty"`
	DirectScanTo     string         `json:",omitempty"`
	Disabled         bool           `json:",omitempty"`
	CurrentFormat    string         `json:",omitempty"`
	MirrorOf         string         `json:",omitempty"`
	AvailableModes   []string       `json:",omitempty"`
}

func (hm HyprMonitor) MarshallJSON() ([]byte, error) {
	return json.Marshal(hm)
}

func (hm *HyprMonitor) UnMarshallJSON(data []byte) error {
	var mon HyprMonitor
	if err := json.Unmarshal(data, &mon); err != nil {
		return err
	}
	*hm = mon
	return nil
}

type Plane struct {
	Monitor string `json:",omitempty"`
	Paper   string `json:",omitempty"`
	Mode    string `json:"Mode,omitempty"`
}

func (p Plane) MarshallJSON() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Plane) UnMarshallJSON(data []byte) error {
	var pln Plane
	if err := json.Unmarshal(data, &pln); err != nil {
		return err
	}
	*p = pln
	return nil
}

func (p *Plane) ToBase64() (string, error) {
	bts, err := os.ReadFile(p.Paper)

	if err != nil {
		return "", err
	}
	result := fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(bts), base64.StdEncoding.EncodeToString(bts))

	return result, nil
}

func (p *Plane) Thumb64() (string, error) {
	fileName := filepath.Base(p.Paper)
	thumbFile := fmt.Sprintf("thumb__%s", fileName)
	thumbPath := filepath.Join(HYPSI_STATE.GetStorePath(), thumbFile)

	if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
		makeThumbNail(p.Paper, thumbPath)
	}

	bts, err := os.ReadFile(thumbPath)

	if err != nil {
		return "", err
	}
	result := fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(bts), base64.StdEncoding.EncodeToString(bts))

	return result, nil
}

type History struct {
	dt   string // timestamp
	data string // json data
}

func (h *History) unfold() []Plane {
	if h.data[0] == '{' {
		// single monitor
		h.data = fmt.Sprintf("[%s]", h.data)
	}
	var target []Plane

	grief := json.Unmarshal([]byte(h.data), &target)
	if grief != nil {
		log.Fatalf("Unable to marshal JSON due to %s", grief)
	}

	return target
}

func openDatabase() *sql.DB {
	dbfile := filepath.Join(HYPSI_STATE.GetStorePath(), "hypsi.db")

	sqlDB, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		defer sqlDB.Close()
		log.Fatal(err)
	}

	sqlStmt := `
	create table if not exists history (
	id integer not null primary key,
	data text);
	create table if not exists state (
	id integer not null primary key,
	rewind integer,
	message text);
	create table if not exists localstorage (
	id integer not null primary key,
	data jsonb);
	`
	_, err = sqlDB.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err, sqlStmt)
	}

	return sqlDB
}

func writeHistory() {
	var id int
	var data string

	sqlData := openDatabase()
	defer sqlData.Close()

	row := sqlData.QueryRow(`select * from history order by id desc limit 1`)

	if row.Scan(&id, &data) != nil {
		id = -1
	}
	data = jsonText()

	stmt := fmt.Sprintf(`insert into history(id, data) values(%d, '%s');`, id+1, data)
	_, err := sqlData.Exec(stmt)
	if err != nil {
		fmt.Printf("%q: %s\n", err, stmt)
	}
}

func readHistory() ([]History, error) {
	var id int
	var data string
	var past []History

	sqlData := openDatabase()
	defer sqlData.Close()

	rows, err := sqlData.Query(`select * from history order by id asc`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&id, &data)
		if err != nil {
			log.Fatal(err)
		}
		past = append(past, History{dt: strconv.Itoa(id), data: data})
	}
	return past, nil
}

type Webpage struct {
	template string

	data struct {
		Version  string
		Style    template.CSS
		Monitors []*Plane       // active
		Hardware []*HyprMonitor // available
		Ivalue   bool
		Rewind   int
		Script   template.JS
		Core     HyprCtlVersion
	}

	funcMap template.FuncMap
}

func (w *Webpage) Print(out io.Writer, i int) {
	monitors, errListing := listActive()
	if errListing != nil {
		log.Fatal(errListing)
	}

	// these values change
	w.data.Rewind = i
	w.data.Ivalue = i >= 0
	w.data.Monitors = monitors

	// 'write => out' the resulting template.HTML
	template.Must(template.New("webpage").Funcs(w.funcMap).Parse(w.template)).Execute(out, w.data)
}

func (w Webpage) _Template() string {
	tmpl, staticError := WEBFOLDER.ReadFile("web/page.html.tmpl")
	if staticError != nil {
		log.Fatal(staticError)
	}
	return string(tmpl)
}
func webInit() Webpage {
	page := Webpage{}

	core, err := hyprCtlVersion()

	if err != nil {
		log.Fatal(err)
	}

	page.data.Core = core

	if hw, err := listMonitors(); err != nil {
		log.Fatal(err)
	} else {
		// static value for now
		page.data.Hardware = hw
	}

	hist, _ := readHistory()

	funcMap := template.FuncMap{
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
		"plusOne": func(n int) int {
			return n + 1
		},
		"lessOne": func(n int) int {
			return n - 1
		},
		"inHistory": func(n int) bool {
			return n < len(hist)
		},
		"gtZero": func(n int) bool {
			return n > 0
		},
	}
	// default page values
	page.template = page._Template()
	page.funcMap = funcMap
	page.data.Version = VERSION
	return page
}
