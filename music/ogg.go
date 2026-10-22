package music

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

// OggOpusReader extrai os pacotes Opus de um stream Ogg (o que o ffmpeg gera
// com `-c:a libopus -f ogg`). Os dois primeiros pacotes (OpusHead e OpusTags)
// são descartados; cada pacote seguinte é um frame Opus pronto para envio.
type OggOpusReader struct {
	r       *bufio.Reader
	queue   [][]byte // pacotes completos já lidos, ainda não entregues
	partial []byte   // pacote que continua na próxima página
}

func NewOggOpusReader(r io.Reader) *OggOpusReader {
	return &OggOpusReader{r: bufio.NewReaderSize(r, 64*1024)}
}

// Next devolve o próximo frame Opus. Retorna io.EOF no fim do stream.
func (o *OggOpusReader) Next() ([]byte, error) {
	for {
		for len(o.queue) > 0 {
			pkt := o.queue[0]
			o.queue = o.queue[1:]
			if len(pkt) == 0 || bytes.HasPrefix(pkt, []byte("OpusHead")) || bytes.HasPrefix(pkt, []byte("OpusTags")) {
				continue
			}
			return pkt, nil
		}
		if err := o.readPage(); err != nil {
			return nil, err
		}
	}
}

// readPage lê uma página Ogg e separa os pacotes dela.
// Layout: "OggS", versão, flags, granule(8), serial(4), seq(4), crc(4), nsegs, tabela de segmentos, dados.
func (o *OggOpusReader) readPage() error {
	var header [27]byte
	if _, err := io.ReadFull(o.r, header[:]); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return io.EOF
		}
		return err
	}
	if string(header[:4]) != "OggS" {
		return fmt.Errorf("página Ogg inválida (magic %q)", header[:4])
	}

	table := make([]byte, header[26])
	if _, err := io.ReadFull(o.r, table); err != nil {
		return err
	}

	var size int
	for _, n := range table {
		size += int(n)
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(o.r, data); err != nil {
		return err
	}

	// Segmento de 255 bytes significa que o pacote continua; menor que 255 termina o pacote.
	offset := 0
	for _, n := range table {
		o.partial = append(o.partial, data[offset:offset+int(n)]...)
		offset += int(n)
		if n < 255 {
			o.queue = append(o.queue, o.partial)
			o.partial = nil
		}
	}
	return nil
}
