## wavestreamer

**wavestreamer** is a lightweight music playback system written in Go, designed to run continuously on a Raspberry Pi.

It is the spiritual successor to *py-radio*, rewritten from scratch for better performance, maintainability, and flexibility.

### Features

- 🎵 Plays music from a local library (any format ffmpeg can handle)
- 🌐 Optional web app to control playback (skip, pause, repeat, schedule)
- 🕒 Plays hourly news (currently supports *Tagesschau in 100 Sekunden*)
- 🧠 Simple, reliable, and built for 24/7 use on low-powered devices


## Installation on a Raspberry Pi

We assume the 64bit version of Raspberry Pi OS.
Setting up the service is optional.
It just starts the program after the Pi boots automatically.

1. For wavestreamer to run on a Raspberry Pi we need to install the following dependencies: 

    ```bash
    sudo apt update
    sudo apt install libportaudio2 ffmpeg screen
    ```

2. Copy files from `pi-files`
    - Copy `start-radio.sh` to the home folder (e.g. `/home/pi`)
    - Copy `radio-service` to `/etc/systemd/system/radio.service` and update the paths inside

3. Setup service:
    - `sudo systemctl daemon-reexec`
    - `sudo systemctl daemon-reload`
    - `sudo systemctl enable radio.service`
    - `sudo systemctl start radio.service`

## Debugging

You can access the running program with `screen -R radio`.
