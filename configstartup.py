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


def disable(command):
    # Disable the command
    command_cls, _ = commands.help.Helpme.find_command(command.split())
    if command_cls == commands.base.Command:
        # There's a command in the config file that isn't a command
        raise InvalidCommand(
            f"Error: {command} is not a command, thus it can't be disabled")
    command_cls.disabled = True


def disable_dependencies():
    # Disable commands which rely on an API key that isn't available
    dependencies = {'YouTube': 'm',
                    'TwitchClientID': 'twitch'}

    # Disable specific commands if their key doesn't exist
    for key, val in dependencies.items():
        if not config['KEYS'].get(key):
            disable(val)


def disable_config_commands():
    # Disable commands specified in the config
    blocked_commands = config['BLOCKED'].get('blockedCommands')
    if blocked_commands:
        # There are some commands to block
        for comm in blocked_commands.split(','):
            disable(comm)


def disable_commands():
    # Disable both dependency and config commands
    disable_dependencies()
    disable_config_commands()
