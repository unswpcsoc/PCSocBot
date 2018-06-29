import urllib.request
import json
import asyncio
from helpers import at

# TODO: make sure you change this before putting on actual PCSoc server
LEADERBOARD_CHANNEL = 'general'
MEE6_URL = 'https://mee6.xyz/api/plugins/levels/leaderboard/462063414408249374'
# LEADERBOARD_CHANNEL = 'bot-testing'
# MEE6_URL = 'https://mee6.xyz/api/plugins/levels/leaderboard/157263595128881153'
SPOOF_AGENT = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36\
    (KHTML, like Gecko) Chrome/66.0.3359.181 Safari/537.36'
DATA_FILE = 'files/leaderboard.json'
MILESTONES = [10, 25, 50]
SLEEP_INTERVAL = 1

async def leaderboard(client, channel):
    req = urllib.request.Request(
        MEE6_URL, 
        data=None, 
        headers={
            'User-Agent': SPOOF_AGENT
        }
    )
    # Event Loop
    while True:
        # Make MEE6 API Request
        res = urllib.request.urlopen(req)
        data = json.loads(res.read().decode('utf-8'))
        new_list = [player['id'] for player in data['players']]
        new_positions = invert(new_list)

        # Compare with existing data
        alerts = []
        try:
            with open(DATA_FILE, 'r') as previous:
                previous_positions = json.load(previous)
                diff_positions = dict()
                for user, value in new_positions.items():
                    rank = value + 1
                    try:
                        if user not in previous_positions or previous_positions[user] > value:
                            prev = new_list[rank]
                            if rank == 1:
                                alerts.append("{} has just taken the #1 spot from {}.".format(
                                    at(user), at(prev)))
                            elif rank in MILESTONES:
                                alerts.append("{} has just entered the top {}, kicking out {}.".format(
                                    at(user), rank, at(prev)))
                            else:
                                alerts.append("{} has overtaken {} and is now rank #{}.".format(
                                    at(user), at(prev), rank))
                    except IndexError:
                        alerts.append("{} has just entered the top 100.".format(at(user)))
        except FileNotFoundError:
            pass
        
        # Save new data
        with open(DATA_FILE, 'w') as new:
            json.dump(new_positions, new)

        # Output any alerts
        for alert in alerts:
            await client.send_message(channel, alert)

        # Sleep
        await asyncio.sleep(SLEEP_INTERVAL)


def invert(array):
    dictionary = dict()
    for index, value in enumerate(array):
        dictionary[value] = index
    return dictionary
