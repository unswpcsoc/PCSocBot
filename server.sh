#!/bin/sh
# Restart the bot everytime it crashes

while true
do
    echo
    echo "Starting bot..."
    python3 bot.py

    sleep 1
done
