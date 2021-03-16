// gopaste is a simple pastebin service written in go
package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	duration "github.com/channelmeter/iso8601duration"
	"github.com/dchest/uniuri"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

//go:generate cp config.example.json $PWD/config.json
//go:generate cp -r assets $PWD/

// Configuration read from config.json
type Configuration struct {
	// ADDRESS that gopaste will return links for
	Address string
	// LENGTH of paste id
	Length int
	// PORT that pastebin will listen on
	Port string
	// DBTYPE database type
	DBType string
	// DBNAME database name
	DBName string
	// DBUSERNAME for database
	DBUsername string
	// DBPASS database password
	DBPassword string
}

var configuration Configuration

// DB configuration used by SQL driver
var dbString, dbType string

// MAX_ID_LEN sets paste ID maximum string length
const MAX_ID_LEN int = 30

// MAX_KEY_LEN sets delkey maximum string length
const MAX_KEY_LEN int = 40

// pasteRegex lists paste ID authorized characters
var pasteRegex *regexp.Regexp = regexp.MustCompile("[^" + string(uniuri.StdChars) + "]")

// templates for HTML rendering
var templates = template.Must(template.ParseFiles(
	"assets/html/paste.html",
	"assets/html/index.html",
))

// InvalidPasteError save & get paste error structs
type InvalidPasteError struct{}

// InvalidPasteError save & get URL error structs
type InvalidURLError struct{}

func (e *InvalidPasteError) Error() string {
	return "Invalid paste"
}
func (e *InvalidURLError) Error() string {
	return "Invalid URL"
}

// Page generation struct
type Page struct {
	Title    string
	Body     []byte
	Raw      string
	Home     string
	Download string
	Clone    string
}

// Response paste struct
type Response struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
	Sha1    string `json:"sha1"`
	URL     string `json:"url"`
	Size    int    `json:"size"`
	Delkey  string `json:"delkey"`
}

// GetDB opens database and returns handle
func GetDB() *sql.DB {
	if dbType == "" || dbString == "" {
		log.Fatal("db.Open: Missing database configuration")
	}
	db, err := sql.Open(dbType, dbString)
	if err != nil {
		log.Fatal("db.Open:", err)
	}
	return db
}

// GenerateName uses uniuri to generate a random string that isn't in the database
func GenerateName() (string, error) {
	if configuration.Length <= 0 {
		return "", errors.New("Paste ID is too short")
	}

	db := GetDB()
	defer db.Close()

	for {
		id := uniuri.NewLen(configuration.Length)

		// query database if id exists and if it does call generateName again
		res, err := db.Query("SELECT id FROM pastebin WHERE id=?", id)
		if err != nil && err != sql.ErrNoRows {
			return "", err
		} else if err == sql.ErrNoRows || !res.Next() {
			return id, nil
		}
	}
}

// ValidPasteId scans paste ID for unauthorized characters
func ValidPasteId(pasteId string) bool {
	return len(pasteId) <= MAX_ID_LEN && !pasteRegex.Match([]byte(pasteId))
}

// Sha1 hashes paste content used for duplicate checks
func Sha1(paste string) string {
	hasher := sha1.New()

	hasher.Write([]byte(paste))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

// DurationFromExpiry parses expiration duration from HTML form
func DurationFromExpiry(expiry string) time.Duration {
	if expiry == "" { // Default duration is almost infinite (20 years)
		expiry = "P20Y"
	}
	dura, err := duration.FromString(expiry) // dura is time.Duration type
	if err != nil {
		log.Println(err)
	}

	if dura.Years > 20 { // Make sure we don't overflow duration at some point
		dura.Years = 20
	}

	return dura.ToDuration()
}

// Save pastes to database
func Save(raw string, expiry string, lang string) (*Response, error) {
	//log.Println(raw)
	httpFtp, _ := regexp.MatchString(`(http|ftp|https)://([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:/~+#-]*[\w@?^=%&/~+#-])?`, raw)
	//log.Println(httpFtp)
	if lang == "url" && !httpFtp {
		return nil, &InvalidURLError{}
	}
	db := GetDB()
	defer db.Close()

	// hash paste data and query database to see if paste exists
	sha := Sha1(raw)
	rows, err := db.Query("SELECT id, hash, data, delkey FROM pastebin WHERE hash=?", sha)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, hash, paste, delkey string
			err := rows.Scan(&id, &hash, &paste, &delkey)
			if err != nil {
				return nil, err
			}
			url := "/" + id
			return &Response{true, id, hash, url, len(paste), delkey}, nil
		}
	} else if err != sql.ErrNoRows {
		return nil, err
	}

	id, err := GenerateName()
	if err != nil {
		return nil, err
	}
	url := "/" + id

	const timeFormat = "2006-01-02 15:04:05"
	expiryTime := time.Now().Add(DurationFromExpiry(expiry)).Format(timeFormat)

	delKey := uniuri.NewLen(MAX_KEY_LEN)
	dataEscaped := html.EscapeString(raw)

	stmt, err := db.Prepare("INSERT INTO pastebin(id, hash, data, delkey, expiry,language) values(?,?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	_, err = stmt.Exec(id, sha, dataEscaped, delKey, expiryTime, lang)
	if err != nil {
		return nil, err
	}

	return &Response{true, id, sha, url, len(dataEscaped), delKey}, nil
}

