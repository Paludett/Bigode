package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/paludett/bot_discord/music"
)

// getVoiceUsers retorna a lista de usuários (humanos) no canal de voz
func getVoiceUsers(c *bot.Client, guildID snowflake.ID, voiceChannelID snowflake.ID) []snowflake.ID {
	var users []snowflake.ID
	for vs := range c.Caches.VoiceStates(guildID) {
		if vs.ChannelID == nil || *vs.ChannelID != voiceChannelID {
			continue
		}
		if vs.UserID == c.ID() {
			continue
		}
		if member, ok := c.Caches.Member(guildID, vs.UserID); ok && member.User.Bot {
			continue
		}
		users = append(users, vs.UserID)
	}
	return users
}

// userVoiceChannel retorna o canal de voz em que o autor da mensagem está.
func userVoiceChannel(m *events.MessageCreate) (snowflake.ID, bool) {
	if m.GuildID == nil {
		return 0, false
	}
	vs, ok := m.Client().Caches.VoiceState(*m.GuildID, m.Message.Author.ID)
	if !ok || vs.ChannelID == nil {
		return 0, false
	}
	return *vs.ChannelID, true
}

func Play(m *events.MessageCreate, args []string) {
	if len(args) < 2 {
		send(m, "❌ Use: `!play <link do YouTube ou nome da música>`")
		return
	}

	// Verifica se o usuário está em um canal de voz
	voiceChannelID, ok := userVoiceChannel(m)
	if !ok {
		send(m, "❌ Você precisa estar em um canal de voz para tocar música!")
		return
	}
	guildID := *m.GuildID

	query := strings.Join(args[1:], " ")

	var videoURL string
	var titulo string

	// Verifica se é um link do YouTube
	if music.IsValidYouTubeURL(query) {
		videoURL = query
		send(m, "📥 Carregando vídeo...")
		titulo = "Vídeo do YouTube"
	} else {
		// Busca a música no YouTube
		send(m, fmt.Sprintf("🔍 Buscando: `%s`...", query))

		url, foundTitle, err := music.SearchYouTube(query)
		if err != nil {
			log.Printf("❌ Erro na busca: %v", err)
			send(m, fmt.Sprintf("❌ Erro ao buscar música: %v", err))
			return
		}

		if url == "" {
			send(m, fmt.Sprintf("❌ Nenhuma música encontrada para: `%s`", query))
			return
		}

		videoURL = url
		titulo = foundTitle
	}

	musica := &music.Musica{
		Titulo: titulo,
		URL:    videoURL,
		User:   m.Message.Author.Username,
	}

	music.FilaMutex.Lock()
	if music.FilaMusica[guildID] == nil {
		music.FilaMusica[guildID] = []*music.Musica{}
	}
	music.FilaMusica[guildID] = append(music.FilaMusica[guildID], musica)
	posicao := len(music.FilaMusica[guildID])
	music.FilaMutex.Unlock()

	if posicao == 1 {
		embed := discord.Embed{
			Title:       "▶️ Tocando Agora",
			Description: fmt.Sprintf("**%s**", musica.Titulo),
			Color:       0x1DB954,
			Footer:      &discord.EmbedFooter{Text: fmt.Sprintf("Pedido por %s", musica.User)},
		}
		sendEmbed(m, embed)
		go music.PlayAudio(m.Client(), guildID, voiceChannelID, m.ChannelID)
	} else {
		send(m, fmt.Sprintf("➕ **%s** adicionado à fila (posição #%d)", musica.Titulo, posicao))
	}
}

func Skip(m *events.MessageCreate) {
	if m.GuildID == nil {
		return
	}
	guildID := *m.GuildID

	if session, ok := music.GetStreamingSession(guildID); ok {
		// Pega informações do usuário
		voiceChannelID, ok := userVoiceChannel(m)
		if !ok {
			send(m, "❌ Você precisa estar em um canal de voz para votar!")
			return
		}

		// Pega usuários no canal de voz
		voiceUsers := getVoiceUsers(m.Client(), guildID, voiceChannelID)

		if len(voiceUsers) <= 1 {
			// Se só tem 1 pessoa, pula direto
			session.Stop()
			send(m, "⏭️ Música pulada!")
			return
		}

		// Inicia votação
		send(m, "🗳️ Votação iniciada para pular música...")
		_, approved, yesVotes, totalVotes := music.StartVote(m.Client(), guildID, m.ChannelID, voiceUsers, "skip")

		if approved {
			session.Stop()
			send(m, fmt.Sprintf("✅ Música pulada! (Votos: %d/%d)", yesVotes, totalVotes))
		} else {
			send(m, fmt.Sprintf("❌ Votação recusada! (Votos: %d/%d)", yesVotes, totalVotes))
		}
	} else {
		send(m, "❌ Não há música tocando!")
	}
}

