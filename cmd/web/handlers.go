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
			fmt.Println("ERR NO RECORD:", id)
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
	decodedForm.CheckField(validator.PermittedValue(decodedForm.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

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

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, http.StatusOK, "signup.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
			return
		} else {
			app.serverError(w, err)
		}

		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, http.StatusOK, "login.html", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	urlPath := app.sessionManager.PopString(r.Context(), "redirectedFromPage")

	fmt.Println("before redirect url = ", urlPath)

	if urlPath == "" {
		urlPath = "/"
	}

	http.Redirect(w, r, urlPath, http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	templateData := app.newTemplateData(r)
	app.render(w, http.StatusOK, "about.html", templateData)
}

func (app *application) userAccountView(w http.ResponseWriter, r *http.Request) {
	templateData := app.newTemplateData(r)

	id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	user, err := app.users.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			app.sessionManager.Put(r.Context(), "redirectedFromPage", r.URL.Path)
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		} else {
			app.serverError(w, err)
		}
		return
	}

	templateData.User = user
	app.render(w, http.StatusOK, "account.html", templateData)
}

type updatePasswordForm struct {
	CurrentPassword     string `form:"currentPassword"`
	NewPassword         string `form:"newPassword"`
	ConfirmNewPassword  string `form:"confirmNewPassword"`
	validator.Validator `form:"-"`
}

func (app *application) updateAccountPassword(w http.ResponseWriter, r *http.Request) {
	templateData := app.newTemplateData(r)
	templateData.Form = updatePasswordForm{}
	app.render(w, http.StatusOK, "change_password.html", templateData)
}

func (app *application) updateAccountPasswordPost(w http.ResponseWriter, r *http.Request) {
	var form updatePasswordForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.CurrentPassword), "currentPassword", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPassword), "newPassword", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.NewPassword, 8), "newPassword", "This field must be at least 8 characters long")
	form.CheckField(validator.NotBlank(form.ConfirmNewPassword), "confirmNewPassword", "This field cannot be blank")
	form.CheckField(form.NewPassword == form.ConfirmNewPassword, "confirmNewPassword", "Passwords do not match")

	if !form.Valid() {
		templateData := app.newTemplateData(r)
		templateData.Form = form
		app.render(w, http.StatusUnprocessableEntity, "change_password.html", templateData)
		return
	}

	id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	err = app.users.UpdatePassword(id, form.CurrentPassword, form.NewPassword)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("You've entered wrong current password")
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your password has been successfully changed!")
	http.Redirect(w, r, "/user/account", http.StatusSeeOther)
}
