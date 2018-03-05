import asyncio
from datetime import date, time, timedelta, datetime
from helpers import CommandFailure

from commands import Command

HIGH_NOON_CHANNEL = 'Gaming'


class HighNoon(Command):
    roles_required = ['mod', 'exec']
    async def eval(self):
        raise CommandFailure("Unicode in channel names breaks me")
        # await self.play_audio('high_noon.mp3')


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
