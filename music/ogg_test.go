package music

import (
	"errors"
	"io"
	"os/exec"
	"testing"
)

// Gera 3 s de áudio com o ffmpeg do jeito que o player faz e confere que o
// parser devolve ~150 frames Opus de 20 ms.
func TestOggOpusReader(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg não encontrado")
	}

	cmd := exec.Command(ffmpeg, "-loglevel", "error", "-f", "lavfi", "-i", "sine=frequency=440:duration=3",
		"-c:a", "libopus", "-b:a", "128k", "-ar", "48000", "-ac", "2", "-frame_duration", "20", "-f", "ogg", "pipe:1")
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	r := NewOggOpusReader(out)
	frames := 0
	for {
		pkt, err := r.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if len(pkt) == 0 || len(pkt) > 1275*3 {
			t.Fatalf("frame com tamanho estranho: %d", len(pkt))
		}
		frames++
	}
	_ = cmd.Wait()

	if frames < 145 || frames > 155 {
		t.Fatalf("esperava ~150 frames, veio %d", frames)
	}
}
