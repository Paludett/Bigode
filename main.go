package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"math/rand"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/kkdai/youtube/v2"
)

var (
	Token             string
	BemVindoChannelID string
	CargoAutoID       string
	filaMusica        = make(map[string][]*Musica)
	filaMutex         sync.Mutex
	voiceConnections  = make(map[string]*discordgo.VoiceConnection)
	streamingSessions = make(map[string]*StreamSession)
)

type Musica struct {
	Titulo string
	URL    string
	User   string
}

type StreamSession struct {
	cancel chan bool
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: .env não encontrado, usando variáveis de ambiente do ambiente de execução")
	}

	Token = os.Getenv("DISCORD_TOKEN")
	BemVindoChannelID = os.Getenv("BEM_VINDO_CHANNEL_ID")
	CargoAutoID = os.Getenv("CARGO_AUTO_ID")

	if Token == "" {
		log.Fatal("Token do Discord não encontrado. Configure a variável de ambiente DISCORD_TOKEN")
	}
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal("Erro ao criar sessão do Discord:", err)
	}

	// Registra handlers para eventos
	dg.AddHandler(messageCreate)
	dg.AddHandler(ready)
	dg.AddHandler(guildMemberAdd)

	// Indica quais intents (eventos) queremos receber
	dg.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildVoiceStates

	err = dg.Open()
	if err != nil {
		log.Fatal("Erro ao abrir conexão:", err)
	}
	defer dg.Close()

	fmt.Println("🤖 Bot está rodando! Pressione CTRL+C para sair.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Printf("✅ Bot conectado como %s#%s\n", event.User.Username, event.User.Discriminator)
	s.UpdateGameStatus(0, "🎵 !ajuda para ver comandos")
}

// Handler para novos membros
func guildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	// Envia mensagem de boas-vindas
	if BemVindoChannelID != "" {
		embed := &discordgo.MessageEmbed{
			Title:       "🎉 Bem-vindo(a)!",
			Description: fmt.Sprintf("Olá %s! Seja bem-vindo(a) ao servidor!\n\nDivirta-se e respeite as regras! 🎮", m.User.Mention()),
			Color:       0x00ff00,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: m.User.AvatarURL(""),
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Aproveite sua estadia!",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		s.ChannelMessageSendEmbed(BemVindoChannelID, embed)
		fmt.Printf("👋 Mensagem de boas-vindas enviada para %s\n", m.User.Username)
	}

	// Adiciona cargo automático
	if CargoAutoID != "" {
		err := s.GuildMemberRoleAdd(m.GuildID, m.User.ID, CargoAutoID)
		if err != nil {
			log.Printf("❌ Erro ao adicionar cargo: %v", err)
		} else {
			fmt.Printf("✅ Cargo adicionado para %s\n", m.User.Username)
		}
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
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
	case "play", "tocar":
		cmdPlay(s, m, args)
	case "skip", "pular":
		cmdSkip(s, m)
	case "stop", "parar":
		cmdStop(s, m)
	case "fila", "queue":
		cmdFila(s, m)
	case "config":
		cmdConfig(s, m, args)
	default:
		s.ChannelMessageSend(m.ChannelID, "❌ Comando não reconhecido! Use `!ajuda` para ver os comandos disponíveis.")
	}
}

func cmdPing(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "🏓 Pong!")
}

func cmdOi(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Olá, %s! 👋", m.Author.Mention()))
}

