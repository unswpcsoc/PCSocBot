from commands.base import Command
import urllib.request
import json
import asyncio
from helpers import *
from configstartup import config

LEADERBOARD_CHANNEL = config['CHANNELS'].get('Leaderboard')
MEE6_URL = 'https://mee6.xyz/api/plugins/levels/leaderboard/'
SPOOF_AGENT = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36\
    (KHTML, like Gecko) Chrome/66.0.3359.181 Safari/537.36'
DATA_FILE = config['FILES'].get('LeaderboardData')
MUTE_FILE = config['FILES'].get('LeaderboardMute')
MILESTONES = [10, 25, 50]
SLEEP_INTERVAL = 60


async def leaderboard(client, channel):
    req = urllib.request.Request(
        MEE6_URL + channel.server.id,
        data=None,
        headers={
            'User-Agent': SPOOF_AGENT
        }
    )
    # Event Loop
    while True:
        # Sleep
        await asyncio.sleep(SLEEP_INTERVAL)

        # Make MEE6 API Request
        try:
            res = urllib.request.urlopen(req)
            data = json.loads(res.read().decode('utf-8'))
            new_list = [player['id'] for player in data['players']]
            new_positions = invert(new_list)
        except urllib.error.URLError as e:
            print(timestamp() + ' LEADERBOARD: "' + str(e) + '"')
            continue

        # Compare with existing data
        alerts = []
        muted_people = []
        try:
            with open(MUTE_FILE, 'r') as f:
                muted_people = json.load(f)
        except FileNotFoundError:
            pass

        try:
            with open(DATA_FILE, 'r') as previous:
                previous_positions = json.load(previous)
                diff_positions = dict()
                for user, value in new_positions.items():
                    rank = value + 1
                    user_ping = bold((await client.get_user_info(user)).display_name) if user in muted_people else at(user)
                    try:
                        if user not in previous_positions or previous_positions[user] > value:
                            prev = new_list[rank]
                            prev_ping = bold((await client.get_user_info(prev)).display_name) if prev in muted_people else at(prev)
                            if rank == 1:
                                alerts.append("{} has just taken the #1 spot from {}.".format(
                                    user_ping, prev_ping))
                            elif rank in MILESTONES:
                                alerts.append("{} has just entered the top {}, kicking out {}.".format(
                                    user_ping, rank, prev_ping))
                            else:
                                alerts.append("{} has overtaken {} and is now rank #{}.".format(
                                    user_ping, prev_ping, rank))
                    except IndexError:
                        alerts.append(
                            "{} has just entered the top 100.".format(
                                user_ping))

                    alerts.append("`!shutup` to stop :ping:")

        except FileNotFoundError:
            pass

        # Save new data
        with open(DATA_FILE, 'w') as new:
            json.dump(new_positions, new)

        # Prevent PR disaster
        if len(alerts) > 10:
            print(timestamp() + ' LEADERBOARD: more than 10 changes to the leaderboard')
            continue

        # Output any alerts
        for alert in alerts:
            await client.send_message(channel, alert)


def invert(array):
    dictionary = dict()
    for index, value in enumerate(array):
        dictionary[value] = index
    return dictionary


class Shutup(Command):
    desc = "Toggles leaderboard notifications."

    def eval(self):
        muted_people = []
        try:
            with open(MUTE_FILE, 'r') as f:
                muted_people = json.load(f)
        except FileNotFoundError:
            print(MUTE_FILE + ' not found, creating...')
        if self.user in muted_people:
            muted_people.remove(self.user)
        else:
            muted_people.append(self.user)
        with open(MUTE_FILE, 'w') as f:
            json.dump(muted_people, f)
        return "We'll %s you about your leaderboard movements." % \
            ("no longer ping" if self.user in muted_people else "resume pinging you")
