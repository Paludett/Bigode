package commands

import (
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

// send envia uma mensagem de texto simples no canal.
func send(e *events.MessageCreate, content string) {
	_, _ = e.Client().Rest.CreateMessage(e.ChannelID, discord.NewMessageCreate().WithContent(content))
}

// sendEmbed envia um embed no canal.
func sendEmbed(e *events.MessageCreate, embed discord.Embed) {
	_, _ = e.Client().Rest.CreateMessage(e.ChannelID, discord.NewMessageCreate().WithEmbeds(embed))
}

func Router(m *events.MessageCreate) {
	if m.Message.Author.ID == m.Client().ID() || m.Message.Author.Bot {
		return
	}

	if !strings.HasPrefix(m.Message.Content, "!") {
		return
	}

	content := strings.TrimPrefix(m.Message.Content, "!")
	args := strings.Fields(content)

	if len(args) == 0 {
		return
	}

	comando := strings.ToLower(args[0])

	switch comando {
	case "ping":
		Ping(m)
	case "oi", "ola":
		Oi(m)
	case "ajuda", "help":
		Ajuda(m)
	case "info":
		Info(m)
	case "dado":
		Dado(m, args)
	case "moeda":
		Moeda(m)
	case "play", "tocar":
		Play(m, args)
	case "skip", "pular":
		Skip(m)
	case "clear", "limpar":
		Clear(m)
	case "stop", "parar":
		Stop(m)
	case "fila", "queue":
		Fila(m)
	case "config":
		Config(m, args)
	default:
		send(m, "❌ Comando não reconhecido! Use `!ajuda` para ver os comandos disponíveis.")
	}
}
