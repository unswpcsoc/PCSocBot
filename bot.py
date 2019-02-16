#!/usr/bin/env python3

import commands
#from commands.highnoon import high_noon, HIGH_NOON_CHANNEL
#from commands.leaderboard import leaderboard, LEADERBOARD_CHANNEL
from commands.twitch import twitch, TWITCH_CHANNEL
from configstartup import config
from commands.report import report, REPORT_CHANNEL
from commands.emoji import emojistats
from commands.birthday import update_birthday

import asyncio
import json
import os
import sys
import discord

client = discord.Client()
high_noon_channel = None
report_channel = None

DEFAULT_PRESENCE = "!helpme"
err = """OOPSIE WOOPSIE!! Uwu We made a fucky wucky!! A wittle fucko boingo!
The code monkeys at our headquarters are working VEWY HAWD to fix this!"""

@client.event
async def on_ready():
    # Set game by CLA or default
    presence = sys.argv[1] if len(sys.argv) == 2 else DEFAULT_PRESENCE

    print('------')
    print('Logged in as')
    print(client.user.name)
    print(client.user.id)
    print("Playing " + presence)

    if not discord.opus.is_loaded():
        discord.opus.load_opus()

    if discord.opus.is_loaded():
        print("Opus Loaded")
    else:
        print("Opus not Loaded!")

    print('------')

    global report_channel

    await client.change_presence(game=discord.Game(name=presence))

    # Birthday checking!
    asyncio.ensure_future(update_birthday(client))

    for channel in client.get_all_channels():

        # if channel.name == HIGH_NOON_CHANNEL:
            # await high_noon(client, channel)

        # if channel.name == LEADERBOARD_CHANNEL:
            #asyncio.ensure_future(leaderboard(client, channel))

        if channel.id == TWITCH_CHANNEL:
            asyncio.ensure_future(twitch(client, channel))

        if channel.id == REPORT_CHANNEL:
            report_channel = channel

@client.event
async def on_message(message):
    try:
        if report_channel and await report(client, report_channel, message):
            return

        await emojistats(message)

        if message.content.startswith(commands.PREFIX) and message.author != client.user:
            args = '\\n '.join(message.content[1:].splitlines()).split()
            if args:
                cls, args = commands.Helpme.find_command(args)
                if not cls.disabled:
                    # Command is enabled
                    output = await cls(client, message).init(*args)
                    if isinstance(output, discord.Embed):
                        await client.send_message(message.channel, embed=output)
                    elif output is not None:
                        if isinstance(output, list):
                            for msg in output:
                                await client.send_message(message.channel, msg)
                        else:
                            await client.send_message(message.channel, output)
    except discord.errors.HTTPException as e:
        print(e)
        await client.send_message(message.channel, err)

client.run(config['KEYS'].get('DiscordToken'))
