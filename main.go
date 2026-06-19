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
	<meta name="color-scheme" content="light dark">
	<title>Number Storage</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
	<script>
	(function () {
		var stored = localStorage.getItem('theme') || 'auto';
		function resolve(t) {
			return t === 'auto'
				? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
				: t;
		}
		function apply(t) {
			document.documentElement.setAttribute('data-theme', resolve(t));
		}
		apply(stored);
		window.__applyTheme = apply;
		window.__initialTheme = stored;
		window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function () {
			if ((localStorage.getItem('theme') || 'auto') === 'auto') apply('auto');
		});
	})();
	</script>
	<style>
		.app-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 1rem;
			flex-wrap: wrap;
		}
		.app-header h1 { margin: 0; }
		.theme-switch {
			display: flex;
			align-items: center;
			gap: .5rem;
			margin: 0;
		}
		.theme-switch label { margin: 0; }
		.theme-switch select { width: auto; margin: 0; }
		.message {
			margin-top: var(--pico-spacing);
			padding: .75rem 1rem;
			border-radius: var(--pico-border-radius);
			border: 1px solid var(--pico-muted-border-color);
			background: var(--pico-code-background-color);
		}
		.message.error { color: var(--pico-del-color); }
		.message.success { color: var(--pico-ins-color); }
		.empty { color: var(--pico-muted-color); font-style: italic; }
	</style>
	<script type="module" src="https://cdn.jsdelivr.net/gh/starfederation/datastar@v1.0.0/bundles/datastar.js"></script>
</head>
<body>
	<main class="container">
		<header class="app-header">
			<h1>Number Storage</h1>
			<div class="theme-switch">
				<label for="themeSelect">Theme</label>
				<select id="themeSelect" data-on:change="updateTheme(el.value)">
					<option value="auto">Auto</option>
					<option value="light">Light</option>
					<option value="dark">Dark</option>
				</select>
			</div>
		</header>
		<div id="app">
			` + content + `
		</div>
	</main>
	<script>document.getElementById('themeSelect').value = window.__initialTheme || 'auto';</script>
	<script>
	function updateTheme(theme) {
		localStorage.setItem('theme', theme);
		window.__applyTheme(theme);
	}
	</script>
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
	b.WriteString(`<article>`)
	b.WriteString(`<form data-on:submit="@post('/api/numbers')">`)
	b.WriteString(`<label for="sevenDigit">7-Digit Number</label>`)
	b.WriteString(`<input type="text" id="sevenDigit" data-bind:seven-digit value="` + template.HTMLEscapeString(sevenDigit) + `" placeholder="1234567" pattern="\d{7}" required title="Exactly 7 digits">`)
	b.WriteString(`<label for="longNumber">10-20 Digit Number</label>`)
	b.WriteString(`<input type="text" id="longNumber" data-bind:long-number value="` + template.HTMLEscapeString(longNumber) + `" placeholder="1234567890" pattern="\d{10,20}" required title="Between 10 and 20 digits">`)
	b.WriteString(`<button type="submit">Store Numbers</button>`)
	if errMsg != "" {
		b.WriteString(`<div class="message error">` + template.HTMLEscapeString(errMsg) + `</div>`)
	}
	if msg != "" {
		b.WriteString(`<div class="message success">` + template.HTMLEscapeString(msg) + `</div>`)
	}
	b.WriteString(`</form>`)
	b.WriteString(`</article>`)

	if len(entries) > 0 {
		b.WriteString(`<h2>Stored Numbers</h2>`)
		b.WriteString(`<table>`)
		b.WriteString(`<thead><tr><th scope="col">ID</th><th scope="col">7-Digit</th><th scope="col">10-20 Digit</th><th scope="col">Created</th></tr></thead>`)
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
