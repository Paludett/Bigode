package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"github.com/paludett/bot_discord/config"
	"github.com/paludett/bot_discord/music"
)

func Ready(e *events.Ready) {
	fmt.Printf("✅ Bot conectado como %s\n", e.User.Username)
	err := e.Client().SetPresence(context.Background(), gateway.WithPlayingActivity("🎵 !ajuda para ver comandos"))
	if err != nil {
		log.Printf("❌ Erro ao definir status: %v", err)
	}
}

func GuildMemberJoin(e *events.GuildMemberJoin) {
	m := e.Member

	// Mensagem de boas-vindas
	if config.AppConfig.BemVindoChannelID != "" {
		channelID, err := snowflake.Parse(config.AppConfig.BemVindoChannelID)
		if err != nil {
			log.Printf("❌ BEM_VINDO_CHANNEL_ID inválido: %v", err)
		} else {
			now := time.Now()
			embed := discord.Embed{
				Title:       "🎉 Bem-vindo(a)!",
				Description: fmt.Sprintf("Olá %s! Seja bem-vindo(a) ao servidor!\n\nDivirta-se e respeite as regras! 🎮", m.User.Mention()),
				Color:       0x00ff00,
				Thumbnail:   &discord.EmbedResource{URL: m.EffectiveAvatarURL()},
				Footer:      &discord.EmbedFooter{Text: "Aproveite sua estadia!"},
				Timestamp:   &now,
			}
			_, err = e.Client().Rest.CreateMessage(channelID, discord.NewMessageCreate().WithEmbeds(embed))
			if err != nil {
				log.Printf("❌ Erro ao enviar boas-vindas: %v", err)
			} else {
				fmt.Printf("👋 Mensagem de boas-vindas enviada para %s\n", m.User.Username)
			}
		}
	}

	// Cargo automático
	if config.AppConfig.CargoAutoID != "" {
		roleID, err := snowflake.Parse(config.AppConfig.CargoAutoID)
		if err != nil {
			log.Printf("❌ CARGO_AUTO_ID inválido: %v", err)
			return
		}
		err = e.Client().Rest.AddMemberRole(e.GuildID, m.User.ID, roleID)
		if err != nil {
			log.Printf("❌ Erro ao adicionar cargo: %v", err)
		} else {
			fmt.Printf("✅ Cargo adicionado para %s\n", m.User.Username)
		}
	}
}

func MessageReactionAdd(e *events.GuildMessageReactionAdd) {
	if e.UserID == e.Client().ID() || e.Emoji.Name == nil {
		return
	}

	music.HandleVoteReaction(e.GuildID, e.UserID, *e.Emoji.Name, true)
}

func MessageReactionRemove(e *events.GuildMessageReactionRemove) {
	if e.UserID == e.Client().ID() || e.Emoji.Name == nil {
		return
	}

	music.HandleVoteReaction(e.GuildID, e.UserID, *e.Emoji.Name, false)
}
