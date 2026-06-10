package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/starfederation/datastar-go/datastar"
	_ "modernc.org/sqlite"
)

type NumberEntry struct {
	ID         int
	SevenDigit string
	LongNumber string
	CreatedAt  time.Time
}

type Signals struct {
	SevenDigit string `json:"sevenDigit"`
	LongNumber string `json:"longNumber"`
}

type App struct {
	DB *sql.DB
}

func main() {
	db, err := sql.Open("sqlite", "numbers.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		log.Fatal(err)
	}

	if _, err := db.Exec(`PRAGMA synchronous = NORMAL;`); err != nil {
		log.Fatal(err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS numbers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			seven_digit TEXT NOT NULL,
			long_number TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		log.Fatal(err)
	}

	app := &App{DB: db}

	http.HandleFunc("/", app.indexHandler)
	http.HandleFunc("/api/numbers", app.apiHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("Server starting on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	entries, _ := a.getEntries()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, pageTemplate(a.renderAppFragment("", "", nil, entries)))
}

func (a *App) apiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		entries, _ := a.getEntries()
		a.writeFragment(w, "", "Invalid request: "+err.Error(), signals, entries)
		return
	}

	sevenDigit := strings.TrimSpace(signals.SevenDigit)
	longNumber := strings.TrimSpace(signals.LongNumber)

	var errMsg string
	if matched, _ := regexp.MatchString(`^\d{7}$`, sevenDigit); !matched {
		errMsg = "Seven-digit number must be exactly 7 digits."
	} else if matched, _ := regexp.MatchString(`^\d{10,20}$`, longNumber); !matched {
		errMsg = "Long number must be between 10 and 20 digits."
	}

	if errMsg == "" {
		_, err := a.DB.Exec(
			"INSERT INTO numbers (seven_digit, long_number) VALUES (?, ?)",
			sevenDigit, longNumber,
		)
		if err != nil {
			errMsg = "Database error: " + err.Error()
		} else {
			// Clear inputs after successful save
			signals.SevenDigit = ""
			signals.LongNumber = ""
		}
	}

	entries, _ := a.getEntries()
	a.writeFragment(w, "", errMsg, signals, entries)
}

func (a *App) writeFragment(w http.ResponseWriter, msg, errMsg string, signals *Signals, entries []NumberEntry) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("datastar-selector", "#app")
	w.Header().Set("datastar-mode", "inner")
	fmt.Fprint(w, a.renderAppFragment(msg, errMsg, signals, entries))
}