// GetPaste extracts paste from database
func GetPaste(paste string) (string, string, error) {
	pasteId := html.EscapeString(paste)
	if !ValidPasteId(pasteId) {
		return "", "", &InvalidPasteError{}
	}

	db := GetDB()
	defer db.Close()

	var s, expiry, lang string
	err := db.QueryRow("SELECT data, expiry,language FROM pastebin WHERE id=?", pasteId).Scan(&s, &expiry, &lang)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", &InvalidPasteError{}
		}
		return "", "", err
	}

	//log.Print(time.Now().Format("2006-01-02T15:04:05Z"))
	//log.Print(expiry)

	if time.Now().Format("2006-01-02T15:04:05Z") >= expiry {
		stmt, err := db.Prepare("DELETE FROM pastebin WHERE id=?")
		if err != nil {
			return "", "", err
		}
		_, err = stmt.Exec(pasteId)
		if err != nil {
			return "", "", err
		}
		return "", "", &InvalidPasteError{}
	}

	return html.UnescapeString(s), lang, nil
}

// RootHandler handles generating the root page
func RootHandler(w http.ResponseWriter, _ *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", &Page{Title: "gopaste", Body: []byte("")})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// LoadConfiguration opens and reads config file, then sets up database
func LoadConfiguration() {
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		panic(err)
	}

	if configuration.Length > MAX_ID_LEN {
		configuration.Length = MAX_ID_LEN
	}

	if configuration.Length <= 0 {
		configuration.Length = 1
	}

	switch configuration.DBType {
	case "mysql":
		dbString = configuration.DBUsername + ":" + configuration.DBPassword + "@/" + configuration.DBName + "?charset=utf8"
		dbType = configuration.DBType
	case "sqlite3":
		dbString = "./" + configuration.DBName + "?charset=utf8"
		dbType = configuration.DBType
	default:
		panic("Incorrect DBType configuration")
	}
}

// CheckDB verifies DB link and table existence. Creates it if necessary
func CheckDB() {
	db := GetDB()
	defer db.Close()

	_, err := db.Query("SELECT 1 FROM pastebin LIMIT 1")
	if err != nil {
		_, err = db.Exec(`
        CREATE TABLE pastebin (
            id VARCHAR(30) NOT NULL,
            hash CHAR(40) DEFAULT NULL,
            data TEXT,
            delkey CHAR(40) default NULL,
            expiry DATETIME,
            language TEXT,
            PRIMARY KEY (id)
        );`)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// RawHandler handles raw paste content display
/*
func RawHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paste := vars["pasteId"]

	s, err := GetPaste(paste)
	if err != nil {
		switch err.(type) {
		case *InvalidPasteError:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8; imeanit=yes")
	// simply write string to browser
	io.WriteString(w, s)
}

// DownloadHandler handles paste content download
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paste := vars["pasteId"]

	s, err := GetPaste(paste)
	if err != nil {
		switch err.(type) {
		case *InvalidPasteError:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Set header to an attachment so browser will automatically download it
	w.Header().Set("Content-Disposition", "attachment; filename="+paste)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	io.WriteString(w, s)
}

// PasteHandler handles the generation of paste pages with the links
func PasteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paste := vars["pasteId"]

	s, err := GetPaste(paste)
	if err != nil {
		switch err.(type) {
		case *InvalidPasteError:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// button links
	link := configuration.Address + "/" + paste + "/raw"
	download := configuration.Address + "/" + paste + "/download"
	clone := configuration.Address + "/" + paste + "/clone"
	// Page struct
	p := &Page{
		Title:    paste,
		Body:     []byte(s),
		Raw:      link,
		Home:     configuration.Address,
		Download: download,
		Clone:    clone,
	}

	err = templates.ExecuteTemplate(w, "paste.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CloneHandler handles clone command and prefill new paste template with content
func CloneHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paste := vars["pasteId"]

	s, err := GetPaste(paste)
	if err != nil {
		switch err.(type) {
		case *InvalidPasteError:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// button links
	link := configuration.Address + paste + "/raw"
	download := configuration.Address + paste + "/download"
	clone := configuration.Address + paste + "/clone"
	// Page struct
	p := &Page{
		Title:    "Clone : " + paste,
		Body:     []byte(s),
		Raw:      link,
		Home:     configuration.Address,
		Download: download,
		Clone:    clone,
	}

	err = templates.ExecuteTemplate(w, "index.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func imain() {
	LoadConfiguration()
	CheckDB()

	router := mux.NewRouter()

	router.HandleFunc("/{pasteId}/raw", RawHandler).Methods("GET")
	router.HandleFunc("/{pasteId}/clone", CloneHandler).Methods("GET")
	router.HandleFunc("/{pasteId}/download", DownloadHandler).Methods("GET")
	router.HandleFunc("/{pasteId}", PasteHandler).Methods("GET")
	router.HandleFunc("/", RootHandler)
	err := http.ListenAndServe(configuration.Port, router)
	if err != nil {
		log.Fatal(err)
	}
}*/
