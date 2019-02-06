from commands.base import Command
from commands.help import Helpme
import commands
import configparser


class ConfigFileNotFound(FileNotFoundError):
    pass


class MissingToken(Exception):
    pass


class InvalidCommand(Exception):
    pass


config = configparser.ConfigParser()
if not config.read('config/config.ini'):
    raise ConfigFileNotFound(
        "Can't find config file! Did you create a file config.ini with similar structure to default.ini?")

if not config['KEYS'].get('DiscordToken'):
    raise MissingToken(
        "Can't find Discord token! Did you add it to the configuration file?")


def disable_commands():
    # Store list of all commands that depend on a certain key
    dependencies = {'YouTube': ['m'],
                    'TwitchClientID': ['twitch']}

    # Disable specific commands if their key doesn't exist
    blocked = []
    for key, val in dependencies.items():
        if not config['KEYS'].get(key):
            blocked.extend(val)

    # Disable commands specified in the config
    blocked_commands = config['BLOCKED'].get('blockedCommands')
    if blocked_commands:
        blocked.extend(blocked_commands.split(','))

    for comm in blocked:
        command_cls, _ = Helpme.find_command(comm.split())
        if command_cls == Command:
            # There's a command in the config file that isn't a command
            raise InvalidCommand(
                f"Error: {comm} is not a command, thus it can't be disabled")
        command_cls.disabled = True
