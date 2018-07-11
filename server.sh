#!/bin/sh
# Restart the bot every 24h to avoid weird crashes

while true
do
    echo
    echo "Updating bot..."
    git pull

    echo
    echo "Starting bot..."
    timeout 24h python3 bot.py
done
