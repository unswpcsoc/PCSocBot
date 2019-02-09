import asyncio
from datetime import date, time, timedelta, datetime
from helpers import *

from commands import Command

#HIGH_NOON_CHANNEL = 'Gaming'


class S(Command):
    desc = "Sounds of PCSoc"


class HighNoon(S):
    desc = "McCree pays you a visit in your voice channel"

    #roles_required = ['mod', 'exec']
    async def eval(self):
        #raise CommandFailure("Unicode in channel names breaks me")
        await self.play_mp3('high_noon.mp3', 20)


async def high_noon(client, channel):
    next_noon = datetime.combine(date.today(), time(hour=12))
    if next_noon < datetime.now():
        next_noon += timedelta(days=1)
    while True:
        duration = (next_noon - datetime.now()).total_seconds()
        await asyncio.sleep(duration)
        await client.send_message(channel, "It's high noon")
        await client.send_file(channel, 'files/mccree.png')
        next_noon += timedelta(days=1)


class Water(S):
    desc = noembed("https://youtu.be/4BbSn1c0V8E>")

    #roles_required = ['mod', 'exec']
    async def eval(self):
        await self.client.send_message(self.message.channel, "https://i.imgur.com/vQ0JLpa.png")
        await self.play_mp3('water.mp3', 30, quiet=True)
        return


class Stop(S):
    desc = noembed("https://youtu.be/yo_wW08OTDU")

    #roles_required = ['mod', 'exec']
    async def eval(self):
        await self.play_mp3('stop.mp3', 100)
        return
