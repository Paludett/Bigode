package music

import (
	"context"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"github.com/thomas-vilte/dave-go/session"
)

var (
	// daveMu protege lastDave e daveSessions. CreateConn cria a sessão DAVE de forma
	// síncrona, então o hook abaixo roda dentro do CreateConn e dá para associar
	// a sessão ao servidor.
	daveMu       sync.Mutex
	lastDave     *session.Session
	daveSessions = make(map[snowflake.ID]*session.Session)
)

// VoiceManagerOpts configura o VoiceManager do disgo com DAVE (Go puro, sem CGO).
func VoiceManagerOpts() []voice.ManagerConfigOpt {
	return []voice.ManagerConfigOpt{
		voice.WithDaveSessionCreateFunc(session.CreateFunc(
			session.WithSessionHook(func(s *session.Session) {
				lastDave = s
			}),
		)),
	}
}

// joinVoice entra no canal de voz do servidor (ou reaproveita a conexão existente).
func joinVoice(c *bot.Client, guildID, channelID snowflake.ID) (voice.Conn, *session.Session, error) {
	daveMu.Lock()
	defer daveMu.Unlock()

	if conn := c.VoiceManager.GetConn(guildID); conn != nil {
		return conn, daveSessions[guildID], nil
	}

	lastDave = nil
	conn := c.VoiceManager.CreateConn(guildID)
	dave := lastDave
	daveSessions[guildID] = dave

	log.Printf("🔊 [voz %s] entrando no canal %s", guildID, channelID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := conn.Open(ctx, channelID, false, false); err != nil {
		delete(daveSessions, guildID)
		closeConn(conn, guildID, dave)
		return nil, nil, err
	}
	log.Printf("🔊 [voz %s] conectado ao canal %s", guildID, channelID)

	if dave != nil {
		go watchDave(guildID, dave)
	}
	return conn, dave, nil
}

// watchDave registra no log quando o handshake DAVE termina (ou não).
func watchDave(guildID snowflake.ID, dave *session.Session) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	took, err := dave.WaitReady(ctx)
	st := dave.State()
	if err != nil {
		log.Printf("⚠️ [voz %s] handshake DAVE não concluiu em 15s (protocolo=%d, época=%d): %v", guildID, st.ProtocolVersion, st.EpochID, err)
		return
	}
	log.Printf("🔐 [voz %s] handshake DAVE concluído em %s (protocolo=%d, época=%d)", guildID, took.Round(time.Millisecond), st.ProtocolVersion, st.EpochID)
}

// Disconnect sai do canal de voz do servidor. Pode ser chamado várias vezes.
func Disconnect(c *bot.Client, guildID snowflake.ID) {
	daveMu.Lock()
	conn := c.VoiceManager.GetConn(guildID)
	dave := daveSessions[guildID]
	delete(daveSessions, guildID)
	daveMu.Unlock()

	if conn == nil {
		return
	}
	closeConn(conn, guildID, dave)
}

func closeConn(conn voice.Conn, guildID snowflake.ID, dave *session.Session) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn.Close(ctx)

	if dave != nil {
		if err := dave.Close(); err != nil {
			slog.Warn("erro ao fechar sessão DAVE", slog.Any("err", err))
		}
	}
	log.Printf("🔊 [voz %s] desconectado", guildID)
}