func (a *App) getEntries() ([]NumberEntry, error) {
	rows, err := a.DB.Query(
		"SELECT id, seven_digit, long_number, created_at FROM numbers ORDER BY id DESC LIMIT 50",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []NumberEntry
	for rows.Next() {
		var e NumberEntry
		if err := rows.Scan(&e.ID, &e.SevenDigit, &e.LongNumber, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func pageTemplate(content string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Number Storage</title>
	<script type="module" src="https://cdn.jsdelivr.net/gh/starfederation/datastar@v1.0.0/bundles/datastar.js"></script>
	<style>
		body {
			font-family: system-ui, -apple-system, sans-serif;
			max-width: 700px;
			margin: 40px auto;
			padding: 0 20px;
			background: #f8f9fa;
			color: #212529;
		}
		#app {
			background: #fff;
			padding: 24px;
			border-radius: 8px;
			box-shadow: 0 2px 8px rgba(0,0,0,0.06);
		}
		h1 {
			margin-top: 0;
			font-size: 1.5rem;
		}
		.form-group {
			margin-bottom: 16px;
		}
		label {
			display: block;
			margin-bottom: 6px;
			font-weight: 600;
			font-size: 0.95rem;
		}
		input[type="text"] {
			width: 100%;
			padding: 10px 12px;
			font-size: 1rem;
			border: 1px solid #ced4da;
			border-radius: 6px;
			box-sizing: border-box;
		}
		input[type="text"]:focus {
			outline: none;
			border-color: #86b7fe;
			box-shadow: 0 0 0 3px rgba(13,110,253,0.15);
		}
		button {
			padding: 10px 18px;
			font-size: 1rem;
			background: #0d6efd;
			color: #fff;
			border: none;
			border-radius: 6px;
			cursor: pointer;
		}
		button:hover {
			background: #0b5ed7;
		}
		.message {
			margin-top: 12px;
			padding: 10px 12px;
			border-radius: 6px;
			font-size: 0.95rem;
		}
		.message.error {
			background: #f8d7da;
			color: #842029;
			border: 1px solid #f5c2c7;
		}
		.message.success {
			background: #d1e7dd;
			color: #0f5132;
			border: 1px solid #badbcc;
		}
		table {
			width: 100%;
			border-collapse: collapse;
			margin-top: 20px;
			font-size: 0.95rem;
		}
		th, td {
			text-align: left;
			padding: 10px 12px;
			border-bottom: 1px solid #dee2e6;
		}
		th {
			font-weight: 600;
			background: #f1f3f5;
		}
		tr:hover td {
			background: #f8f9fa;
		}
		.empty {
			margin-top: 16px;
			color: #6c757d;
			font-style: italic;
		}
	</style>
</head>
<body>
	<div id="app">
		` + content + `
	</div>
</body>
</html>`
}

func (a *App) renderAppFragment(msg, errMsg string, signals *Signals, entries []NumberEntry) string {
	sevenDigit := ""
	longNumber := ""
	if signals != nil {
		sevenDigit = signals.SevenDigit
		longNumber = signals.LongNumber
	}

	var b strings.Builder
	b.WriteString(`<h1>Number Storage</h1>`)
	b.WriteString(`<form data-on:submit="@post('/api/numbers')">`)
	b.WriteString(`<div class="form-group">`)
	b.WriteString(`<label for="sevenDigit">7-Digit Number</label>`)
	b.WriteString(`<input type="text" id="sevenDigit" data-bind:seven-digit value="` + template.HTMLEscapeString(sevenDigit) + `" placeholder="1234567" pattern="\d{7}" required title="Exactly 7 digits">`)
	b.WriteString(`</div>`)
	b.WriteString(`<div class="form-group">`)
	b.WriteString(`<label for="longNumber">10-20 Digit Number</label>`)
	b.WriteString(`<input type="text" id="longNumber" data-bind:long-number value="` + template.HTMLEscapeString(longNumber) + `" placeholder="1234567890" pattern="\d{10,20}" required title="Between 10 and 20 digits">`)
	b.WriteString(`</div>`)
	b.WriteString(`<button type="submit">Store Numbers</button>`)
	if errMsg != "" {
		b.WriteString(`<div class="message error">` + template.HTMLEscapeString(errMsg) + `</div>`)
	}
	if msg != "" {
		b.WriteString(`<div class="message success">` + template.HTMLEscapeString(msg) + `</div>`)
	}
	b.WriteString(`</form>`)

	if len(entries) > 0 {
		b.WriteString(`<h2>Stored Numbers</h2>`)
		b.WriteString(`<table>`)
		b.WriteString(`<thead><tr><th>ID</th><th>7-Digit</th><th>10-20 Digit</th><th>Created</th></tr></thead>`)
		b.WriteString(`<tbody>`)
		for _, e := range entries {
			b.WriteString(`<tr>`)
			b.WriteString(`<td>` + strconv.Itoa(e.ID) + `</td>`)
			b.WriteString(`<td>` + template.HTMLEscapeString(e.SevenDigit) + `</td>`)
			b.WriteString(`<td>` + template.HTMLEscapeString(e.LongNumber) + `</td>`)
			b.WriteString(`<td>` + e.CreatedAt.Format("2006-01-02 15:04:05") + `</td>`)
			b.WriteString(`</tr>`)
		}
		b.WriteString(`</tbody></table>`)
	} else {
		b.WriteString(`<p class="empty">No numbers stored yet.</p>`)
	}

	return b.String()
}
