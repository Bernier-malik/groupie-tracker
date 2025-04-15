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

    -- GamePlayers Table (Relationship between Players & Games)
    CREATE TABLE IF NOT EXISTS GamePlayers (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        game_id INTEGER NOT NULL,
        player_id INTEGER NOT NULL,
        score INTEGER DEFAULT 0,
        joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (game_id) REFERENCES Games(id) ON DELETE CASCADE,
        FOREIGN KEY (player_id) REFERENCES Players(id) ON DELETE CASCADE
    );

    -- GameRounds Table (Tracks each round in a game)
    CREATE TABLE IF NOT EXISTS GameRounds (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        game_id INTEGER NOT NULL,
        song_id INTEGER NOT NULL,
        round_number INTEGER NOT NULL,
        start_time DATETIME DEFAULT CURRENT_TIMESTAMP,
        end_time DATETIME,
        FOREIGN KEY (game_id) REFERENCES Games(id) ON DELETE CASCADE,
        FOREIGN KEY (song_id) REFERENCES Songs(id) ON DELETE CASCADE
    );

    -- Answers Table (Tracks answers from players)
    CREATE TABLE IF NOT EXISTS Answers (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        game_id INTEGER NOT NULL,
        player_id INTEGER NOT NULL,
        round_id INTEGER NOT NULL,
        answer TEXT NOT NULL,
        is_correct BOOLEAN DEFAULT false,
        time_taken INTEGER,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (game_id) REFERENCES Games(id) ON DELETE CASCADE,
        FOREIGN KEY (player_id) REFERENCES Players(id) ON DELETE CASCADE,
        FOREIGN KEY (round_id) REFERENCES GameRounds(id) ON DELETE CASCADE
    );`

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
