package player

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
)

type DecodingProcess struct {
	filepath string
	cmd      *exec.Cmd
	stdout   io.ReadCloser
	reader   *bufio.Reader
}

func NewDecodingProcess(filepath string) DecodingProcess {
	cmd := exec.Command(
		"ffmpeg",
		"-i", filepath, // input file
		"-f", "s16le", // output format (signed 16bit integer little endian)
		"-ac", strconv.Itoa(CHANNELS),
		"-ar", strconv.Itoa(SAMPLE_RATE),
		"pipe:1", // output to stdout
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	return DecodingProcess{
		filepath,
		cmd,
		stdout,
		nil,
	}
}

func (process *DecodingProcess) StartDecoding() {
	if process.reader != nil {
		log.Fatal(fmt.Printf("Already started decoding %s.", process.filepath))
	}

	if err := process.cmd.Start(); err != nil {
		log.Fatal(err)
	}

	process.reader = bufio.NewReader(process.stdout)
}

func (process *DecodingProcess) Close() {
	process.stdout.Close()
}

// readFrame reads two 16-bit samples from the PCM stream
func (process *DecodingProcess) ReadFrame() (error, float32, float32) {
	var left, right int16
	if err := binary.Read(process.reader, binary.LittleEndian, &left); err != nil {
		return err, 0, 0
	}
	if err := binary.Read(process.reader, binary.LittleEndian, &right); err != nil {
		return err, 0, 0
	}
	return nil, float32(left) / 32768.0, float32(right) / 32768.0
}

func (process *DecodingProcess) WaitForExit() {
	if waitErr := process.cmd.Wait(); waitErr != nil {
		log.Fatal(waitErr)
	}
}
