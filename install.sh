#!/bin/sh

# this stuff is required for everything
sudo pip3 install discord.py

# this stuff is required for !tags
sudo pip3 install pony

# this stuff is only required for voice channels
sudo apt install python3-setuptools python3-dev libffi-dev ffmpeg
sudo pip3 install PyNaCL mutagen youtube_dl
