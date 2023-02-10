package main

import (
	"fmt"
	"net/http"
)

func (app *application) recoverFromPanic(nxtHandler http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()

		nxtHandler.ServeHTTP(w, r)
	})
}
