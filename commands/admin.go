package commands

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/paludett/bot_discord/config"
)

func Config(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Verifica permissões de administrador
	perms, err := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
	if err != nil {
		log.Printf("❌ Erro ao obter permissões: %v", err)
		s.ChannelMessageSend(m.ChannelID, "❌ Erro ao verificar permissões!")
		return
	}

	if perms&discordgo.PermissionAdministrator == 0 {
		s.ChannelMessageSend(m.ChannelID, "❌ Você precisa ser administrador para usar este comando!")
		return
	}

	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, "❌ Use:\n`!config bemvindo <id_canal>` - Define canal de boas-vindas\n`!config cargo <id_cargo>` - Define cargo automático")
		return
	}

	tipo := strings.ToLower(args[1])
	id := args[2]

	// Lê arquivo .env
	envFile, err := os.ReadFile(".env")
	var envContent string
	if err == nil {
		envContent = string(envFile)
	}

	switch tipo {
	case "bemvindo", "welcome":
		envContent = updateEnvLine(envContent, "BEM_VINDO_CHANNEL_ID", id)
		config.AppConfig.BemVindoChannelID = id
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Canal de boas-vindas definido para: <#%s>", id))

	case "cargo", "role":
		envContent = updateEnvLine(envContent, "CARGO_AUTO_ID", id)
		config.AppConfig.CargoAutoID = id
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Cargo automático definido para: <@&%s>", id))

	default:
		s.ChannelMessageSend(m.ChannelID, "❌ Tipo inválido! Use `bemvindo` ou `cargo`")
		return
	}

	// Escreve de volta no .env
	err = os.WriteFile(".env", []byte(envContent), 0644)
	if err != nil {
		log.Printf("❌ Erro ao salvar .env: %v", err)
		s.ChannelMessageSend(m.ChannelID, "⚠️ Configuração salva em memória, mas falhou ao atualizar .env")
	}
}

// updateEnvLine atualiza ou adiciona uma linha no conteúdo do .env
func updateEnvLine(content, key, value string) string {
	lines := strings.Split(content, "\n")
	found := false

	for i, line := range lines {
		if strings.HasPrefix(line, key+"=") {
			lines[i] = fmt.Sprintf("%s=\"%s\"", key, value)
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, fmt.Sprintf("%s=\"%s\"", key, value))
	}

	return strings.Join(lines, "\n")
}
