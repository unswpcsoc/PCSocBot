from commands.base import Command

DESCRIPTION = "This command could be used to check if the bot is up. "
"Or entertainment when you're bored"


class Ping(Command):
    desc = f"Pong! {DESCRIPTION}."

    def eval(self):
        return "Pong!"


class Pong(Command):
    desc = f"Ping! {DESCRIPTION}."

    def eval(self):
        return "Ping!"
