package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ByChanderZap/snippetbox/internal/models"
	"github.com/ByChanderZap/snippetbox/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type snippetCreateForm struct {
	Title   string `form:"title"`
	Content string `form:"content"`
	Expires int    `form:"expires"`
	// This "-" tells the decoder to ignore this field
	validator.Validator `form:"-"`
}

type userSignupForm struct {
	Name     string `form:"name"`
	Email    string `form:"email"`
	Password string `form:"password"`
	// This "-" tells the decoder to ignore this field
	validator.Validator `form:"-"`
}

type userSignInForm struct {
	Email    string `form:"email"`
	Password string `form:"password"`

	validator.Validator `form:"-"`
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
	}

	data := app.newTemplateData(r)
	data.Snippets = snippets

	app.render(w, r, http.StatusOK, "home.tmpl", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	s, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
			return
		}
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippet = s

	app.render(w, r, http.StatusOK, "view.tmpl", data)
}

func (app *application) snippetCreateForm(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	data.Form = snippetCreateForm{
		Expires: 365,
	}

	app.render(w, r, http.StatusOK, "create.tmpl", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var form snippetCreateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValues(form.Expires, 1, 7, 365), "expires", "This field must be one of 1, 7, 365")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusBadRequest, "create.tmpl", data)
		return
	}

	id, err := app.snippets.Insert(models.InsertSnippetParams{
		Title:   form.Title,
		Content: form.Content,
		Expires: form.Expires,
	})

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}

	app.render(w, r, http.StatusOK, "signup.tmpl", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, r, http.StatusBadRequest)
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "this field cannot be empty")
	form.CheckField(validator.NotBlank(form.Email), "email", "this field cannot be empty")
	form.CheckField(validator.NotBlank(form.Password), "password", "this field cannot be empty")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "password must be at least 8 characters long")
	form.CheckField(validator.ValidEmail(form.Email, validator.EmailRX), "email", "invalid email")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusBadRequest, "signup.tmpl", data)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), 12)
	if err != nil {
		app.serverError(w, r, err)
	}

	err = app.users.Insert(models.InsertUserParams{
		Name:     form.Name,
		Email:    form.Email,
		Password: string(hashedPassword),
	})
	if err != nil {
		// i dont really like this, and one work around could be creating like a "UserModel.EmailInUse" function
		// but that approach might cause other issues, like a race condition where two users try to sign up with the same email at the same time
		// both users would pass that validation but then one would be created and the other one would violate the unique constraint
		if errors.Is(err, models.ErrDuplicatedEmail) {
			form.AddFieldError("email", "email already in use")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusBadRequest, "signup.tmpl", data)
			return
		}

		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "user successfully created, please sign in")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignInForm{}
	app.render(w, r, http.StatusOK, "login.tmpl", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userSignInForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "email cannot be empty")
	form.CheckField(validator.NotBlank(form.Password), "password", "password cannot be empty")
	form.CheckField(validator.ValidEmail(form.Email, validator.EmailRX), "email", "this is an invalid email address")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}

	uId, err := app.users.Authenticate(models.AuthenticateUserParams{Email: form.Email, Password: form.Password})
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Invalid email or password")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusBadRequest, "login.tmpl", data)
			return
		}
		app.serverError(w, r, err)
		return
	}

	if err = app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserId", uId)
	app.sessionManager.Put(r.Context(), "flash", "Sign in successfully")

	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	app.sessionManager.Remove(r.Context(), "authenticatedUserId")
	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
	}

	app.sessionManager.Put(r.Context(), "flash", "logout successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
