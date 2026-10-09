package main

import (
	"database/sql"
	"log"
)

func InitSchema(db *sql.DB) error {

	users_schema := `CREATE TABLE IF NOT EXISTS users (
				id INT UNSIGNED NOT NULL AUTO_INCREMENT,
				user_id CHAR(26) NOT NULL,

				username VARCHAR(50) NOT NULL,
				password_hash VARCHAR(255) NOT NULL,
				email VARCHAR(255) NOT NULL,

				is_admin BOOLEAN NOT NULL DEFAULT FALSE,
				is_active BOOLEAN NOT NULL DEFAULT TRUE,
				is_verified BOOLEAN NOT NULL DEFAULT FALSE,

				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
					ON UPDATE CURRENT_TIMESTAMP,
				last_login_at TIMESTAMP NULL DEFAULT NULL,

				PRIMARY KEY (id),
				UNIQUE KEY uq_users_user_id (user_id),
				UNIQUE KEY uq_users_username (username),
				UNIQUE KEY uq_users_email (email)
			);`

	vendors_schema := `CREATE TABLE IF NOT EXISTS vendors (
				id INT PRIMARY KEY AUTO_INCREMENT,
				name VARCHAR(50) NOT NULL UNIQUE,
				phone_number VARCHAR(25) NOT NULL,
				email VARCHAR(255) NOT NULL,
				auto_debit BOOLEAN NOT NULL DEFAULT FALSE,
				to_delete BOOLEAN NOT NULL DEFAULT FALSE
			);`	
	
	shippers_schema := `CREATE TABLE IF NOT EXISTS shippers (
				id INT PRIMARY KEY AUTO_INCREMENT,
				name VARCHAR(50) NOT NULL UNIQUE,
				phone_number VARCHAR(25) NOT NULL,
				email VARCHAR(255) NOT NULL,
				to_delete BOOLEAN NOT NULL DEFAULT FALSE
			);`
	
	orders_schema := `CREATE TABLE IF NOT EXISTS orders (
				id INT PRIMARY KEY AUTO_INCREMENT,
				shipper_id INT NOT NULL,
				order_number VARCHAR(100) NOT NULL UNIQUE,
				tracking_number VARCHAR(100) NOT NULL UNIQUE,
				order_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				estimated_arrival_date DATETIME NULL,
				is_received BOOLEAN NOT NULL DEFAULT FALSE,
				received_by VARCHAR(50) NULL,
				to_delete BOOLEAN NOT NULL DEFAULT FALSE,

				CONSTRAINT fk_orders_shippers
					FOREIGN KEY (shipper_id)
					REFERENCES shippers(id)
					ON UPDATE CASCADE
					ON DELETE RESTRICT
			);`
	
	notes_schema := `CREATE TABLE IF NOT EXISTS notes (
				id INT PRIMARY KEY AUTO_INCREMENT,
				user_id INT UNSIGNED NOT NULL,
				note VARCHAR(280) NOT NULL,
				to_delete BOOLEAN NOT NULL DEFAULT FALSE,

				CONSTRAINT fk_notes_users
					FOREIGN KEY (user_id)
					REFERENCES users(id)
					ON UPDATE CASCADE
					ON DELETE CASCADE
				);`

	items_recommendations_schema := `CREATE TABLE IF NOT EXISTS items_recommendations (
						id INT PRIMARY KEY AUTO_INCREMENT,
						user_id INT UNSIGNED NOT NULL,
						item VARCHAR(280) NOT NULL,
						link VARCHAR(280),
						to_delete BOOLEAN NOT NULL DEFAULT FALSE,

						CONSTRAINT fk_items_recommendations_users
							FOREIGN KEY (user_id)
							REFERENCES users(id)
							ON UPDATE CASCADE
							ON DELETE CASCADE
						);`

	
        _, err := db.Exec(users_schema)
        if err != nil {
                log.Fatal("Users table err: ", err)
        }

	_, err = db.Exec(vendors_schema)
	if err != nil {
		log.Fatal("Vendors table err: ", err)
	}

	_, err = db.Exec(shippers_schema)
	if err != nil {
		log.Fatal("Shippers table err: ", err)
	}

	_, err = db.Exec(orders_schema)
	if err != nil {
		log.Fatal("Orders table err: ", err)
	}

	_, err = db.Exec(notes_schema)
	if err != nil {
		log.Fatal("Notes table err: ", err)
	}

	_, err = db.Exec(items_recommendations_schema)
	if err != nil {
		log.Fatal("Items Recommendations table err: ", err)
	}

        log.Println("Schemas applied")
        return nil
}
