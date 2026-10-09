package main

import (
	"log"
	"strings"
	"net/http"
)

type ItemRecommendation struct {
	ID 		int
	Asked_By	string
	Item		string
	Link		string
	Dash_ID int
}

func RecommendationAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	user_id := session_manager.GetInt(r.Context(), "user_id")

	item_name := strings.TrimSpace(r.FormValue("item_name"))
	item_link := strings.TrimSpace(r.FormValue("item_link"))

	if item_name == "" {
		http.Error(w, "Item name is required", http.StatusBadRequest)
		return
	}

	stmt := `
		INSERT INTO items_recommendations
			(user_id, item, link, to_delete)
		VALUES
			(?, ?, ?, 0)
		`

	_, err := db.Exec(stmt, user_id, item_name, item_link)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to add recommendation. Please try later.", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
