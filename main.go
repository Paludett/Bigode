package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/paludett/bot_discord/commands"
	"github.com/paludett/bot_discord/config"
	"github.com/paludett/bot_discord/handlers"
	"github.com/paludett/bot_discord/music"
)

func main() {
	config.Load()

	client, err := disgo.New(config.AppConfig.DiscordToken,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentMessageContent,
				gateway.IntentGuildMembers,
				gateway.IntentGuildVoiceStates,
				gateway.IntentGuildMessageReactions,
			),
		),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagGuilds, cache.FlagRoles, cache.FlagChannels, cache.FlagMembers, cache.FlagVoiceStates),
		),
		bot.WithVoiceManagerConfigOpts(music.VoiceManagerOpts()...),
		bot.WithEventListenerFunc(handlers.Ready),
		bot.WithEventListenerFunc(handlers.GuildMemberJoin),
		bot.WithEventListenerFunc(handlers.MessageReactionAdd),
		bot.WithEventListenerFunc(handlers.MessageReactionRemove),
		bot.WithEventListenerFunc(commands.Router),
	)
	if err != nil {
		log.Fatal("❌ Erro ao criar cliente do Discord:", err)
	}

	if err = client.OpenGateway(context.Background()); err != nil {
		log.Fatal("❌ Erro ao abrir conexão:", err)
	}

	fmt.Println("🤖 Bot está rodando! Pressione CTRL+C para sair.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client.Close(ctx)
}
