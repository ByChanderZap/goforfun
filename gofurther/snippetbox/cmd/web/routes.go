package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /snippet/view/{id}", dynamic.ThenFunc(app.snippetView))
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))

	// Routes that required user to be authenticated
	// in case i dont want to use alice middlewares can be chained like this: (which somehow i think feels easier to understand)
	// mux.Handle("POST /snippet/create", app.sessionManager.LoadAndSave(app.requireAuthentication(http.HandlerFunc(app.snippetCreate)))
	authRoutes := dynamic.Append(app.requireAuth)
	mux.Handle("GET /snippet/create", authRoutes.ThenFunc(app.snippetCreateForm))
	mux.Handle("POST /snippet/create", authRoutes.ThenFunc(app.snippetCreatePost))
	mux.Handle("POST /user/logout", authRoutes.ThenFunc(app.userLogoutPost))

	standardMiddlewares := alice.New(app.recoverPanic, app.logRequest, commonHeader)
	return standardMiddlewares.Then(mux)
}
