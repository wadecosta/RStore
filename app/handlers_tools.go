package main

import (
	"net/http"
	"log"
)

type Tools struct {
	User User
}

func HandlerTools(w http.ResponseWriter, r *http.Request) {
	if (!RequireLogin(w, r)) {
		return
	}

	user_id := session_manager.GetInt(r.Context(), "user_id")

	user, err := UserLookup(user_id)
	if err != nil {
		log.Println(err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user.IsLoggedIn = session_manager.Exists(r.Context(), "user_id")

	var tools Tools

	tools.User = user

	err = tpl.ExecuteTemplate(w, "tools.html", tools)
	if err != nil {
		log.Println(err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
