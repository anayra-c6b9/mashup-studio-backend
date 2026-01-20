package handlers

import (
	"encoding/json"
	"net/http"
)

type Track struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// GET /api/library
func Library(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tracks := []Track{
		{
			ID:    "gotye_somebody",
			Title: "Gotye - Somebody That I Used To Know (feat. Kimbra)",
			URL:   "/music/Gotye%20-%20Somebody%20That%20I%20Used%20To%20Know%20(feat.%20Kimbra)%20%5BOfficial%20Music%20Video%5D.mp3",
		},
		{
			ID:    "jvke_golden_hour",
			Title: "JVKE - golden hour",
			URL:   "/music/JVKE%20-%20golden%20hour%20(official%20music%20video).mp3",
		},
		{
			ID:    "yung_kai_blue",
			Title: "yung kai - blue",
			URL:   "/music/yung%20kai%20-%20blue%20(official%20music%20video).mp3",
		},
	}

	_ = json.NewEncoder(w).Encode(tracks)
}
