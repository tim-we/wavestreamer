package decoder

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/tim-we/wavestreamer/config"
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
		"-ac", strconv.Itoa(config.CHANNELS),
		"-ar", strconv.Itoa(config.SAMPLE_RATE),
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

func (process *DecodingProcess) StartDecoding() error {
	if process.reader != nil {
		return fmt.Errorf("Decoding for %s has already started.", process.filepath)
	}

	if err := process.cmd.Start(); err != nil {
		return err
	}

	process.reader = bufio.NewReader(process.stdout)

	return nil
}

func (process *DecodingProcess) Close() {
	process.stdout.Close()

	if process.cmd.Process == nil {
		log.Println("Process closed before its started.")
		return
	}

	if process.cmd.ProcessState == nil || !process.cmd.ProcessState.Exited() {
		if err := process.cmd.Process.Kill(); err != nil && err != os.ErrProcessDone {
			log.Printf("Failed to kill process: %v", err)
		}
	}

	if waitErr := process.cmd.Wait(); waitErr != nil && waitErr != os.ErrProcessDone {
		// Only log unexpected errors (e.g. not "signal: killed")
		if exitErr, ok := waitErr.(*exec.ExitError); ok && exitErr.ExitCode() != -1 {
			log.Printf("Process exited with error: %v", waitErr)
		}
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
