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
	Dash_ID		int
}

func RecommendationAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	user_id := session_manager.GetString(r.Context(), "user_id")

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

func RecommendationEditHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("RAN RECOMMENDATION EDIT HANDLER")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	user_id := session_manager.GetString(r.Context(), "user_id")

	/* Get and normalize form values */
	id := strings.TrimSpace(r.FormValue("id"))
	item_name := strings.TrimSpace(r.FormValue("name"))
	item_link := strings.TrimSpace(r.FormValue("link"))

	if id == "" || item_name == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	log.Println(id)
	log.Println(item_name)
	log.Println(item_link)

	stmt := `
		UPDATE items_recommendations
		SET
			user_id=?,
			item=?,
			link=?
		WHERE id=?
		`
	
	_, err := db.Exec(
		stmt,
		user_id,
		item_name,
		item_link,
		id,
	)

	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to update recommendation", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
