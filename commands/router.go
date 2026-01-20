package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func Router(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	content := strings.TrimPrefix(m.Content, "!")
	args := strings.Fields(content)

	if len(args) == 0 {
		return
	}

	comando := strings.ToLower(args[0])

	switch comando {
	case "ping":
		Ping(s, m)
	case "oi", "ola":
		Oi(s, m)
	case "ajuda", "help":
		Ajuda(s, m)
	case "info":
		Info(s, m)
	case "dado":
		Dado(s, m, args)
	case "moeda":
		Moeda(s, m)
	case "play", "tocar":
		Play(s, m, args)
	case "skip", "pular":
		Skip(s, m)
	case "clear", "limpar":
		Clear(s, m)
	case "stop", "parar":
		Stop(s, m)
	case "fila", "queue":
		Fila(s, m)
	case "config":
		Config(s, m, args)
	default:
		s.ChannelMessageSend(m.ChannelID, "❌ Comando não reconhecido! Use `!ajuda` para ver os comandos disponíveis.")
	}
}
