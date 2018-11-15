#!/bin/sh

# this stuff is required for everything
pip3 install --user discord.py

# this stuff is required for !tags
pip3 install --user pony

# this stuff is only required for voice channels and music
pip3 install --user PyNaCL mutagen youtube_dl isodate

# this stuff is required for music youtube searching
pip3 install --user --upgrade google-api-python-client