func Stop(m *events.MessageCreate) {
	if m.GuildID == nil {
		return
	}
	guildID := *m.GuildID

	music.FilaMutex.Lock()
	music.FilaMusica[guildID] = []*music.Musica{}
	music.FilaMutex.Unlock()

	if session, ok := music.GetStreamingSession(guildID); ok {
		session.Stop()
	}

	music.Disconnect(m.Client(), guildID)

	send(m, "⏹️ Música parada e fila limpa!")
}

func Fila(m *events.MessageCreate) {
	if m.GuildID == nil {
		return
	}

	music.FilaMutex.Lock()
	defer music.FilaMutex.Unlock()

	fila := music.FilaMusica[*m.GuildID]
	if len(fila) == 0 {
		send(m, "📭 A fila está vazia!")
		return
	}

	notInline := false
	var fields []discord.EmbedField

	for i, musica := range fila {
		if i == 0 {
			fields = append(fields, discord.EmbedField{
				Name:   "🎵 Tocando Agora",
				Value:  fmt.Sprintf("**%s**\nPedido por %s", musica.Titulo, musica.User),
				Inline: &notInline,
			})
		} else if i <= 10 {
			fields = append(fields, discord.EmbedField{
				Name:   fmt.Sprintf("#%d", i),
				Value:  fmt.Sprintf("**%s**", musica.Titulo),
				Inline: &notInline,
			})
		}
	}

	if len(fila) > 11 {
		fields = append(fields, discord.EmbedField{
			Name:   "...",
			Value:  fmt.Sprintf("E mais %d música(s)", len(fila)-11),
			Inline: &notInline,
		})
	}

	embed := discord.Embed{
		Title:  "🎵 Fila de Músicas",
		Color:  0x1DB954,
		Fields: fields,
		Footer: &discord.EmbedFooter{Text: fmt.Sprintf("Total: %d música(s)", len(fila))},
	}

	sendEmbed(m, embed)
}

func Clear(m *events.MessageCreate) {
	if m.GuildID == nil {
		return
	}
	guildID := *m.GuildID

	// Pega informações do usuário
	voiceChannelID, ok := userVoiceChannel(m)
	if !ok {
		send(m, "❌ Você precisa estar em um canal de voz para votar!")
		return
	}

	// Verifica se há música tocando
	music.FilaMutex.Lock()
	fila := music.FilaMusica[guildID]
	music.FilaMutex.Unlock()

	if len(fila) == 0 {
		send(m, "📭 A fila já está vazia!")
		return
	}

	// Pega usuários no canal de voz
	voiceUsers := getVoiceUsers(m.Client(), guildID, voiceChannelID)

	if len(voiceUsers) <= 1 {
		// Se só tem 1 pessoa, limpa direto
		clearQueue(guildID)
		send(m, "🗑️ Fila limpa!")
		return
	}

	// Inicia votação
	send(m, "🗳️ Votação iniciada para limpar a fila...")
	_, approved, yesVotes, totalVotes := music.StartVote(m.Client(), guildID, m.ChannelID, voiceUsers, "clear")

	if approved {
		clearQueue(guildID)
		send(m, fmt.Sprintf("✅ Fila limpa! (Votos: %d/%d)", yesVotes, totalVotes))
	} else {
		send(m, fmt.Sprintf("❌ Votação recusada! (Votos: %d/%d)", yesVotes, totalVotes))
	}
}

// clearQueue esvazia a fila e interrompe a música atual (encerrando yt-dlp e ffmpeg).
func clearQueue(guildID snowflake.ID) {
	music.FilaMutex.Lock()
	music.FilaMusica[guildID] = []*music.Musica{}
	music.FilaMutex.Unlock()

	if session, ok := music.GetStreamingSession(guildID); ok {
		session.Stop()
	}
}
