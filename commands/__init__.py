from commands.help import *
from commands.birthday import *
from commands.base import *
from commands.pingpong import *
from commands.tags import *
from commands.sounds import *
from commands.someone import *
from commands.archive import *
# TODO Fix leaderboard
#from commands.leaderboard import *
from commands.playing import *
from commands.crashme import *
from commands.music import *
from commands.poll import *
from commands.twitch import *
from commands.mod import *
from commands.quote import *
from commands.auto import *
from commands.emoji import *
from commands.piglatin import *
from commands.wish import *
from commands.roles import *
from commands.branch import *
from commands.handbook import *

# Disable commands based on config
from configstartup import disable_commands
disable_commands()
