package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/paludett/bot_discord/music"
)

// getVoiceUsers retorna lista de usuários no canal de voz
func getVoiceUsers(s *discordgo.Session, guildID string, voiceChannelID string) []string {
	guild, err := s.Guild(guildID)
	if err != nil {
		return []string{}
	}

	var users []string
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == voiceChannelID {
			users = append(users, vs.UserID)
		}
	}
	return users
}

func Play(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "❌ Use: `!play <link do YouTube ou nome da música>`")
		return
	}

	// Verifica se o usuário está em um canal de voz
	voiceState, err := s.State.VoiceState(m.GuildID, m.Author.ID)
	if err != nil || voiceState.ChannelID == "" {
		s.ChannelMessageSend(m.ChannelID, "❌ Você precisa estar em um canal de voz para tocar música!")
		return
	}

	query := strings.Join(args[1:], " ")

	var videoURL string
	var titulo string

	// Verifica se é um link do YouTube
	if music.IsValidYouTubeURL(query) {
		videoURL = query
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("📥 Carregando vídeo..."))
		titulo = "Vídeo do YouTube"
	} else {
		// Busca a música no YouTube
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔍 Buscando: `%s`...", query))

		url, foundTitle, err := music.SearchYouTube(query)
		if err != nil {
			log.Printf("❌ Erro na busca: %v", err)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Erro ao buscar música: %v", err))
			return
		}

		if url == "" {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Nenhuma música encontrada para: `%s`", query))
			return
		}

		videoURL = url
		titulo = foundTitle
	}

	musica := &music.Musica{
		Titulo: titulo,
		URL:    videoURL,
		User:   m.Author.Username,
	}

	music.FilaMutex.Lock()
	if music.FilaMusica[m.GuildID] == nil {
		music.FilaMusica[m.GuildID] = []*music.Musica{}
	}
	music.FilaMusica[m.GuildID] = append(music.FilaMusica[m.GuildID], musica)
	posicao := len(music.FilaMusica[m.GuildID])
	music.FilaMutex.Unlock()

	if posicao == 1 {
		embed := &discordgo.MessageEmbed{
			Title:       "▶️ Tocando Agora",
			Description: fmt.Sprintf("**%s**", musica.Titulo),
			Color:       0x1DB954,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Pedido por %s", musica.User),
			},
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		go music.PlayAudio(s, m.GuildID, voiceState.ChannelID, m.ChannelID)
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("➕ **%s** adicionado à fila (posição #%d)", musica.Titulo, posicao))
	}
}

func Skip(s *discordgo.Session, m *discordgo.MessageCreate) {
	if session, ok := music.StreamingSessions[m.GuildID]; ok {
		// Pega informações do usuário
		voiceState, err := s.State.VoiceState(m.GuildID, m.Author.ID)
		if err != nil || voiceState.ChannelID == "" {
			s.ChannelMessageSend(m.ChannelID, "❌ Você precisa estar em um canal de voz para votar!")
			return
		}

		// Pega usuários no canal de voz
		voiceUsers := getVoiceUsers(s, m.GuildID, voiceState.ChannelID)

		if len(voiceUsers) <= 1 {
			// Se só tem 1 pessoa, pula direto
			session.Cancel <- true
			s.ChannelMessageSend(m.ChannelID, "⏭️ Música pulada!")
			return
		}

		// Inicia votação
		s.ChannelMessageSend(m.ChannelID, "🗳️ Votação iniciada para pular música...")
		_, approved, yesVotes, totalVotes := music.StartVote(s, m.GuildID, m.ChannelID, voiceUsers, "skip")

		if approved {
			session.Cancel <- true
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Música pulada! (Votos: %d/%d)", yesVotes, totalVotes))
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Votação recusada! (Votos: %d/%d)", yesVotes, totalVotes))
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, "❌ Não há música tocando!")
	}
}

func Stop(s *discordgo.Session, m *discordgo.MessageCreate) {
	music.FilaMutex.Lock()
	music.FilaMusica[m.GuildID] = []*music.Musica{}
	music.FilaMutex.Unlock()

	if session, ok := music.StreamingSessions[m.GuildID]; ok {
		session.Cancel <- true
		delete(music.StreamingSessions, m.GuildID)
	}

	if vc, ok := music.VoiceConnections[m.GuildID]; ok {
		vc.Disconnect()
		delete(music.VoiceConnections, m.GuildID)
	}

	s.ChannelMessageSend(m.ChannelID, "⏹️ Música parada e fila limpa!")
}

func Fila(s *discordgo.Session, m *discordgo.MessageCreate) {
	music.FilaMutex.Lock()
	defer music.FilaMutex.Unlock()

	fila := music.FilaMusica[m.GuildID]
	if len(fila) == 0 {
		s.ChannelMessageSend(m.ChannelID, "📭 A fila está vazia!")
		return
	}

	var fields []*discordgo.MessageEmbedField

	for i, musica := range fila {
		if i == 0 {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "🎵 Tocando Agora",
				Value:  fmt.Sprintf("**%s**\nPedido por %s", musica.Titulo, musica.User),
				Inline: false,
			})
		} else if i <= 10 {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("#%d", i),
				Value:  fmt.Sprintf("**%s**", musica.Titulo),
				Inline: false,
			})
		}
	}

	if len(fila) > 11 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "...",
			Value:  fmt.Sprintf("E mais %d música(s)", len(fila)-11),
			Inline: false,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:  "🎵 Fila de Músicas",
		Color:  0x1DB954,
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Total: %d música(s)", len(fila)),
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

func Clear(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Pega informações do usuário
	voiceState, err := s.State.VoiceState(m.GuildID, m.Author.ID)
	if err != nil || voiceState.ChannelID == "" {
		s.ChannelMessageSend(m.ChannelID, "❌ Você precisa estar em um canal de voz para votar!")
		return
	}

	// Verifica se há música tocando
	music.FilaMutex.Lock()
	fila := music.FilaMusica[m.GuildID]
	music.FilaMutex.Unlock()

	if len(fila) == 0 {
		s.ChannelMessageSend(m.ChannelID, "📭 A fila já está vazia!")
		return
	}

	// Pega usuários no canal de voz
	voiceUsers := getVoiceUsers(s, m.GuildID, voiceState.ChannelID)

	if len(voiceUsers) <= 1 {
		// Se só tem 1 pessoa, limpa direto
		music.FilaMutex.Lock()
		music.FilaMusica[m.GuildID] = []*music.Musica{}
		music.FilaMutex.Unlock()

		if session, ok := music.StreamingSessions[m.GuildID]; ok {
			session.Cancel <- true
		}

		s.ChannelMessageSend(m.ChannelID, "🗑️ Fila limpa!")
		return
	}

	// Inicia votação
	s.ChannelMessageSend(m.ChannelID, "🗳️ Votação iniciada para limpar a fila...")
	_, approved, yesVotes, totalVotes := music.StartVote(s, m.GuildID, m.ChannelID, voiceUsers, "clear")

	if approved {
		music.FilaMutex.Lock()
		music.FilaMusica[m.GuildID] = []*music.Musica{}
		music.FilaMutex.Unlock()

		if session, ok := music.StreamingSessions[m.GuildID]; ok {
			session.Cancel <- true
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Fila limpa! (Votos: %d/%d)", yesVotes, totalVotes))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Votação recusada! (Votos: %d/%d)", yesVotes, totalVotes))
	}
}
