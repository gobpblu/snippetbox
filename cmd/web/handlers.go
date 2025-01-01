package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"snippetbox.gobpo2002.io/internal/models"
	"snippetbox.gobpo2002.io/internal/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// if r.URL.Path != "/" {
	// app.notFound(w)
	// return
	// }

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData := app.newTemplateData(r)
	templateData.Snippets = snippets

	app.render(w, http.StatusOK, "home.html", templateData)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	app.infoLog.Print("id", id)
	if err != nil || id < 1 {
		app.errorLog.Print(err.Error())
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

	templateData := app.newTemplateData(r)
	templateData.Snippet = snippet

	app.render(w, http.StatusOK, "view.html", templateData)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	data.Form = snippetCreateForm{
		Expires: 365,
	}

	app.render(w, http.StatusOK, "create.html", data)
}

type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	// if err != nil {
	// 	app.clientError(w, http.StatusBadRequest)
	// 	return
	// }

	var decodedForm snippetCreateForm

	err = app.decodePostForm(r, &decodedForm)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// form := snippetCreateForm{
	// 	Title:   r.PostForm.Get("title"),
	// 	Content: r.PostForm.Get("content"),
	// 	Expires: expires,
	// }

	decodedForm.CheckField(validator.NotBlank(decodedForm.Title), "title", "This field cannot be blank")
	decodedForm.CheckField(validator.MaxChars(decodedForm.Title, 100), "title", "This field cannot be more than 100 characters long")
	decodedForm.CheckField(validator.NotBlank(decodedForm.Content), "content", "This field cannot be blank")
	decodedForm.CheckField(validator.PermittedInt(decodedForm.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !decodedForm.Valid() {
		data := app.newTemplateData(r)
		data.Form = decodedForm
		app.render(w, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	id, err := app.snippets.Insert(decodedForm.Title, decodedForm.Content, decodedForm.Expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
	// w.Write([]byte("Create a new snippet..."))
}
