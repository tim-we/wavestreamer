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
