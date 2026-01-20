package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/paludett/bot_discord/commands"
	"github.com/paludett/bot_discord/config"
	"github.com/paludett/bot_discord/handlers"
)

func main() {
	config.Load()

	dg, err := discordgo.New("Bot " + config.AppConfig.DiscordToken)
	if err != nil {
		log.Fatal("❌ Erro ao criar sessão do Discord:", err)
	}

	dg.AddHandler(handlers.Ready)
	dg.AddHandler(handlers.GuildMemberAdd)
	dg.AddHandler(handlers.MessageReactionAdd)
	dg.AddHandler(handlers.MessageReactionRemove)
	dg.AddHandler(commands.Router)

	dg.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMessageReactions

	err = dg.Open()
	if err != nil {
		log.Fatal("❌ Erro ao abrir conexão:", err)
	}
	defer dg.Close()

	fmt.Println("🤖 Bot está rodando! Pressione CTRL+C para sair.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
