package commands

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func Ping(m *events.MessageCreate) {
	send(m, "🏓 Pong!")
}

func Oi(m *events.MessageCreate) {
	send(m, fmt.Sprintf("Olá, %s! 👋", m.Message.Author.Mention()))
}

func Ajuda(m *events.MessageCreate) {
	embed := discord.Embed{
		Title:       "📋 Comandos Disponíveis",
		Description: "Aqui estão todos os comandos que você pode usar:",
		Color:       0x9b59b6,
		Fields: []discord.EmbedField{
			{
				Name:   "🎵 Música",
				Value:  "`!play <link/nome>` - Toca música do YouTube\n`!skip` - Pula música atual\n`!stop` - Para a música e limpa fila\n`!fila` - Mostra fila de músicas",
				Inline: ptr(false),
			},
			{
				Name:   "🎮 Geral",
				Value:  "`!ping` - Testa latência\n`!oi` - Saudação\n`!info` - Informações do servidor\n`!dado [lados]` - Rola um dado\n`!moeda` - Cara ou coroa",
				Inline: ptr(false),
			},
			{
				Name:   "⚙️ Admin",
				Value:  "`!config bemvindo <id_canal>` - Define canal de boas-vindas\n`!config cargo <id_cargo>` - Define cargo automático",
				Inline: ptr(false),
			},
		},
		Footer: &discord.EmbedFooter{
			Text: "Bot criado utilizando Go por Paludett",
		},
	}

	sendEmbed(m, embed)
}

func Info(m *events.MessageCreate) {
	if m.GuildID == nil {
		return
	}
	guildID := *m.GuildID

	guild, ok := m.Client().Caches.Guild(guildID)
	if !ok {
		log.Printf("❌ Erro ao obter informações do servidor: guild %s fora do cache", guildID)
		send(m, "❌ Erro ao obter informações do servidor.")
		return
	}

	// Número aproximado de membros: usa a contagem da API; cai para o cache
	membros := guild.MemberCount
	if rg, err := m.Client().Rest.GetGuild(guildID, true); err == nil && rg.ApproximateMemberCount > 0 {
		membros = rg.ApproximateMemberCount
	}

	canais := 0
	for range m.Client().Caches.ChannelsForGuild(guildID) {
		canais++
	}
	cargos := 0
	for range m.Client().Caches.Roles(guildID) {
		cargos++
	}

	// Obtém data de criação do servidor (a partir do ID)
	// Discord IDs contêm timestamp
	createdAt := guildID.Time()

	embed := discord.Embed{
		Title: "ℹ️ Informações do Servidor",
		Color: 0x3498db,
		Fields: []discord.EmbedField{
			{Name: "Nome", Value: guild.Name, Inline: ptr(true)},
			{Name: "ID", Value: guildID.String(), Inline: ptr(true)},
			{Name: "Membros", Value: fmt.Sprintf("%d", membros), Inline: ptr(true)},
			{Name: "Criado em", Value: createdAt.Format("02/01/2006 15:04"), Inline: ptr(true)},
			{Name: "Canais de Texto", Value: fmt.Sprintf("%d", canais), Inline: ptr(true)},
			{Name: "Cargos", Value: fmt.Sprintf("%d", cargos), Inline: ptr(true)},
		},
	}
	if icon := guild.IconURL(); icon != nil {
		embed.Thumbnail = &discord.EmbedResource{URL: *icon}
	}

	sendEmbed(m, embed)
}

func Dado(m *events.MessageCreate, args []string) {
	rand.Seed(time.Now().UnixNano())

	lados := 6
	if len(args) > 1 {
		if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
			lados = n
		}
	}

	resultado := rand.Intn(lados) + 1
	send(m, fmt.Sprintf("🎲 Você rolou um dado de %d lados e tirou: **%d**", lados, resultado))
}

func Moeda(m *events.MessageCreate) {
	rand.Seed(time.Now().UnixNano())

	resultado := "Cara"
	emoji := "🪙"
	if rand.Intn(2) == 0 {
		resultado = "Coroa"
		emoji = "👑"
	}

	send(m, fmt.Sprintf("%s A moeda caiu em: **%s**", emoji, resultado))
}

func ptr[T any](v T) *T { return &v }
