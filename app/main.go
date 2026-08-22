package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "html/template"
    "os"
    "time"
    "unicode"
    "strings"
    _ "embed"

    "golang.org/x/crypto/bcrypt"
    "github.com/alexedwards/scs/v2"
    "github.com/go-sql-driver/mysql"
)

var tpl *template.Template
var db *sql.DB
var session_manager *scs.SessionManager

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
	)

	var err error

	for i := 0; i < 10; i++ {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			err = db.Ping()
		}

		if (err == nil) {
			break
		}

		log.Println("Waiting for MariaDB...")
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err := InitSchema(db); err != nil {
		log.Fatal("Schema init err:", err)
	}

	tpl, err = template.ParseGlob("templates/*.html")
	if (err != nil) {
		log.Fatal("Error loading templates: ", err)
	}

	session_manager = scs.New()
	session_manager.Lifetime = 24 *  time.Hour
	session_manager.Cookie.HttpOnly = true
	session_manager.Cookie.SameSite = http.SameSiteLaxMode

	/* TODO CHANGE VALUE IF OUTSIDE OF LOCAL NETWORK */
	session_manager.Cookie.Secure = false

	mux := http.NewServeMux()
	
	// Using Go standard library ServeMux
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("GET /", HomeHandler)
	
	/* Login */
	mux.HandleFunc("GET /login", LoginHandler)
	mux.HandleFunc("POST /login", LoginSubmitHandler)

	/* Logout */
	mux.HandleFunc("POST /logout", LogoutHandler)

	/* Register */
	mux.HandleFunc("GET /register", RegisterHandler)
	mux.HandleFunc("POST /register", RegisterSubmitHandler)

	/* Profile Page */
	mux.HandleFunc("GET /profile", ProfileHandler)

	/* Dashboard Page */
	mux.HandleFunc("GET /dashboard", DashboardHandler)

	/* Vendor */
	mux.HandleFunc("POST /vendor-add", AddVendorHandler)
	mux.HandleFunc("POST /vendor-edit", EditVendorHandler)
	mux.HandleFunc("POST /vendor-delete", DeleteVendorHandler)

	/* Shipper */
	mux.HandleFunc("POST /shipper-add", AddShipperHandler)
	mux.HandleFunc("POST /shipper-edit", EditShipperHandler)
	mux.HandleFunc("POST /shipper-delete", DeleteShipperHandler)

	/* Notes */
	mux.HandleFunc("POST /note-add", AddNoteHandler)
	mux.HandleFunc("POST /note-edit", EditNoteHandler)
	mux.HandleFunc("POST /note-delete", DeleteNoteHandler)

	/* Combine PDFs */
	mux.HandleFunc("GET /merge-pdf", MergePDFHandler)
	mux.HandleFunc("POST /merge-pdf", MergePDFSubmitHandler)

	log.Println("Server listening on :8080")

	err = http.ListenAndServe(":8080", session_manager.LoadAndSave(mux))
	if err != nil {
		log.Fatal(err)
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	/* If the path isn't exactly "/", return a 404 instead of the home page */
	if (r.URL.Path != "/") {
		http.NotFound(w,r)
		return
	}

	if session_manager.Exists(r.Context(), "user_id") {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	is_logged_in := session_manager.Exists(r.Context(), "user_id")
	username := session_manager.GetString(r.Context(), "username")

	data := map[string]any {
		"is_logged_in" : is_logged_in,
		"username" : username,
	}

	tpl.ExecuteTemplate(w, "home.html", data)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if session_manager.Exists(r.Context(), "user_id") {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
        }

	err := tpl.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		log.Print(err)
	}
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
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

	var dash Dash

	dash.User = user
	
	dash.Vendors = []Vendor{}
	func() {
		stmt := `SELECT id, name, phone_number, email, auto_debit FROM vendors WHERE to_delete=0`
		rows, err := db.Query(stmt)
		if err != nil {
			log.Print("Error loading vendors:", err)
			return
		}
		defer rows.Close()

		idx := 0

		for rows.Next() {
			var v Vendor
			if err := rows.Scan(&v.ID, &v.Name, &v.Phone_Number, &v.Email, &v.Auto_Debit); err != nil {
				log.Print("Error scanning vendor:", err)
				continue
			}
			v.Dash_ID = idx
			dash.Vendors = append(dash.Vendors, v)
			idx++
		}
	}()

	dash.Shippers = []Shipper{}
	func() {
		stmt := `SELECT id, name, phone_number, email FROM shippers WHERE to_delete=0`
		rows, err := db.Query(stmt)
		if err != nil {
			log.Println("Error loading shippers:", err)
			return
		}
		defer rows.Close()

		idx := 0

		for rows.Next() {
			var s Shipper
			if err := rows.Scan(&s.ID, &s.Name, &s.Phone_Number, &s.Email); err != nil {
				log.Println("Error scanning shipper:", err)
				continue
			}
			s.Dash_ID = idx
			dash.Shippers = append(dash.Shippers, s)
			idx++
		}
	}()

	err = tpl.ExecuteTemplate(w, "dashboard.html", dash)
	if err != nil {
		log.Println(err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func LoginSubmitHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	var id int
	var password_hash string

	log.Println(id)
	log.Println(password_hash)
	
	stmt := "SELECT id, password FROM users WHERE username = ?"
	err := db.QueryRow(stmt, username).Scan(&id, &password_hash)
	if (err == sql.ErrNoRows) {
		tpl.ExecuteTemplate(w, "login.html", "Invaild username or password")
		return
	} else if (err != nil) {
		log.Println("Database error during login:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(password))
	if (err != nil) {
		tpl.ExecuteTemplate(w, "login.html", "Invaild username or password")
		return
	}

	err = session_manager.RenewToken(r.Context())
	if (err != nil) {
		log.Println("Session token renewal failed:", err)
		return
	}

	session_manager.Put(r.Context(), "user_id", id)

	log.Println("Success")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	err := session_manager.Destroy(r.Context())
	if (err != nil) {
		log.Println("Session destruction error:", err)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if session_manager.Exists(r.Context(), "user_id") {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	err := tpl.ExecuteTemplate(w, "register.html", nil)
	if err != nil {
		log.Print(err)
	}
}

func RegisterSubmitHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")

	/* Check username for only alphanumeric chars */
	var name_alphanumeric = true
	for _, char := range username {
		if (unicode.IsLetter(char) == false) && (unicode.IsNumber(char) == false) {
			name_alphanumeric = false
		}
	}

	/* Check length of username */
	var name_length bool
	if (len(username) >= 4) && (len(username) <= 50) {
		name_length = true
	}

	if (!name_alphanumeric || !name_length) {
		tpl.ExecuteTemplate(w, "register.html", "Please check username criteria")
		return
	}

	/* Check to make sure all values are lowercase */
	all_lower := AllLower(username)
	if (!all_lower) {
		tpl.ExecuteTemplate(w, "register.html", "Please only use lowercase for username")
		return
	}

	email := r.FormValue("email")
	all_lower = AllLower(email)
	if (!all_lower) {
		tpl.ExecuteTemplate(w, "register.html", "Please only use lowercase for email")
		return
	}

	password := r.FormValue("password")

	stmt := "SELECT id FROM users WHERE username = ? OR email = ?"
	row := db.QueryRow(stmt, username, email)

	var u_id string
	err := row.Scan(&u_id)

	if (err != sql.ErrNoRows) {
		tpl.ExecuteTemplate(w, "register.html", "username and/or email is already taken")
		return
	}

	/* Create hash from password */
	var password_hash []byte
	password_hash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if (err != nil) {
		log.Println("bcrypt err:", err)
		tpl.ExecuteTemplate(w, "register.html", "There is a problem registering this account")
		return
	}

	/* Insert user data into database */
	var insert_stmt *sql.Stmt
	insert_stmt, err = db.Prepare("INSERT INTO users (username, password, email, is_admin) VALUES (?, ?, ?, 0);")
	if (err != nil) {
		log.Println("error preparing statement:", err)
		tpl.ExecuteTemplate(w, "register.html", "There was a problem registering this account")
		return
	}

	defer insert_stmt.Close()

	var result sql.Result
	result, err = insert_stmt.Exec(username, password_hash, email)
	if (err != nil) {
		log.Println("error inserting new user: ", err)
		tpl.ExecuteTemplate(w, "register.html", "There was a problem registering this account")
		return
	}

	user_id, err := result.LastInsertId()
	if (err == nil) {
		_ = session_manager.RenewToken(r.Context())
		session_manager.Put(r.Context(), "user_id", int(user_id))
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
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

	var profile Profile

	profile.User = user

	profile.Notes = []Note{}
	func() {
		stmt := `SELECT id, note FROM notes WHERE (user_id=? AND to_delete=0)`
		rows, err := db.Query(stmt, user_id)
		if err != nil {
			log.Println("Error loading notes:", err)
			return
		}
		defer rows.Close()

		idx := 0

		for rows.Next() {
			var n Note
			if err := rows.Scan(&n.ID, &n.Message); err != nil {
				log.Println("Error scanning note:", err)
				continue
			}
			n.Dash_ID = idx
			profile.Notes = append(profile.Notes, n)
			idx++
		}
	}()

	err = tpl.ExecuteTemplate(w, "profile.html", profile)
	if err != nil {
		log.Println(err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func AddNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	user_id := session_manager.GetInt(r.Context(), "user_id")

	/* read form value */
	note_text := strings.TrimSpace(r.FormValue("noteText"))

	if note_text == "" {
		http.Error(w, "Note text is required", http.StatusBadRequest)
	}

	stmt :=	`
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

func EditNoteHandler(w http.ResponseWriter, r *http.Request) {
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

	user_id := session_manager.GetInt(r.Context(), "user_id")

	stmt := `UPDATE notes SET note=? WHERE (id=? AND user_id=?)`

	_, err := db.Exec(stmt, message, id, user_id)

	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to update note", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")

	if id == "" {
		http.Error(w, "Missing vendor ID", http.StatusBadRequest)
		return
	}

	user_id := session_manager.GetInt(r.Context(), "user_id")
	
	stmt := `UPDATE notes SET to_delete=1 WHERE (id=? AND user_id=?)`
	
	_, err := db.Exec(stmt, id, user_id)

	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to delete note", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func AddShipperHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	/* read form values */
	name := strings.TrimSpace(r.FormValue("name"))
	phone := strings.TrimSpace(r.FormValue("phone_number"))
	email := strings.TrimSpace(r.FormValue("email"))

	if name == "" {
		http.Error(w, "Vendor name is required", http.StatusBadRequest)
		return
	}

	if phone == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	stmt := `
		INSERT INTO shippers
			(name, phone_number, email, to_delete)
		VALUES
			(?, ?, ?, 0)
		`
	_, err := db.Exec(stmt, name, phone, email)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			http.Error(w, "Shipper already exists", http.StatusConflict)
			return
		}

		log.Println(err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func EditShipperHandler(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
                http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
                return
        }

        id := r.FormValue("id")
        name := strings.TrimSpace(r.FormValue("name"))
        phone := strings.TrimSpace(r.FormValue("phone_number"))
        email := strings.TrimSpace(r.FormValue("email"))

        if id == "" || name == "" || email == "" {
                http.Error(w, "Missing required fields", http.StatusBadRequest)
                return
        }

        stmt := `
                UPDATE shippers
                SET
                        name=?,
                        phone_number=?,
                        email=?
                WHERE id=?
                `

        _, err := db.Exec(
                stmt,
                name,
                phone,
                email,
                id,
        )

        if err != nil {
                log.Println(err)
                http.Error(w, "Unable to update vendor", http.StatusInternalServerError)
                return
        }

        http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func DeleteShipperHandler(w http.ResponseWriter, r *http.Request) {

    if r.Method != http.MethodPost {
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        return
    }

    id := r.FormValue("id")

    if id == "" {
        http.Error(w, "Missing vendor ID", http.StatusBadRequest)
        return
    }

    _, err := db.Exec(
        "UPDATE shippers SET to_delete=1 WHERE id=?",
        id,
    )

    if err != nil {
        log.Println(err)
        http.Error(w, "Unable to delete shipper", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}


func AddVendorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	/* read form values */
	name := strings.TrimSpace(r.FormValue("name"))
	phone := strings.TrimSpace(r.FormValue("phone_number"))
	email := strings.TrimSpace(r.FormValue("email"))

	/* checkbox returns "on" if checked, ""if unchecked */
	auto_debit := r.FormValue("auto_debit") == "on"

	/* basic validation */
	if name == "" {
		http.Error(w, "Vendor name is required", http.StatusBadRequest)
		return
	}

	if phone == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}


	stmt := `
		INSERT INTO vendors
			(name, phone_number, email, auto_debit, to_delete)
		VALUES
			(?, ?, ?, ?, 0)
		`
	
	_, err := db.Exec(stmt, name, phone, email, auto_debit)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			http.Error(w, "Vendor already exists", http.StatusConflict)
			return
		}

		log.Println(err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func EditVendorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	name := strings.TrimSpace(r.FormValue("name"))
	phone := strings.TrimSpace(r.FormValue("phone_number"))
	email := strings.TrimSpace(r.FormValue("email"))

	auto_debit := r.FormValue("auto_debit") == "on"

	if id == "" || name == "" || email == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	stmt := `
		UPDATE vendors
		SET
			name=?,
			phone_number=?,
			email=?,
			auto_debit=?
		WHERE id=?
		`
	
	_, err := db.Exec(
		stmt,
		name,
		phone,
		email,
		auto_debit,
		id,
	)

	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to update vendor", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func DeleteVendorHandler(w http.ResponseWriter, r *http.Request) {

    if r.Method != http.MethodPost {
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        return
    }

    id := r.FormValue("id")

    if id == "" {
        http.Error(w, "Missing vendor ID", http.StatusBadRequest)
        return
    }

    _, err := db.Exec(
        "UPDATE vendors SET to_delete=1 WHERE id=?",
        id,
    )

    if err != nil {
        log.Println(err)
        http.Error(w, "Unable to delete vendor", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
