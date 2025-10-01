package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ByChanderZap/snippetbox/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

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
	w.Write([]byte("this is the form to create snippets"))
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	title := "O snail"
	content := "O snail\nClimb Mount Fuji\nBut slowly\n\n- kobayashi issa"
	expires := 7

	id, err := app.snippets.Insert(models.InsertSnippetParams{
		Title:   title,
		Content: content,
		Expires: expires,
	})
	if err != nil {
		app.serverError(w, r, err)
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