func cmdAjuda(s *discordgo.Session, m *discordgo.MessageCreate) {
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
				Name:   "⚙️ Configuração (Admin)",
				Value:  "`!config bemvindo <id_canal>` - Define canal de boas-vindas\n`!config cargo <id_cargo>` - Define cargo automático para novos membros",
				Inline: false,
			},
			{
				Name:   "🎮 Geral",
				Value:  "`!ping` - Testa latência\n`!oi` - Saudação\n`!info` - Informações do servidor\n`!dado [lados]` - Rola um dado\n`!moeda` - Cara ou coroa",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Bot criado utilizando go por Paludett",
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

func cmdInfo(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Erro ao obter informações do servidor.")
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

// Comando: !play
func cmdPlay(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
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

	// Verifica se é um link do YouTube
	var videoURL string
	if strings.Contains(query, "youtube.com") || strings.Contains(query, "youtu.be") {
		videoURL = query
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔍 Buscando: `%s`...", query))
		// Aqui você poderia implementar busca no YouTube
		// Por enquanto, vamos usar apenas links diretos
		s.ChannelMessageSend(m.ChannelID, "❌ Por favor, use um link direto do YouTube por enquanto!\nExemplo: `!play https://youtube.com/watch?v=...`")
		return
	}

	// Baixa informações do vídeo
	client := youtube.Client{}
	video, err := client.GetVideo(videoURL)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Erro ao obter informações do vídeo. Verifique o link!")
		return
	}

	musica := &Musica{
		Titulo: video.Title,
		URL:    videoURL,
		User:   m.Author.Username,
	}

	filaMutex.Lock()
	if filaMusica[m.GuildID] == nil {
		filaMusica[m.GuildID] = []*Musica{}
	}
	filaMusica[m.GuildID] = append(filaMusica[m.GuildID], musica)
	posicao := len(filaMusica[m.GuildID])
	filaMutex.Unlock()

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
		go playAudio(s, m.GuildID, voiceState.ChannelID, m.ChannelID)
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("➕ **%s** adicionado à fila (posição #%d)", musica.Titulo, posicao))
	}
}

// Toca áudio no canal de voz
func playAudio(s *discordgo.Session, guildID, channelID, textChannelID string) {
	filaMutex.Lock()
	if len(filaMusica[guildID]) == 0 {
		filaMutex.Unlock()
		return
	}
	musica := filaMusica[guildID][0]
	filaMutex.Unlock()

	// Conecta ao canal de voz
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		log.Printf("Erro ao conectar ao canal de voz: %v", err)
		return
	}
	voiceConnections[guildID] = vc

	// Baixa o áudio usando yt-dlp ou youtube-dl
	cmd := exec.Command("yt-dlp", "-f", "bestaudio", "-o", "-", musica.URL)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Erro ao criar pipe: %v", err)
		vc.Disconnect()
		return
	}

	// Converte para formato adequado com ffmpeg
	ffmpeg := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1")
	ffmpeg.Stdin = stdout
	ffmpegout, err := ffmpeg.StdoutPipe()
	if err != nil {
		log.Printf("Erro ao criar pipe ffmpeg: %v", err)
		vc.Disconnect()
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Erro ao iniciar yt-dlp: %v", err)
		vc.Disconnect()
		return
	}

	if err := ffmpeg.Start(); err != nil {
		log.Printf("Erro ao iniciar ffmpeg: %v", err)
		vc.Disconnect()
		return
	}

	// Cria sessão de streaming
	session := &StreamSession{cancel: make(chan bool)}
	streamingSessions[guildID] = session

	// Stream de áudio
	vc.Speaking(true)
	defer vc.Speaking(false)

	buffer := make([]byte, 3840)
	for {
		select {
		case <-session.cancel:
			cmd.Process.Kill()
			ffmpeg.Process.Kill()
			return
		default:
			n, err := ffmpegout.Read(buffer)
			if err != nil {
				if err != io.EOF {
					log.Printf("Erro ao ler áudio: %v", err)
				}
				goto cleanup
			}
			vc.OpusSend <- buffer[:n]
		}
	}

cleanup:
	cmd.Wait()
	ffmpeg.Wait()

	// Remove da fila e toca próxima
	filaMutex.Lock()
	if len(filaMusica[guildID]) > 0 {
		filaMusica[guildID] = filaMusica[guildID][1:]
	}
	temMais := len(filaMusica[guildID]) > 0
	filaMutex.Unlock()

	if temMais {
		playAudio(s, guildID, channelID, textChannelID)
	} else {
		time.Sleep(time.Second * 2)
		vc.Disconnect()
		delete(voiceConnections, guildID)
		delete(streamingSessions, guildID)
	}
}

