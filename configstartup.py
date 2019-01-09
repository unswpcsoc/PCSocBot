import configparser

class ConfigFileNotFound(FileNotFoundError):
    pass


class MissingToken(Exception):
    pass

config = configparser.ConfigParser()

def toggle_command(key, command):
    if not config['KEYS'].get(key):
        config['COMMANDS'][command] = 'off'

# Store list of all commands that depend on a certain key
dependencies = {'YouTube': ['M'],
                'TwitchClientID': ['Twitch']}

if not config.read('config/config.ini'):
    raise ConfigFileNotFound(
        "Can't find config file! Did you create a file config.ini with similar structure to default.ini?")

if not config['KEYS'].get('DiscordToken'):
    raise MissingToken(
        "Can't find Discord token! Did you add it to the configuration file?")

# Disable specific commands if their key doesn't exist
for key, val in dependencies.items():
    for command in val:
        toggle_command(key, command)