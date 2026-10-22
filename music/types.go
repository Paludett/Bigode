package music

import (
	"context"
	"sync"

	"github.com/disgoorg/snowflake/v2"
)

type Musica struct {
	Titulo string
	URL    string
	User   string
}

// StreamSession representa a música que está tocando num servidor.
// Stop encerra o yt-dlp e o ffmpeg e pode ser chamado várias vezes.
type StreamSession struct {
	cancel context.CancelFunc
}

func (s *StreamSession) Stop() {
	s.cancel()
}

var (
	FilaMusica        = make(map[snowflake.ID][]*Musica)
	FilaMutex         sync.Mutex
	StreamingSessions = make(map[snowflake.ID]*StreamSession)
)

// GetStreamingSession retorna a sessão de áudio ativa do servidor, se houver.
func GetStreamingSession(guildID snowflake.ID) (*StreamSession, bool) {
	FilaMutex.Lock()
	defer FilaMutex.Unlock()
	session, ok := StreamingSessions[guildID]
	return session, ok
}
