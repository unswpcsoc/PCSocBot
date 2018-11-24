#!/bin/sh

# WARNING: THIS SCRIPT IS DEPRECATED, REFER TO README FOR INSTALLATION!
# This is a 'stupid' install, it installs required packages for the user

# This script ASSUMES the user is using apt-get as their package manager

test "$(whoami)" != 'root' && (echo "Please use `sudo`"; exit 1)
sudo apt-get install python3-setuptools python3-dev libffi-dev ffmpeg
pip3 install --user --upgrade discord.py pony PyNaCL mutagen youtube_dl isodate google-api-python-client requests beautifulsoup4