const Eris = require("eris");
var Datastore = require('nedb')
  , db = new Datastore({ filename: 'users.db', autoload: true });

var request = require('request').defaults({encoding: null});

var bot = new Eris("token"); //Insert Bot token here.

bot.on("ready", () => {
    console.log("Ready!");
});

bot.on("messageCreate", (message) => {
    var username = message.author.username;
    var id = message.author.id;
    var argv = message.content.match;

    if(message.content === "!ping") {
        bot.createMessage(message.channel.id, "Pong!");
    }

    if(message.content === "!meme") {
        request.get("http://puu.sh/s0lxS.gif", function(err, res, buffer) { //URL is a placeholder GIF
            bot.createMessage(channelID, message.author.mention, {name: 'meme.gif', file: buffer}); //Replace channelID with the ID of the channel that you want the image to be posted in. To send in same channel as the command was used in use message.channel.id
        });
    }

    if (argv !== null) {
        var argc = argv.length;
        if(argv[0] === "!tags") {
            if (argc === 1) {
                bot.createMessage(message.channel.id, "Player tag storage for the UNSW PCSoc discord server.\n\n**`!tags`** `add` __`platform/game`__ __`tag`__\n    Adds/changes a player tag with associated platform/game to the list\n**`!tags`** `remove` __`platform/game`__\n    Removes a player tag from the list\n**`!tags`** `get` __`platform/game`__\n    Returns player tag for that discord user\n**`!tags`** __`platform/game`__\n    Displays all player tags stored for platform/game\n");
            } else if(argv[1] === "add" && argc === 4) {
                var app = argv[2].toLowerCase();
                var tag = argv[3];
                additem(id, app, tag, message, username);
            } else if(argv[1] === "get" && argc === 3) {
                var app = argv[2].toLowerCase();
                getitem(id, app, message, username);
            } else if(argv[1] === "remove" && argc === 3) {
                var app = argv[2].toLowerCase();
                removeitem(id, app, message, username);
            } else if (argc === 2) {
                var app = argv[1].toLowerCase();
                printtags(app, message);
            } else {
                bot.createMessage(message.channel.id, "Unknown argument(s)...");
            }
        }
    }
});

var t = setInterval(highNoon, 1000);

function highNoon() {
    let date = new Date();
    let h = date.getHours();
    let m = date.getMinutes();
    let s = date.getSeconds();

    if (h === 12 && m === 00 && s === 0) {
        request.get("http://vignette3.wikia.nocookie.net/overwatch/images/f/f3/Mccree_portrait.png", function(err, res, buffer) {
            bot.createMessage(channelID, "It's high noon", {name: 'mccree.png', file: buffer}); //Replace channelID with the ID of the text channel that you wish to use this function.
        });
    }
}


function additem(id, app, tag, message, username) {
    db.findOne({ _id: id }, function (err, doc) {
        if (doc === null) {
            var temp = { _id: id, "user": username };
            temp[app] = tag;
            db.insert(temp);
        } else {
            db.update({ _id: id }, { $set: { [app]: tag, user: username } } );
        }
        bot.createMessage(tag + " added as " + app + " tag for " + username);
        console.log(tag + " added as " + app + " tag for " + username);
    });
}

function getitem(id, app, message, username) {
    db.findOne({ _id: id }, function (err, doc) {
        if (doc === null || !doc.hasOwnProperty("user")) {
            bot.createMessage(message.channel.id, "User not found!");
        } else if (!doc.hasOwnProperty(app)) {
            bot.createMessage(message.channel.id, "Platform/game not found!");
        } else {
            bot.createMessage(message.channel.id, "The " + app + " tag of " + doc.user + " is " + doc[app]);
        }
    });
}

function removeitem(id, app, message, username) {
        db.findOne({ _id: id }, function (err, doc) {
        if (doc === null) {
            bot.createMessage(message.channel.id, "User not found!");
        } else if (!doc.hasOwnProperty(app)) {
            bot.createMessage(message.channel.id, "Platform/game not found!");
        } else {
            db.update({ _id: id }, { $unset: { [app]: true } });
            bot.createMessage(message.channel.id, app + " tag for " + username + " removed");
            console.log(app + " tag for " + username + " removed");
        }
    });
}

function printtags(app, message) {
    db.find( { [app] : { $exists: true } }, function (err, docs) {
        if (docs.length === 0) {
            bot.createMessage(message.channel.id, "Platform/game not found!");
        } else {
            var mymessage = "Tags stored for " + app + ":\n";
            for (var i = 0; i < docs.length; i++) {
                mymessage += docs[i][app] + " [" + docs[i].user + "]\n";
            }
            bot.createMessage(message.channel.id, mymessage);
        }
    });
}

bot.connect();
