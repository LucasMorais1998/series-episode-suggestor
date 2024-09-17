package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/LucasMorais1998/series-episode-suggestor/cache"
	"github.com/LucasMorais1998/series-episode-suggestor/models"

	_ "github.com/mattn/go-sqlite3"
)

const (
	apiEndpoint = "https://api.tvmaze.com/shows/66/episodes"
	dbFile      = "db/episodes.db"
)

func initEpisodesTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS watched_episodes (
			id INTEGER PRIMARY KEY
	);`
	_, err := db.Exec(query)
	return err
}

// func resetSuggestions(db *sql.DB) error {
// 	_, err := db.Exec("DELETE FROM watched_episodes")
// 	return err
// }

func markEpisodeAsWatched(db *sql.DB, episodeID int) error {
	_, err := db.Exec("INSERT INTO watched_episodes (id) VALUES (?)", episodeID)
	return err
}

func getWatchedEpisodes(db *sql.DB) ([]int, error) {
	rows, err := db.Query("SELECT id FROM watched_episodes")
	if err != nil {
			return nil, err
	}
	defer rows.Close()

	var watchedEpisodes []int
	for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
					return nil, err
			}
			watchedEpisodes = append(watchedEpisodes, id)
	}

	return watchedEpisodes, nil
}

func filterUnwatchedEpisodes(allEpisodes []models.Episode, watchedEpisodes []int) []models.Episode {
	watchedMap := make(map[int]bool)
	for _, id := range watchedEpisodes {
			watchedMap[id] = true
	}

	var unwatched []models.Episode
	for _, ep := range allEpisodes {
			if !watchedMap[ep.ID] {
					unwatched = append(unwatched, ep)
			}
	}

	return unwatched
}

func fetchEpisodes(db *sql.DB) ([]models.Episode, error) {
	cacheKey := "episodes"
	cachedData, err := cache.ReadCache(db, cacheKey)
	if err == nil && cachedData != nil {
			return cachedData, nil
	}

	resp, err := http.Get(apiEndpoint)
	if err != nil {
			return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
			return nil, err
	}

	var episodes []models.Episode
	err = json.Unmarshal(body, &episodes)
	if err != nil {
			return nil, err
	}

	err = cache.WriteCache(db, cacheKey, episodes)
	if err != nil {
			return nil, err
	}

	return episodes, nil
}

func suggestRandomEpisode(db *sql.DB) error {
	seed := time.Now().UnixNano()
	random := rand.New(rand.NewSource(seed))

	allEpisodes, err := fetchEpisodes(db)
	if err != nil {
			return err
	}

	watchedEpisodes, err := getWatchedEpisodes(db)
	if err != nil {
			return err
	}

	unwatchedEpisodes := filterUnwatchedEpisodes(allEpisodes, watchedEpisodes)
	if len(unwatchedEpisodes) == 0 {
			fmt.Println("Você já assistiu todos os episódios!")
			return nil
	}

	randomIndex := random.Intn(len(unwatchedEpisodes))
	randomEpisode := unwatchedEpisodes[randomIndex]

	fmt.Printf("Sugestão de episódio: %s (S%d - E%d)\n", randomEpisode.Name, randomEpisode.Season, randomEpisode.Number)
	return markEpisodeAsWatched(db, randomEpisode.ID)
}

func main() {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
			fmt.Println("Erro ao abrir banco de dados:", err)
			return
	}
	defer db.Close()

	err = cache.InitCacheTable(db)
	if err != nil {
			fmt.Println("Erro ao inicializar tabela de cache:", err)
			return
	}

	err = initEpisodesTable(db)
	if err != nil {
			fmt.Println("Erro ao inicializar tabela de episódios:", err)
			return
	}

	// Uncomment the line below to reset the suggestions
	// err = resetSuggestions(db)
	// if err != nil {
	//  fmt.Println("Erro ao resetar sugestões:", err)
	//  return
	// }

	err = suggestRandomEpisode(db)
	if err != nil {
			fmt.Println("Erro ao sugerir episódio:", err)
	}
}