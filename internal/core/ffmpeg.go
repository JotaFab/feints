package core

import (
	"encoding/binary"
	"io"
	"log"
	"os/exec"

	"layeh.com/gopus"
)

// StreamFromPathToOpusChan ejecuta ffmpeg, lee PCM y envía frames Opus al canal.
// Devuelve el *exec.Cmd para control (Stop, Wait) y el canal de salida de frames.
func StreamFromPathToOpusChan(input string, opusChan chan []byte) (*exec.Cmd, error) {
	cmd := exec.Command(
		"ffmpeg",
		"-i", input, // puede ser URL o archivo local
		"-f", "s16le", // PCM crudo
		"-ar", "48000", // sample rate
		"-ac", "2", // estéreo
		"pipe:1",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Configuramos el encoder Opus
	encoder, err := gopus.NewEncoder(48000, 2, gopus.Audio)
	if err != nil {
		return nil, err
	}

	// Goroutine para procesar PCM → Opus
	go func() {
		defer close(opusChan)

		const frameSize = 960                // 20ms @ 48000Hz
		pcmBuf := make([]int16, frameSize*2) // 2 canales

		for {
			// Leer PCM crudo (16 bits LE)
			err := binary.Read(stdout, binary.LittleEndian, pcmBuf)
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}

			// Codificar a Opus
			opus, err := encoder.Encode(pcmBuf, frameSize, 960*6) // MTU seguro
			if err != nil {
				continue
			}
			log.Println("frame enviado")
			// Enviar frame
			opusChan <- opus
		}
	}()

	return cmd, nil
}
