package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"server/config"
	"server/models"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MusicRequest struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// GetSongText godoc
// @Summary Get song text by ID and paginate verses
// @Description Fetch the text of a specific song by ID, paginated by a fixed number of verses
// @Tags Songs
// @Accept json
// @Produce json
// @Param id path string true "Song ID"
// @Param page query int true "Page number (starting from 1)"
// @Success 200 {object} map[string]interface{} "Paginated verses of the song"
// @Failure 400 {string} string "Invalid parameters"
// @Failure 404 {string} string "Song not found"
// @Failure 500 {string} string "Internal server error"
// @Router /songs/{id}/text [get]
func GetSongText(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songID := vars["id"]

	pageParam := r.URL.Query().Get("page")
	if pageParam == "" {
		log.Printf("ERROR: Page parameter is required")
		http.Error(w, "Page parameter is required", http.StatusBadRequest)
		return
	}

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		log.Printf("ERROR: Invalid page parameter: %v", err)
		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
		return
	}

	const versesPerPage = 2

	db, err := gorm.Open(postgres.Open(config.GetDBConfig()), &gorm.Config{})
	if err != nil {
		log.Printf("ERROR: Database connection failed: %v", err)
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	var song models.Song
	if err := db.First(&song, songID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("INFO: Song with ID %s not found", songID)
			http.Error(w, fmt.Sprintf("Song with ID %s not found", songID), http.StatusNotFound)
		} else {
			log.Printf("ERROR: Failed to fetch song with ID %s: %v", songID, err)
			http.Error(w, "Failed to fetch song", http.StatusInternalServerError)
		}
		return
	}

	verses := strings.Split(song.Text, "\n")

	start := (page - 1) * versesPerPage
	end := start + versesPerPage
	if start >= len(verses) {
		log.Printf("ERROR: Page exceeds total verses")
		http.Error(w, "Page exceeds total verses", http.StatusBadRequest)
		return
	}
	if end > len(verses) {
		end = len(verses)
	}

	paginatedVerses := verses[start:end]

	response := map[string]interface{}{
		"song_id": song.ID,
		"title":   song.Name,
		"page":    page,
		"verses":  paginatedVerses,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ERROR: Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
// GetGroupWithSongs godoc
// @Summary Get songs by group
// @Description Fetch songs of a specific group with optional filters
// @Tags Songs
// @Accept json
// @Produce json
// @Param name query string false "Name of the song"
// @Param group_name query string false "Name of the group"
// @Param text query string false "Text contained in the song"
// @Param release_date query string false "Release date of the song"
// @Success 200 {object} map[string]interface{} "List of songs"
// @Failure 400 {string} string "Invalid parameters"
// @Failure 500 {string} string "Internal server error"
// @Router /songs [get]
func GetGroupWithSongs(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	groupName := r.URL.Query().Get("group_name")
	text := r.URL.Query().Get("text")
	releaseDate := r.URL.Query().Get("release_date")

	db, err := gorm.Open(postgres.Open(config.GetDBConfig()), &gorm.Config{})
	if err != nil {
		log.Printf("ERROR: Database connection failed: %v", err)
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	query := db.Table("songs").Select("songs.*, groups.id as group_id, groups.name as group_name").
		Joins("LEFT JOIN groups ON songs.group_id = groups.id")

	if name != "" {
		query = query.Where("songs.name ILIKE ?", "%"+name+"%")
		log.Printf("INFO: Filtering songs by name: %s", name)
	}
	if groupName != "" {
		query = query.Where("groups.name ILIKE ?", "%"+groupName+"%")
		log.Printf("INFO: Filtering songs by group: %s", groupName)
	}
	if text != "" {
		query = query.Where("songs.text ILIKE ?", "%"+text+"%")
		log.Printf("INFO: Filtering songs by text: %s", text)
	}
	if releaseDate != "" {
		query = query.Where("songs.release_date = ?", releaseDate)
		log.Printf("INFO: Filtering songs by release date: %s", releaseDate)
	}

	var songs []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		ReleaseDate string `json:"release_date"`
		Text        string `json:"text"`
		GroupID     int    `json:"group_id"`
		GroupName   string `json:"group_name"`
	}
	if err := query.Scan(&songs).Error; err != nil {
		log.Printf("ERROR: Failed to fetch songs: %v", err)
		http.Error(w, "Failed to fetch songs", http.StatusInternalServerError)
		return
	}

	if len(songs) == 0 {
		log.Printf("INFO: No songs found")
		http.Error(w, "No songs found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"group": map[string]interface{}{
			"id":   songs[0].GroupID,
			"name": songs[0].GroupName,
		},
		"songs": songs,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ERROR: Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// UpdateMusic godoc
// @Summary Update a song
// @Description Update the fields of an existing song by its ID
// @Tags Songs
// @Accept json
// @Produce json
// @Param id path string true "Song ID"
// @Param body body map[string]interface{} true "Fields to update (e.g., name, release_date, text)"
// @Success 200 {string} string "Song updated successfully"
// @Failure 400 {string} string "Invalid request body or date format"
// @Failure 404 {string} string "Song not found"
// @Failure 500 {string} string "Internal server error"
// @Router /songs/{id} [put]
func UpdateMusic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("INFO: Start updating song with ID %s", id)

	db, err := gorm.Open(postgres.Open(config.GetDBConfig()), &gorm.Config{})
	if err != nil {
		log.Printf("ERROR: Database connection failed: %v", err)
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var updatedFields map[string]interface{}
	if err := json.Unmarshal(body, &updatedFields); err != nil {
		log.Printf("ERROR: Failed to parse song info: %v", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	var existingSong models.Song
	if err := db.First(&existingSong, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("INFO: Song with ID %s not found", id)
			http.Error(w, fmt.Sprintf("Song with ID %s not found", id), http.StatusNotFound)
		} else {
			log.Printf("ERROR: Database query failed: %v", err)
			http.Error(w, "Database query failed", http.StatusInternalServerError)
		}
		return
	}

	if releaseDate, ok := updatedFields["release_date"]; ok {
		parsedDate, err := time.Parse("2006-01-02", releaseDate.(string))
		if err != nil {
			log.Printf("ERROR: Invalid release_date format: %v", err)
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		updatedFields["release_date"] = parsedDate
	}

	if err := db.Model(&existingSong).Updates(updatedFields).Error; err != nil {
		log.Printf("ERROR: Failed to update song with ID %s: %v", id, err)
		http.Error(w, fmt.Sprintf("Failed to update song with ID %s", id), http.StatusInternalServerError)
		return
	}

	log.Printf("INFO: Song with ID %s has been updated successfully", id)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Song with ID %s has been updated successfully", id)
}

// DeleteMusic godoc
// @Summary Delete a song
// @Description Delete a song by its ID
// @Tags Songs
// @Accept json
// @Produce json
// @Param id path string true "Song ID"
// @Success 200 {string} string "Song deleted successfully"
// @Failure 404 {string} string "Song not found"
// @Failure 500 {string} string "Internal server error"
// @Router /songs/{id} [delete]
func DeleteMusic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	log.Printf("INFO: Start deleting song with ID %s", id)

	db, err := gorm.Open(postgres.Open(config.GetDBConfig()), &gorm.Config{})
	if err != nil {
		log.Printf("ERROR: Database connection failed: %v", err)
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	if err := db.Delete(&models.Song{}, id).Error; err != nil {
		log.Printf("ERROR: Failed to delete song with ID %s: %v", id, err)
		http.Error(w, fmt.Sprintf("Failed to delete song with ID %s", id), http.StatusInternalServerError)
		return
	}

	log.Printf("INFO: Song with ID %s has been deleted", id)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Song with ID %s has been deleted", id)
}

// PostMusic godoc
// @Summary Add a new song
// @Description Add a new song to the database, including its group and details fetched from an external API
// @Tags Songs
// @Accept json
// @Produce json
// @Param body body handlers.MusicRequest true "Request body containing group and song names"
// @Success 200 {string} string "Song added successfully"
// @Failure 400 {string} string "Invalid JSON format or missing fields"
// @Failure 500 {string} string "Internal server error"
// @Router /songs [post]
func PostMusic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("ERROR: Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req MusicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: Invalid JSON format")
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	if req.Group == "" || req.Song == "" {
		log.Printf("ERROR: Missing 'group' or 'song' fields")
		http.Error(w, "Missing 'group' or 'song' fields", http.StatusBadRequest)
		return
	}
	apiURL := "http://server:8000/info"
	queryParams := fmt.Sprintf("?group=%s&song=%s", url.QueryEscape(req.Group), url.QueryEscape(req.Song))
	fullURL := apiURL + queryParams
	log.Printf("INFO: Calling /info with URL: %s", fullURL)

	resp, err := http.Get(fullURL)
	if err != nil {
		log.Printf("ERROR: Failed to fetch song info from /info")
		http.Error(w, "Failed to fetch song info from /info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read response from /info")
		http.Error(w, "Failed to read response from /info", http.StatusInternalServerError)
		return
	}

	var songDetail SongDetail
	log.Printf("INFO: Response from /info: %s", string(body))

	if err := json.Unmarshal(body, &songDetail); err != nil {
		log.Printf("ERROR: Failed to parse song info: %v", err)
		http.Error(w, "Failed to parse song info", http.StatusInternalServerError)
		return
	}

	db, err := gorm.Open(postgres.Open(config.GetDBConfig()), &gorm.Config{})
	if err != nil {
		log.Printf("ERROR: Database connection failed: %v", err)
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("ERROR: Transaction failed to start: %v", tx.Error)
		http.Error(w, "Transaction failed to start", http.StatusInternalServerError)
		return
	}
	var groupID uint
	err = tx.Raw("SELECT id FROM groups WHERE name = ?", req.Group).Scan(&groupID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		log.Printf("ERROR: Error checking group existence: %v", err)
		http.Error(w, "Error checking group existence", http.StatusInternalServerError)
		return
	}
	if groupID == 0 {
		// Вставка новой группы
		err = tx.Exec("INSERT INTO groups (name) VALUES (?)", req.Group).Error
		if err != nil {
			tx.Rollback()
			log.Printf("ERROR: Failed to create group: %v", err)
			http.Error(w, "Failed to create group", http.StatusInternalServerError)
			return
		}

		// Получаем ID новой группы
		err = tx.Raw("SELECT id FROM groups WHERE name = ?", req.Group).Scan(&groupID).Error
		if err != nil {
			tx.Rollback()
			log.Printf("ERROR: Error retrieving group ID: %v", err)
			http.Error(w, "Error retrieving group ID", http.StatusInternalServerError)
			return
		}
	}
	parsedDate, err := time.Parse("02.01.2006", songDetail.ReleaseDate) // Преобразуем дату из формата строки
	if err != nil {
		tx.Rollback()
		log.Printf("ERROR: Invalid date format: %v", err)
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}
	err = tx.Exec(`INSERT INTO songs (name, release_date, text, group_id) VALUES (?, ?, ?, ?)`, req.Song, parsedDate, songDetail.Text, groupID).Error

	if err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to insert song: %v", err)
		http.Error(w, "Failed to insert song", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to commit transaction: %v", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	log.Printf("INFO: Song added successfully")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Song added successfully"))
}
