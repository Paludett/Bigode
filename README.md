# Bot Discord Bigode

Um bot do Discord com suporte a reprodução de músicas do YouTube e outras funcionalidades como a exibição de informações do servidor! Esse bot foi criado para meu servidor privado com meus amigos e foi totalmente feito usando a linguagem Go     

## Setup

### Requisitos
- Go 1.23+
- ffmpeg
- yt-dlp

### Instalação

1. Clone ou copie o projeto

2. Configure `.env`:
```
DISCORD_TOKEN="seu_token_aqui"
BEM_VINDO_CHANNEL_ID="seu_canal_id"
CARGO_AUTO_ID="seu_cargo_id"
```

3. Instale dependências:
```bash
go mod tidy
```

4. Compile:
```bash
go build -o bot .
```

5. Execute:
```bash
./bot
```

## Arquitetura

```
bot_discord/
├── main.go                # Inicialização do bot
├── config/                # Configurações
│   └── config.go          # Carregamento de .env
├── handlers/              # Event handlers
│   └── events.go          # Eventos de Discord
├── commands/              # Comandos do bot
│   ├── router.go          # Roteador de comandos
│   ├── general.go         # Comandos gerais
│   ├── music.go           # Comandos de música
│   └── admin.go           # Comandos de admin
└── music/                 # Lógica de música
    ├── types.go           # Estruturas (Musica, StreamSession)
    ├── player.go          # Reprodutor de áudio
    ├── search.go          # Busca do YouTube
    └── voting.go          # Sistema de votação
```

## Comandos de Música

### Reprodução
- `!play <url ou nome>` / `!tocar <url ou nome>` - Reproduz uma música do YouTube
- `!skip` / `!pular` - Pula para a próxima música (requer votação se houver múltiplos usuários no voice)
- `!stop` / `!parar` - Para a reprodução e desconecta
- `!fila` / `!queue` - Mostra a fila de músicas
- `!clear` / `!limpar` - Limpa toda a fila (requer votação se houver múltiplos usuários no voice)

### Sistema de Votação 
Quando um usuário está sozinho no canal de voz, comandos como `!skip` e `!clear` executam imediatamente. Se houver múltiplos usuários, será aberta uma votação de 30 segundos:

- ✅ para sim
- ❌ para não

A ação só será executada se **50% ou mais** dos usuários votarem a favor.

### Exemplo
```
Usuário 1: !skip
Bot: Votação iniciada! Reaja com ✅ ou ❌ (30 segundos)
(Se 50%+ votar ✅) → Música pulada
(Se menos de 50% vota ✅) → Votação rejeitada
```

## Outros Comandos

- `!ping` - Verifica a latência do bot
- `!oi` / `!ola` - Resposta amigável
- `!dado` - Rola um dado (1-6)
- `!moeda` - Cara ou coroa?
- `!info` - Informações do servidor (nome, criação, membros)
- `!ajuda` / `!help` - Lista de todos os comandos

## Troubleshooting

### Música não toca
- Verifique se `ffmpeg` e `yt-dlp` estão instalados
- Verifique se o bot tem permissão para conectar em canais de voz

### Permissões não funcionam
- O bot agora usa `discordgo.PermissionAdministrator` corretamente
- Usuário deve ter a permissão "Administrador" no Discord

### Informações do servidor incorretas
- A data de criação é extraída do ID do servidor (Snowflake do Discord)
- Número de membros usa `ApproximateMemberCount` da API do Discord

### Votação não funciona
- Verifique se o bot tem permissão para gerenciar reações (Manage Reactions)
- Verifique se o bot tem a intent `IntentsGuildMessageReactions` ativada
