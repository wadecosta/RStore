package main

import (
	"log"
	"strings"
	"net/http"
)

type Note struct {
	ID	int
	Message	string
	Dash_ID int
}

func NoteAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	user_id := session_manager.GetString(r.Context(), "user_id")

	/* read form value */
	note_text := strings.TrimSpace(r.FormValue("noteText"))

	if note_text == "" {
		http.Error(w, "Note text is required", http.StatusBadRequest)
	}

	stmt := `
	INSERT INTO notes
			(user_id, note, to_delete)
		VALUES
			(?, ?, 0)
		`
	
	_, err := db.Exec(stmt, user_id, note_text)
	if err != nil {
		log.Println(err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func NoteEditHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	
	id := r.FormValue("id")
	message := strings.TrimSpace(r.FormValue("message"))

	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}
	
	if message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}
	
	user_id := session_manager.GetString(r.Context(), "user_id")
	
	stmt := `UPDATE notes SET note=? WHERE (id=? AND user_id=?)`

	_, err := db.Exec(stmt, message, id, user_id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to update note", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func NoteDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")

	if id == "" {
		http.Error(w, "Missing vendor ID", http.StatusBadRequest)
		return
	}

	user_id := session_manager.GetString(r.Context(), "user_id")

	stmt := `UPDATE notes SET to_delete=1 WHERE (id=? AND user_id=?)`

	_, err := db.Exec(stmt, id, user_id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to delete note", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
