package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"math/rand"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	Token string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env")
	}
	Token = os.Getenv("DISCORD_TOKEN")
	if Token == "" {
		log.Fatal("Token do Discord não encontrado. Configure a variável de ambiente DISCORD_TOKEN")
	}
}

func main() {
	// Cria uma nova sessão do Discord
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal("Erro ao criar sessão do Discord:", err)
	}

	// Registra handlers para eventos
	dg.AddHandler(messageCreate)
	dg.AddHandler(ready)

	// Indica quais intents (eventos) queremos receber
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	// Abre conexão com o Discord
	err = dg.Open()
	if err != nil {
		log.Fatal("Erro ao abrir conexão:", err)
	}
	defer dg.Close()

	fmt.Println("Bot está rodando! Pressione CTRL+C para sair.")

	// Aguarda sinal de interrupção
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

// Handler quando o bot está pronto
func ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Printf("Bot conectado como %s#%s\n", event.User.Username, event.User.Discriminator)

	// Define status do bot
	s.UpdateGameStatus(0, "Use !ajuda para ver comandos")
}

// Handler para mensagens
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignora mensagens do próprio bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Verifica se a mensagem começa com o prefixo
	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	// Remove o prefixo e divide em comando e argumentos
	content := strings.TrimPrefix(m.Content, "!")
	args := strings.Fields(content)

	if len(args) == 0 {
		return
	}

	comando := strings.ToLower(args[0])

	// Roteamento de comandos
	switch comando {
	case "ping":
		cmdPing(s, m)
	case "oi", "ola":
		cmdOi(s, m)
	case "ajuda", "help":
		cmdAjuda(s, m)
	case "info":
		cmdInfo(s, m)
	case "dado":
		cmdDado(s, m, args)
	case "moeda":
		cmdMoeda(s, m)
	default:
		s.ChannelMessageSend(m.ChannelID, "Comando não reconhecido! Use `!ajuda` para ver os comandos disponíveis.")
	}
}

// Comando: !ping
func cmdPing(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "🏓 Pong!")
}

// Comando: !oi
func cmdOi(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Olá, %s! 👋", m.Author.Mention()))
}

// Comando: !ajuda
func cmdAjuda(s *discordgo.Session, m *discordgo.MessageCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "📋 Comandos Disponíveis",
		Description: "Aqui estão todos os comandos que você pode usar:",
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "!ping",
				Value:  "Verifica se o bot está respondendo",
				Inline: false,
			},
			{
				Name:   "!oi",
				Value:  "O bot te cumprimenta",
				Inline: false,
			},
			{
				Name:   "!info",
				Value:  "Mostra informações sobre o servidor",
				Inline: false,
			},
			{
				Name:   "!dado [lados]",
				Value:  "Rola um dado (padrão: 6 lados)",
				Inline: false,
			},
			{
				Name:   "!moeda",
				Value:  "Joga uma moeda (cara ou coroa)",
				Inline: false,
			},
			{
				Name:   "!ajuda",
				Value:  "Mostra esta mensagem",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Bot criado utilizando go por Paludett",
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// Comando: !info
func cmdInfo(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Erro ao obter informações do servidor.")
		return
	}

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
				Name:   "Membros",
				Value:  fmt.Sprintf("%d", guild.MemberCount),
				Inline: true,
			},
			{
				Name:   "Criado em",
				Value:  guild.JoinedAt.Format("02/01/2006"),
				Inline: true,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: guild.IconURL(""),
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// Comando: !dado
func cmdDado(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
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

// Comando: !moeda
func cmdMoeda(s *discordgo.Session, m *discordgo.MessageCreate) {

	rand.Seed(time.Now().UnixNano())

	resultado := "Cara"
	emoji := "🪙"
	if rand.Intn(2) == 0 {
		resultado = "Coroa"
		emoji = "👑"
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s A moeda caiu em: **%s**", emoji, resultado))
}
