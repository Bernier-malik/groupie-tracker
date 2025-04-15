package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "app.db")
	if err != nil {
		log.Fatal("Erreur de connexion à SQLite:", err)
	}
	fmt.Println("Connexion à SQLite réussie!")

	migrate()
}

func migrate() {
	query := `
	CREATE TABLE IF NOT EXISTS Players (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE NOT NULL,
        password TEXT UNIQUE NOT NULL,
        pseudo TEXT UNIQUE NOT NULL,
        score INTEGER DEFAULT 0,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    
	CREATE TABLE IF NOT EXISTS scores (
        gameId TEXT,
        pseudo TEXT,
        score  INTEGER,
        PRIMARY KEY(gameId, pseudo)
    );
`

	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal("Erreur lors de la migration:", err)
	}
	fmt.Println("Tables créées avec succès!")
}

// Close database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
		fmt.Println("Connexion SQLite fermée.")
	}
}