func cmdSkip(s *discordgo.Session, m *discordgo.MessageCreate) {
	if session, ok := streamingSessions[m.GuildID]; ok {
		session.cancel <- true
		s.ChannelMessageSend(m.ChannelID, "⏭️ Música pulada!")
	} else {
		s.ChannelMessageSend(m.ChannelID, "❌ Não há música tocando!")
	}
}

func cmdStop(s *discordgo.Session, m *discordgo.MessageCreate) {
	filaMutex.Lock()
	filaMusica[m.GuildID] = []*Musica{}
	filaMutex.Unlock()

	if session, ok := streamingSessions[m.GuildID]; ok {
		session.cancel <- true
		delete(streamingSessions, m.GuildID)
	}

	if vc, ok := voiceConnections[m.GuildID]; ok {
		vc.Disconnect()
		delete(voiceConnections, m.GuildID)
	}

	s.ChannelMessageSend(m.ChannelID, "⏹️ Música parada e fila limpa!")
}

func cmdFila(s *discordgo.Session, m *discordgo.MessageCreate) {
	filaMutex.Lock()
	defer filaMutex.Unlock()

	fila := filaMusica[m.GuildID]
	if len(fila) == 0 {
		s.ChannelMessageSend(m.ChannelID, "📭 A fila está vazia!")
		return
	}

	var fields []*discordgo.MessageEmbedField

	for i, musica := range fila {
		if i == 0 {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "▶️ Tocando Agora",
				Value:  fmt.Sprintf("**%s**\nPedido por: %s", musica.Titulo, musica.User),
				Inline: false,
			})
		} else {
			if i <= 10 {
				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("#%d", i),
					Value:  fmt.Sprintf("%s\nPedido por: %s", musica.Titulo, musica.User),
					Inline: false,
				})
			}
		}
	}

	if len(fila) > 11 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "...",
			Value:  fmt.Sprintf("E mais %d músicas", len(fila)-11),
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

func cmdConfig(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Verifica permissões de administrador
	perms, err := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
	if err != nil || perms&discordgo.PermissionAdministrator == 0 {
		s.ChannelMessageSend(m.ChannelID, "❌ Você precisa ser administrador para usar este comando!")
		return
	}

	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, "❌ Use:\n`!config bemvindo <id_canal>` - Define canal de boas-vindas\n`!config cargo <id_cargo>` - Define cargo automático")
		return
	}

	tipo := strings.ToLower(args[1])
	id := args[2]

	// Atualiza arquivo .env
	envFile, err := os.ReadFile(".env")
	var envContent string
	if err == nil {
		envContent = string(envFile)
	}

	switch tipo {
	case "bemvindo", "welcome":
		BemVindoChannelID = id
		envContent = updateEnvLine(envContent, "BEM_VINDO_CHANNEL_ID", id)
		os.WriteFile(".env", []byte(envContent), 0644)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Canal de boas-vindas configurado: <#%s>", id))

	case "cargo", "role":
		CargoAutoID = id
		envContent = updateEnvLine(envContent, "CARGO_AUTO_ID", id)
		os.WriteFile(".env", []byte(envContent), 0644)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Cargo automático configurado: <@&%s>", id))

	default:
		s.ChannelMessageSend(m.ChannelID, "❌ Tipo inválido! Use `bemvindo` ou `cargo`")
	}
}

// Atualiza ou adiciona uma linha no arquivo .env
func updateEnvLine(content, key, value string) string {
	lines := strings.Split(content, "\n")
	found := false

	for i, line := range lines {
		if strings.HasPrefix(line, key+"=") {
			lines[i] = fmt.Sprintf("%s=%s", key, value)
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(lines, "\n")
}
