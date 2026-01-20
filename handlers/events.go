package handlers

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/paludett/bot_discord/config"
	"github.com/paludett/bot_discord/music"
)

func Ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Printf("✅ Bot conectado como %s#%s\n", event.User.Username, event.User.Discriminator)
	s.UpdateGameStatus(0, "🎵 !ajuda para ver comandos")
}

func GuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	// Mensagem de boas-vindas
	if config.AppConfig.BemVindoChannelID != "" {
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
		s.ChannelMessageSendEmbed(config.AppConfig.BemVindoChannelID, embed)
		fmt.Printf("👋 Mensagem de boas-vindas enviada para %s\n", m.User.Username)
	}

	// Cargo automático
	if config.AppConfig.CargoAutoID != "" {
		err := s.GuildMemberRoleAdd(m.GuildID, m.User.ID, config.AppConfig.CargoAutoID)
		if err != nil {
			log.Printf("❌ Erro ao adicionar cargo: %v", err)
		} else {
			fmt.Printf("✅ Cargo adicionado para %s\n", m.User.Username)
		}
	}
}

func MessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == s.State.User.ID {
		return
	}

	music.HandleVoteReaction(s, r.GuildID, r.ChannelID, r.MessageID, r.UserID, r.Emoji.Name, true)
}

func MessageReactionRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	if r.UserID == s.State.User.ID {
		return
	}

	music.HandleVoteReaction(s, r.GuildID, r.ChannelID, r.MessageID, r.UserID, r.Emoji.Name, false)
}
