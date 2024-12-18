package main

import (
	"errors"
	"fmt"

	// "html/template"
	"net/http"
	"strconv"

	"snippetbox.gobpo2002.io/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	for _, snippet := range snippets {
		fmt.Fprintf(w, "%+v\n", snippet)
	}

	// files := []string{
	// "./ui/html/base.html",
	// "./ui/html/pages/home.html",
	// "./ui/html/partials/nav.html",
	// }

	// ts, err := template.ParseFiles(files...)

	// if err != nil {
	// app.serverError(w, err)
	// return
	// }

	// err = ts.ExecuteTemplate(w, "base", nil)
	// if err != nil {
	// app.serverError(w, err)
	// }

	// w.Write([]byte("Hello from snippetbox"))
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	snippet, err := app.snippets.Get(id)

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// fmt.Fprintf(w, "Display a specific snippet with ID %d...", id)
	fmt.Fprintf(w, "%+v", snippet)
	// w.Write([]byte("Display a specific snippet..."))
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		// w.WriteHeader(405)
		// w.Write([]byte("Method Not Allowed"))
		// http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	title := "Merry Christmas!"
	content := "Jesus Christ is our Lord and Savior"
	expires := 365

	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet/view?id=%d", id), http.StatusSeeOther)
	w.Write([]byte("Create a new snippet..."))
}
