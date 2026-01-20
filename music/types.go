package music

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

type Musica struct {
	Titulo string
	URL    string
	User   string
}

type StreamSession struct {
	Cancel chan bool
}

var (
	FilaMusica        = make(map[string][]*Musica)
	FilaMutex         sync.Mutex
	VoiceConnections  = make(map[string]*discordgo.VoiceConnection)
	StreamingSessions = make(map[string]*StreamSession)
)
