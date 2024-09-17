package cache

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/LucasMorais1998/series-episode-suggestor/models"

	_ "github.com/mattn/go-sqlite3"
)

const (
    oneDayMS = 24 * 60 * 60 * 1000
)

func InitCacheTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS cache (
        id TEXT PRIMARY KEY,
        data TEXT,
        timestamp INTEGER
    );`
    _, err := db.Exec(query)
    return err
}

func ReadCache(db *sql.DB, key string) ([]models.Episode, error) {
    var data string
    var timestamp int64

    err := db.QueryRow("SELECT data, timestamp FROM cache WHERE id = ?", key).Scan(&data, &timestamp)
    if err != nil {
        return nil, err
    }

    if time.Now().Unix()*1000-timestamp < oneDayMS {
        var episodes []models.Episode
        err := json.Unmarshal([]byte(data), &episodes)
        if err != nil {
            return nil, err
        }
        return episodes, nil
    }

    return nil, nil
}

func WriteCache(db *sql.DB, key string, data []models.Episode) error {
    timestamp := time.Now().Unix() * 1000
    dataJSON, err := json.Marshal(data)
    if err != nil {
        return err
    }

    _, err = db.Exec("INSERT OR REPLACE INTO cache (id, data, timestamp) VALUES (?, ?, ?)", key, string(dataJSON), timestamp)
    return err
}