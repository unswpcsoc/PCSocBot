const Eris = require("eris");
var Datastore = require('nedb')
  , db = new Datastore({ filename: 'users.db', autoload: true });

var request = require('request').defaults({encoding: null});

var bot = new Eris.CommandClient("token", {}, { //Insert Bot token here.
    description: "PC Enthusiasts Society Discod bot made with Eris",
    owner: "David Sison, Josh Wason",
    prefix: "!"
});


bot.on("ready", () => {
    console.log("Ready!");
});

/* =========================COMMANDS========================= */

bot.registerCommand("ping", "Pong!", {
    description: "Pong!",
    fullDescription: "This command could be used to check if the bot is up. Or entertainment when you're bored.",
    caseInsensitive: true
});

bot.registerCommand("pong", "Ping!", {
    description: "Ping!",
    fullDescription: "This command could be used to check if the bot is up. Or entertainment when you're bored.",
    caseInsensitive: true
});

bot.registerCommand("highnoon", () => {
    playHighNoon();
}, {
    caseInsensitive: true,
    requirements: {
        roleNames: ["Exec", "Moderator"]
    }
}, 10000);

var tags_CMD = bot.registerCommand("tags", "Player tag storage for the UNSW PCSoc discord server.\n\n**`!tags`** `add` __`platform/game`__ __`tag`__\n    Adds/changes a player tag with associated platform/game to the list\n**`!tags`** `remove` __`platform/game`__\n    Removes a player tag from the list\n**`!tags`** `get` __`platform/game`__\n    Returns player tag for that discord user\n**`!tags`** __`platform/game`__\n    Displays all player tags stored for platform/game\n", {
    description: "Player tag storage for the UNSW PCSoc discord server.",
    fullDescription: "This command stores user/player tags for any platform. Can be used to search up your own tags and other users/players on the server.",
    caseInsensitive: true
});

tags_CMD.registerSubcommand("add", (msg, args) => {
    if(args.length !== 2) {
        return "Invalid input";
    } else {
        additem(msg.member.id, args[0].toLowerCase(), args[1], msg, msg.member.user.username);
    }
}, {
    description: "Add tags.",
    fullDescription: "Adds a user/player tag to the bot.",
    usage: "<platform/game> <tag>"
});

tags_CMD.registerSubcommand("get", (msg, args) => {
    if(args.length !== 1) {
        return "Invalid input";
    } else {
        getitem(msg.member.id, args[0].toLowerCase(), msg, msg.member.user.username);
    }
}, {
    description: "Get tags.",
    fullDescription: "Returns your own tag for a platform/game",
    usage: "<platform/game>"
});

tags_CMD.registerSubcommand("remove", (msg, args) => {
    if(args.length !== 1) {
        return "Invalid input";
    } else {
        removeitem(msg.member.id, args[0].toLowerCase(), msg, msg.member.user.username);
    }
}, {
    description: "Removes tags.",
    fullDescription: "Removes a user/player tag to the bot.",
    usage: "<platform/game>"
});

tags_CMD.registerSubcommand("list", (msg, args) => {
    if(args.length !== 1) {
        return "Invalid input";
    } else {
        printtags(args[0].toLowerCase(), msg);
    }
}, {
    description: "List tags.",
    fullDescription: "Returns a list of user tags for a specified platform.",
    usage: "<platform/game>"
});

/* ========================================================== */

var t = setInterval(highNoon, 1000);

function highNoon() {
    let date = new Date();
    let h = date.getHours();
    let m = date.getMinutes();
    let s = date.getSeconds();

    if (h === 12 && m === 0 && s === 0) {
        request.get("http://vignette3.wikia.nocookie.net/overwatch/images/f/f3/Mccree_portrait.png", function(err, res, buffer) {
            bot.createMessage(channelID, "It's high noon", {name: 'mccree.png', file: buffer}); //Replace channelID with the ID of the text channel that you wish to use this function.
        });

        playHighNoon();
    }
}

function playHighNoon() {
    bot.joinVoiceChannel(channelID).then(connection => {
        connection.play('HighNoon.mp3');
        connection.on('end', () => {
            bot.leaveVoiceChannel(channelID);
        });
    });
}

function additem(id, app, tag, msg, username) {
    db.findOne({ _id: id }, function (err, doc) {
        if (doc === null) {
            var temp = { _id: id, "user": username };
            temp[app] = tag;
            db.insert(temp);
        } else {
            db.update({ _id: id }, { $set: { [app]: tag, user: username } } );
        }
        bot.createMessage(msg.channel.id, tag + " added as " + app + " tag for " + username);
        console.log(tag + " added as " + app + " tag for " + username);
    });
}

function getitem(id, app, msg, username) {
    db.findOne({ _id: id }, function (err, doc) {
        if (doc === null || !doc.hasOwnProperty("user")) {
            bot.createMessage(msg.channel.id, "User not found!");
        } else if (!doc.hasOwnProperty(app)) {
            bot.createMessage(msg.channel.id, "Platform/game not found!");
        } else {
            bot.createMessage(msg.channel.id, "The " + app + " tag of " + doc.user + " is " + doc[app]);
        }
    });
}

function removeitem(id, app, msg, username) {
        db.findOne({ _id: id }, function (err, doc) {
        if (doc === null) {
            bot.createMessage(msg.channel.id, "User not found!");
        } else if (!doc.hasOwnProperty(app)) {
            bot.createMessage(msg.channel.id, "Platform/game not found!");
        } else {
            db.update({ _id: id }, { $unset: { [app]: true } });
            bot.createMessage(msg.channel.id, app + " tag for " + username + " removed");
            console.log(app + " tag for " + username + " removed");
        }
    });
}

function printtags(app, msg) {
    db.find( { [app] : { $exists: true } }, function (err, docs) {
        if (docs.length === 0) {
            bot.createMessage(msg.channel.id, "Platform/game not found!");
        } else {
            var mymessage = "Tags stored for " + app + ":\n";
            for (var i = 0; i < docs.length; i++) {
                mymessage += docs[i][app] + " [" + docs[i].user + "]\n";
            }
            bot.createMessage(msg.channel.id, mymessage);
        }
    });
}

bot.connect();
