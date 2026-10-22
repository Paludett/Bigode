package music

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"github.com/thomas-vilte/dave-go/session"
)

// opusProvider entrega os frames Opus do ffmpeg para o AudioSender do disgo,
// que chama ProvideOpusFrame a cada 20 ms.
type opusProvider struct {
	guildID snowflake.ID
	ogg     *OggOpusReader
	dave    *session.Session
	done    chan struct{}
	once    sync.Once
	started bool
}

func (p *opusProvider) finish() {
	p.once.Do(func() { close(p.done) })
}

func (p *opusProvider) ProvideOpusFrame() ([]byte, error) {
	select {
	case <-p.done:
		return nil, io.EOF
	default:
	}

	// Durante o handshake DAVE o áudio sairia sem criptografia e seria descartado
	if p.dave != nil && p.dave.ShouldHoldFrames() {
		return nil, nil
	}

	frame, err := p.ogg.Next()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			log.Printf("❌ [voz %s] erro ao ler áudio: %v", p.guildID, err)
		}
		p.finish()
		return nil, io.EOF
	}
	if !p.started {
		p.started = true
		log.Printf("🎶 [voz %s] começou a enviar áudio", p.guildID)
	}
	return frame, nil
}

func (p *opusProvider) Close() {
	p.finish()
}

// PlayAudio toca a fila do servidor até acabar e então sai do canal de voz.
func PlayAudio(c *bot.Client, guildID, channelID, textChannelID snowflake.ID) {
	conn, dave, err := joinVoice(c, guildID, channelID)
	if err != nil {
		log.Printf("❌ [voz %s] erro ao conectar ao canal de voz: %v", guildID, err)
		FilaMutex.Lock()
		FilaMusica[guildID] = nil
		FilaMutex.Unlock()
		return
	}

	for {
		FilaMutex.Lock()
		if len(FilaMusica[guildID]) == 0 {
			FilaMutex.Unlock()
			break
		}
		musica := FilaMusica[guildID][0]
		FilaMutex.Unlock()

		log.Printf("▶️ [voz %s] tocando: %s", guildID, musica.Titulo)
		provider := &opusProvider{guildID: guildID, dave: dave, done: make(chan struct{})}
		streamOne(guildID, conn, musica, provider)

		FilaMutex.Lock()
		if len(FilaMusica[guildID]) > 0 {
			FilaMusica[guildID] = FilaMusica[guildID][1:]
		}
		FilaMutex.Unlock()
	}

	time.Sleep(time.Second * 2)

	FilaMutex.Lock()
	vazia := len(FilaMusica[guildID]) == 0
	FilaMutex.Unlock()
	if vazia {
		Disconnect(c, guildID)
	}
}

// streamOne roda yt-dlp -> ffmpeg (Opus em Ogg) e envia os frames até a música acabar ou ser interrompida.
func streamOne(guildID snowflake.ID, conn voice.Conn, musica *Musica, provider *opusProvider) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ytdlp := exec.CommandContext(ctx, "yt-dlp", "-f", "bestaudio", "-o", "-", musica.URL)
	ytdlp.Stderr = os.Stderr
	ytdlpOut, err := ytdlp.StdoutPipe()
	if err != nil {
		log.Printf("❌ [voz %s] erro ao criar pipe: %v", guildID, err)
		return
	}

	ffmpeg := exec.CommandContext(ctx, "ffmpeg", "-loglevel", "error", "-i", "pipe:0", "-vn",
		"-c:a", "libopus", "-b:a", "128k", "-ar", "48000", "-ac", "2",
		"-frame_duration", "20", "-f", "ogg", "pipe:1")
	ffmpeg.Stdin = ytdlpOut
	ffmpeg.Stderr = os.Stderr
	ffmpegOut, err := ffmpeg.StdoutPipe()
	if err != nil {
		log.Printf("❌ [voz %s] erro ao criar pipe ffmpeg: %v", guildID, err)
		return
	}

	if err := ytdlp.Start(); err != nil {
		log.Printf("❌ [voz %s] erro ao iniciar yt-dlp: %v", guildID, err)
		return
	}
	if err := ffmpeg.Start(); err != nil {
		log.Printf("❌ [voz %s] erro ao iniciar ffmpeg: %v", guildID, err)
		cancel()
		_ = ytdlp.Wait()
		return
	}

	provider.ogg = NewOggOpusReader(ffmpegOut)

	session := &StreamSession{cancel: cancel}
	FilaMutex.Lock()
	StreamingSessions[guildID] = session
	FilaMutex.Unlock()
	defer func() {
		FilaMutex.Lock()
		if StreamingSessions[guildID] == session {
			delete(StreamingSessions, guildID)
		}
		FilaMutex.Unlock()
	}()

	conn.SetOpusFrameProvider(provider)

	select {
	case <-provider.done:
	case <-ctx.Done():
		log.Printf("⏭️ [voz %s] música interrompida", guildID)
	}

	// Encerra os dois processos e espera os dois, para não deixar zumbis
	cancel()
	conn.SetOpusFrameProvider(nil)
	_ = ffmpeg.Wait()
	_ = ytdlp.Wait()
}
