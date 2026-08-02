package main

import (
	"time"
	"net/http"
)

type User struct {
	ID		int
	Username	string
	Email		string
	IsAdmin		bool
	CreatedAt	time.Time

	IsLoggedIn	bool
}

type UserInfo struct {
	User User
	IsLoggedIn bool
}

type TemplateData struct {
	IsLoggedIn bool
	Username string
}

func UserLookup(user_id int) (User, error) {
	var user User

	stmt := `SELECT
                        id,
                        username,
                        email,
                        is_admin,
                        created_at
                FROM users
                WHERE id = ?`

        err := db.QueryRow(stmt, user_id).Scan(
                &user.ID,
                &user.Username,
                &user.Email,
                &user.IsAdmin,
                &user.CreatedAt,
	)

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func RequireLogin(w http.ResponseWriter, r *http.Request) bool {
	if (!session_manager.Exists(r.Context(), "user_id")) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false
	}
	return true
}
