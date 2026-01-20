package music

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// VoteSession armazena uma votação ativa
type VoteSession struct {
	GuildID     string             // ID do servidor
	ChannelID   string             // ID do canal de texto (para enviar mensagens)
	MessageID   string             // ID da mensagem de votação
	VoiceUsers  []string           // IDs dos usuários no canal de voz
	Votes       map[string]bool    // Map de userID -> voto (true = sim, false = não)
	VoteMutex   sync.Mutex         // Proteção de concorrência
	Action      string             // "skip" ou "clear"
	StartTime   time.Time          // Quando a votação começou
	TimeoutTime time.Duration      // Tempo para votação expirar
	Session     *discordgo.Session // Session para enviar mensagens
}

var (
	ActiveVotes = make(map[string]*VoteSession) // Map guildID -> VoteSession
	VoteMutex   sync.Mutex                      // Proteção do mapa
)

// NewVoteSession cria uma nova sessão de votação
func NewVoteSession(s *discordgo.Session, guildID, channelID string, users []string, action string) *VoteSession {
	return &VoteSession{
		GuildID:     guildID,
		ChannelID:   channelID,
		VoiceUsers:  users,
		Votes:       make(map[string]bool),
		Action:      action,
		StartTime:   time.Now(),
		TimeoutTime: 30 * time.Second,
		Session:     s,
	}
}

// AddVote adiciona um voto à sessão
func (v *VoteSession) AddVote(userID string, vote bool) bool {
	v.VoteMutex.Lock()
	defer v.VoteMutex.Unlock()

	if time.Since(v.StartTime) > v.TimeoutTime {
		return false // Votação expirou
	}

	// Verifica se o usuário está no canal de voz
	userInVoice := false
	for _, id := range v.VoiceUsers {
		if id == userID {
			userInVoice = true
			break
		}
	}

	if !userInVoice {
		return false // Usuário não está no canal de voz
	}

	v.Votes[userID] = vote
	return true
}

// GetResult retorna se a votação foi aprovada (>= 50%)
func (v *VoteSession) GetResult() (approved bool, yesVotes int, totalVotes int) {
	v.VoteMutex.Lock()
	defer v.VoteMutex.Unlock()

	// Conta votos sim
	yesVotes = 0
	totalVotes = len(v.VoiceUsers)

	for _, vote := range v.Votes {
		if vote {
			yesVotes++
		}
	}

	// Precisa de pelo menos 50% de aprovação
	if totalVotes == 0 {
		return false, 0, 0
	}

	percentage := (yesVotes * 100) / totalVotes
	approved = percentage >= 50

	return approved, yesVotes, totalVotes
}

// StartVote inicia uma votação e aguarda respostas
func StartVote(s *discordgo.Session, guildID, channelID string, users []string, action string) (*VoteSession, bool, int, int) {
	VoteMutex.Lock()

	// Verifica se já existe votação ativa
	if _, exists := ActiveVotes[guildID]; exists {
		VoteMutex.Unlock()
		return nil, false, 0, 0 // Já existe votação ativa
	}

	voteSession := NewVoteSession(s, guildID, channelID, users, action)
	ActiveVotes[guildID] = voteSession
	VoteMutex.Unlock()

	// Envia mensagem de votação
	var actionText string
	if action == "skip" {
		actionText = "pular"
	} else {
		actionText = "limpar a fila"
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🗳️ Votação em Andamento",
		Description: fmt.Sprintf("Vote com ✅ para %s ou ❌ para cancelar", actionText),
		Color:       0xFF9900,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Ação",
				Value:  actionText,
				Inline: true,
			},
			{
				Name:   "Tempo",
				Value:  "30 segundos",
				Inline: true,
			},
			{
				Name:   "Votos Necessários",
				Value:  fmt.Sprintf("%d de %d (50%%)", (len(users)/2)+1, len(users)),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Reaja com ✅ ou ❌ para votar",
		},
	}

	msg, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err == nil {
		voteSession.MessageID = msg.ID
		// Adiciona reações automáticas
		s.MessageReactionAdd(channelID, msg.ID, "✅")
		s.MessageReactionAdd(channelID, msg.ID, "❌")
	}

	// Aguarda 30 segundos
	time.Sleep(30 * time.Second)

	// Remove votação ativa
	VoteMutex.Lock()
	delete(ActiveVotes, guildID)
	VoteMutex.Unlock()

	// Retorna resultado
	approved, yesVotes, totalVotes := voteSession.GetResult()
	return voteSession, approved, yesVotes, totalVotes
}

// HandleVoteReaction processa reações de votação
func HandleVoteReaction(s *discordgo.Session, guildID string, channelID string, messageID string, userID string, emoji string, isAdded bool) {
	VoteMutex.Lock()
	voteSession, exists := ActiveVotes[guildID]
	VoteMutex.Unlock()

	if !exists {
		return
	}

	// Apenas processar se for uma reação adicionada
	if !isAdded {
		return
	}

	var vote bool
	switch emoji {
	case "✅":
		vote = true
	case "❌":
		vote = false
	default:
		return
	}

	voteSession.AddVote(userID, vote)
}
