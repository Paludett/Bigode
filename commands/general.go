package commands

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Ping(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "🏓 Pong!")
}

func Oi(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Olá, %s! 👋", m.Author.Mention()))
}

func Ajuda(s *discordgo.Session, m *discordgo.MessageCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "📋 Comandos Disponíveis",
		Description: "Aqui estão todos os comandos que você pode usar:",
		Color:       0x9b59b6,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🎵 Música",
				Value:  "`!play <link/nome>` - Toca música do YouTube\n`!skip` - Pula música atual\n`!stop` - Para a música e limpa fila\n`!fila` - Mostra fila de músicas",
				Inline: false,
			},
			{
				Name:   "🎮 Geral",
				Value:  "`!ping` - Testa latência\n`!oi` - Saudação\n`!info` - Informações do servidor\n`!dado [lados]` - Rola um dado\n`!moeda` - Cara ou coroa",
				Inline: false,
			},
			{
				Name:   "⚙️ Admin",
				Value:  "`!config bemvindo <id_canal>` - Define canal de boas-vindas\n`!config cargo <id_cargo>` - Define cargo automático",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Bot criado utilizando Go por Paludett",
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

func Info(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		log.Printf("❌ Erro ao obter informações do servidor: %v", err)
		s.ChannelMessageSend(m.ChannelID, "❌ Erro ao obter informações do servidor.")
		return
	}

	// Obtém data de criação do servidor (a partir do ID)
	// Discord IDs contêm timestamp
	createdAt := getDiscordTimestamp(guild.ID)

	embed := &discordgo.MessageEmbed{
		Title: "ℹ️ Informações do Servidor",
		Color: 0x3498db,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Nome",
				Value:  guild.Name,
				Inline: true,
			},
			{
				Name:   "ID",
				Value:  guild.ID,
				Inline: true,
			},
			{
				Name:   "Membros",
				Value:  fmt.Sprintf("%d", guild.ApproximateMemberCount),
				Inline: true,
			},
			{
				Name:   "Criado em",
				Value:  createdAt.Format("02/01/2006 15:04"),
				Inline: true,
			},
			{
				Name:   "Canais de Texto",
				Value:  fmt.Sprintf("%d", len(guild.Channels)),
				Inline: true,
			},
			{
				Name:   "Cargos",
				Value:  fmt.Sprintf("%d", len(guild.Roles)),
				Inline: true,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: guild.IconURL(""),
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

func Dado(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	rand.Seed(time.Now().UnixNano())

	lados := 6
	if len(args) > 1 {
		if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
			lados = n
		}
	}

	resultado := rand.Intn(lados) + 1
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🎲 Você rolou um dado de %d lados e tirou: **%d**", lados, resultado))
}

func Moeda(s *discordgo.Session, m *discordgo.MessageCreate) {
	rand.Seed(time.Now().UnixNano())

	resultado := "Cara"
	emoji := "🪙"
	if rand.Intn(2) == 0 {
		resultado = "Coroa"
		emoji = "👑"
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s A moeda caiu em: **%s**", emoji, resultado))
}

// getDiscordTimestamp extrai o timestamp de um ID Discord
func getDiscordTimestamp(id string) time.Time {
	// IDs Discord usam Unix timestamp em milissegundos nos primeiros 42 bits
	var snowflake int64
	fmt.Sscanf(id, "%d", &snowflake)
	timestamp := (snowflake >> 22) + 1420070400000
	return time.UnixMilli(timestamp)
}
