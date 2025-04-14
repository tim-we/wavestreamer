package decoder

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/tim-we/wavestreamer/player"
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
		"-ac", strconv.Itoa(player.CHANNELS),
		"-ar", strconv.Itoa(player.SAMPLE_RATE),
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
		log.Fatalf("Decoding for %s has already started.", process.filepath)
	}

	if err := process.cmd.Start(); err != nil {
		log.Fatal(err)
	}

	process.reader = bufio.NewReader(process.stdout)
}

func (process *DecodingProcess) Close() {
	process.stdout.Close()

	if process.cmd.ProcessState == nil || !process.cmd.ProcessState.Exited() {
		if err := process.cmd.Process.Kill(); err != nil && err != os.ErrProcessDone {
			log.Printf("Failed to kill process: %v", err)
		}
	}

	if waitErr := process.cmd.Wait(); waitErr != nil && waitErr != os.ErrProcessDone {
		log.Printf("Error while waiting for process to exit: %v", waitErr)
	}
}

// readFrame reads two 16-bit samples from the PCM stream
func (process *DecodingProcess) ReadFrame() (float32, float32, error) {
	var left, right int16
	if err := binary.Read(process.reader, binary.LittleEndian, &left); err != nil {
		return 0, 0, err
	}
	if err := binary.Read(process.reader, binary.LittleEndian, &right); err != nil {
		return 0, 0, err
	}
	return float32(left) / 32768.0, float32(right) / 32768.0, nil
}

func (process *DecodingProcess) WaitForExit() {
	if waitErr := process.cmd.Wait(); waitErr != nil {
		log.Fatal(waitErr)
	}
}
