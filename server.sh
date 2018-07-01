#!/bin/sh

echo
echo "Updating bot..."
git pull

echo
echo "Starting bot..."

if [ $# -eq 1 ]
then
    python3 bot.py "$1"
else
    python3 bot.py
fi
