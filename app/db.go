package main

import (
	"database/sql"
	"log"
)

func InitSchema(db *sql.DB) error {

	users_schema := `CREATE TABLE IF NOT EXISTS users (
				id INT PRIMARY KEY AUTO_INCREMENT,
				username VARCHAR(50) NOT NULL UNIQUE,
				password VARCHAR(255) NOT NULL,
				email VARCHAR(255) NOT NULL UNIQUE,
				is_admin BOOLEAN NOT NULL DEFAULT FALSE,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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
				user_id INT NOT NULL,
				note VARCHAR(280) NOT NULL,
				to_delete BOOLEAN NOT NULL DEFAULT FALSE,

				CONSTRAINT fk_notes_users
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

        log.Println("Schemas applied")
        return nil
}
