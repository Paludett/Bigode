package music

import (
	"encoding/json"
	"log"
	"os/exec"
	"strings"
)

type YouTubeVideo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// SearchYouTube busca uma música no YouTube e retorna o primeiro resultado
func SearchYouTube(query string) (string, string, error) {
	// Usa yt-dlp para buscar
	cmd := exec.Command("yt-dlp",
		"-j", // JSON output
		"--no-warnings",
		"-f", "best",
		"ytsearch1:"+query) // Busca apenas 1 resultado

	output, err := cmd.Output()
	if err != nil {
		log.Printf("❌ Erro na busca: %v", err)
		return "", "", err
	}

	var result YouTubeVideo
	err = json.Unmarshal(output, &result)
	if err != nil {
		log.Printf("❌ Erro ao fazer parse: %v", err)
		return "", "", err
	}

	if result.ID == "" {
		return "", "", nil
	}

	url := "https://www.youtube.com/watch?v=" + result.ID
	return url, result.Title, nil
}

// ExtractYouTubeID extrai o ID do vídeo da URL
func ExtractYouTubeID(url string) string {
	if strings.Contains(url, "youtube.com/watch?v=") {
		parts := strings.Split(url, "v=")
		if len(parts) > 1 {
			return strings.Split(parts[1], "&")[0]
		}
	}
	if strings.Contains(url, "youtu.be/") {
		parts := strings.Split(url, "youtu.be/")
		if len(parts) > 1 {
			return strings.Split(parts[1], "?")[0]
		}
	}
	return ""
}

// IsValidYouTubeURL verifica se é uma URL do YouTube
func IsValidYouTubeURL(url string) bool {
	return strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")
}
