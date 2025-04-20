#!/bin/bash

# Only start if the screen session isn't already running
if ! screen -list | grep -q "\.radio"; then
  screen -dmS radio ./wavestreamer -d ./wc-music -news
fi
