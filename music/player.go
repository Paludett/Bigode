package music

import (
	"io"
	"log"
	"os/exec"
	"time"

	"github.com/bwmarrin/discordgo"
)

func PlayAudio(s *discordgo.Session, guildID, channelID, textChannelID string) {
	FilaMutex.Lock()
	if len(FilaMusica[guildID]) == 0 {
		FilaMutex.Unlock()
		return
	}
	musica := FilaMusica[guildID][0]
	FilaMutex.Unlock()

	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		log.Printf("❌ Erro ao conectar ao canal de voz: %v", err)
		return
	}
	VoiceConnections[guildID] = vc

	cmd := exec.Command("yt-dlp", "-f", "bestaudio", "-o", "-", musica.URL)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("❌ Erro ao criar pipe: %v", err)
		vc.Disconnect()
		return
	}

	ffmpeg := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1")
	ffmpeg.Stdin = stdout
	ffmpegout, err := ffmpeg.StdoutPipe()
	if err != nil {
		log.Printf("❌ Erro ao criar pipe ffmpeg: %v", err)
		vc.Disconnect()
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("❌ Erro ao iniciar yt-dlp: %v", err)
		vc.Disconnect()
		return
	}

	if err := ffmpeg.Start(); err != nil {
		log.Printf("❌ Erro ao iniciar ffmpeg: %v", err)
		vc.Disconnect()
		return
	}

	session := &StreamSession{Cancel: make(chan bool)}
	StreamingSessions[guildID] = session

	vc.Speaking(true)
	defer vc.Speaking(false)

	buffer := make([]byte, 3840)
	for {
		select {
		case <-session.Cancel:
			cmd.Process.Kill()
			ffmpeg.Process.Kill()
			goto cleanup
		default:
			n, err := ffmpegout.Read(buffer)
			if err != nil {
				if err == io.EOF {
					goto cleanup
				}
				log.Printf("Erro ao ler áudio: %v", err)
				goto cleanup
			}
			vc.OpusSend <- buffer[:n]
		}
	}

cleanup:
	cmd.Wait()
	ffmpeg.Wait()

	FilaMutex.Lock()
	if len(FilaMusica[guildID]) > 0 {
		FilaMusica[guildID] = FilaMusica[guildID][1:]
	}
	temMais := len(FilaMusica[guildID]) > 0
	FilaMutex.Unlock()

	if temMais {
		PlayAudio(s, guildID, channelID, textChannelID)
	} else {
		time.Sleep(time.Second * 2)
		vc.Disconnect()
		delete(VoiceConnections, guildID)
		delete(StreamingSessions, guildID)
	}
}
