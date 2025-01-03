package player

import "testing"

func TestAddToQueue(t *testing.T) {
	// TODO: fix relative file path issue
	if ok, _ := HasFileQueued("test-audio.ogg"); ok {
		t.Error("File should not have been queued.")
	}

	insertFileIntoQueue("test-audio.ogg", false, t)

	if ok, _ := HasFileQueued("test-audio.ogg"); !ok {
		t.Error("File should have been queued.")
	}
}

func insertFileIntoQueue(file string, next bool, t *testing.T) {
	var err error

	if next {
		err = QueueAudioNext(file)
	} else {
		err = QueueAudio(file)
	}

	if err != nil {
		t.Errorf("Unexpected error while queueing file.\nFile: %v.\nError: %v", file, err)
	}
}
