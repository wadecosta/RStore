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
    "errors"
    _ "embed"

    "golang.org/x/crypto/bcrypt"
    "github.com/alexedwards/scs/v2"
    "github.com/go-sql-driver/mysql"

    "github.com/oklog/ulid/v2"
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

	/* Tools Page */
	mux.HandleFunc("GET /tools", HandlerTools)

	/* User */
	mux.HandleFunc("POST /account-edit", UserEditHandler)

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
	if err := r.ParseForm(); err != nil {
		log.Println("error parsing login form:", err)
		tpl.ExecuteTemplate(w, "login.html", "Invalid username or password")
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	var (
		id           int
		passwordHash string
		isActive     bool
	)

	const query = `SELECT
				id,
				password_hash,
				is_active
			FROM users
			WHERE username = ?
			LIMIT 1
		`

	err := db.QueryRow(query, username).Scan(
		&id,
		&passwordHash,
		&isActive,
	)

	if errors.Is(err, sql.ErrNoRows) {
		tpl.ExecuteTemplate(w, "login.html", "Invalid username or password")
		return
	}

	if err != nil {
		log.Println("Database error during login:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError,)
		return
	}

	if !isActive {
		tpl.ExecuteTemplate(w, "login.html", "Invalid username or password")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))

	if err != nil {
		tpl.ExecuteTemplate(w, "login.html", "Invalid username or password")
		return
	}

	_, err = db.Exec(
		"UPDATE users SET last_login_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	if err != nil {
		log.Println("Unable to update last_login_at:", err)
	}

	if err := session_manager.RenewToken(r.Context()); err != nil {
		log.Println("Session token renewal failed:", err)

		tpl.ExecuteTemplate(
			w,
			"login.html",
			"There was a problem logging you in",
		)
		return
	}

	session_manager.Put(r.Context(), "user_id",id)

	log.Println("Successful login for user ID:", id)

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
	if err := r.ParseForm(); err != nil {
		log.Println("Error parsing registration form:", err)
		tpl.ExecuteTemplate(
			w,
			"register.html",
			"There was a problem registering this account",
		)
		return
	}

	/* Get and normalize form values */
	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")

	/* Validate username */
	if err := ValidateUsername(username); err != nil {
		tpl.ExecuteTemplate(w, "register.html", err.Error())
		return
	}

	/* Validate email */
	if err := ValidateEmail(email); err != nil {
		tpl.ExecuteTemplate(w, "register.html", err.Error())
		return
	}

	/* Check whether username already exists */
	var exists bool

	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)",
		username,
	).Scan(&exists)

	if err != nil {
		log.Println("Error checking username:", err)
		tpl.ExecuteTemplate(
			w,
			"register.html",
			"There was a problem registering this account",
		)
		return
	}

	if exists {
		tpl.ExecuteTemplate(
			w,
			"register.html",
			"Username is already taken",
		)
		return
	}

	/* Check whether email already exists */
	err = db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)",
		email,
	).Scan(&exists)

	if err != nil {
		log.Println("Error checking email:", err)
		tpl.ExecuteTemplate(
			w,
			"register.html",
			"There was a problem registering this account",
		)
		return
	}

	if exists {
		tpl.ExecuteTemplate(
			w,
			"register.html",
			"Email is already registered",
		)
		return
	}

	/* Create password hash */
	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		log.Println("bcrypt error:", err)
		tpl.ExecuteTemplate(
			w,
			"register.html",
			"There was a problem registering this account",
		)
		return
	}

	/* Generate public ULID */
	publicUserID := ulid.Make().String()

	/*
		Insert new user.

		id is generated automatically by MariaDB.
		user_id is the public ULID.
	*/
	const stmt = `
		INSERT INTO users (
			user_id,
			username,
			password_hash,
			email,
			is_admin
		)
		VALUES (?, ?, ?, ?, FALSE)
	`

	result, err := db.Exec(
		stmt,
		publicUserID,
		username,
		passwordHash,
		email,
	)

	if err != nil {
		var mysqlErr *mysql.MySQLError

		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			switch {
			case strings.Contains(mysqlErr.Message, "uq_users_username"):
				tpl.ExecuteTemplate(
					w,
					"register.html",
					"Username is already taken",
				)
				return

			case strings.Contains(mysqlErr.Message, "uq_users_email"):
				tpl.ExecuteTemplate(
					w,
					"register.html",
					"Email is already registered",
				)
				return

			case strings.Contains(mysqlErr.Message, "uq_users_user_id"):
				/*
					Extremely unlikely with ULIDs, but retrying would
					be another reasonable strategy here.
				*/
				log.Println("ULID collision:", err)

				tpl.ExecuteTemplate(
					w,
					"register.html",
					"There was a problem registering this account. Please try again.",
				)
				return
			}
		}

		log.Println("Error inserting new user:", err)

		tpl.ExecuteTemplate(
			w,
			"register.html",
			"There was a problem registering this account",
		)
		return
	}

	/* Get internal database ID */
	internalUserID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error getting new user ID:", err)
		tpl.ExecuteTemplate(
			w,
			"register.html",
			"There was a problem registering this account",
		)
		return
	}

	/* Create authenticated session */
	if err := session_manager.RenewToken(r.Context()); err != nil {
		log.Println("Session token renewal failed:", err)
		tpl.ExecuteTemplate(
			w,
			"register.html",
			"There was a problem creating your session",
		)
		return
	}

	session_manager.Put(
		r.Context(),
		"user_id",
		int(internalUserID),
	)

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

func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return errors.New("missing username field")
	}

	if len(username) < 4 || len(username) > 50 {
		return errors.New("username must be between 4 and 50 characters")
	}

	if !AllLower(username) {
		return errors.New("please only use lowercase for username")
	}

	for _, char := range username {
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
			return errors.New("username may only contain letters and numbers")
		}
	}

	return nil
}

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return errors.New("missing email field")
	}

	if !AllLower(email) {
		return errors.New("please only use lowercase for email")
	}

	return nil
}

func UserEditHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	/* Get and normalize form values */
	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))

	/* Validate username */
	if err := ValidateUsername(username); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	/* Validate email */
	if err := ValidateEmail(email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	/*
		Get the internal database user ID from the session.

		This is users.id, not the public ULID.
	*/
	userID := session_manager.GetInt(r.Context(), "user_id")

	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	const stmt = `
		UPDATE users
		SET
			username = ?,
			email = ?
		WHERE
			id = ?
	`

	_, err := db.Exec(
		stmt,
		username,
		email,
		userID,
	)

	if err != nil {
		var mysqlErr *mysql.MySQLError

		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			/*
				Distinguish between duplicate username and
				duplicate email using the named constraints.
			*/
			switch {
			case strings.Contains(mysqlErr.Message, "uq_users_username"):
				http.Error(
					w,
					"Username is already taken",
					http.StatusConflict,
				)
				return

			case strings.Contains(mysqlErr.Message, "uq_users_email"):
				http.Error(
					w,
					"Email is already registered",
					http.StatusConflict,
				)
				return

			default:
				http.Error(
					w,
					"Username or email is already in use",
					http.StatusConflict,
				)
				return
			}
		}

		log.Println("Error updating user details:", err)

		http.Error(
			w,
			"Unable to update user details",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
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
