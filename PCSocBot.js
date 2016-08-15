/*
 * File: PCSocBot.js
 * --------------------
 * Name: David Sison
 * discord.js Bot for UNSW PCSoc Discord.
 */

const Discord = require('discord.js');

var Datastore = require('nedb')
  , db = new Datastore({ filename: 'users.db', autoload: true });
  
var client = new Discord.Client();
//login token goes here:
client.loginWithToken('Token', output);

function output(error, token) {
    if (error) {
        console.log('There was an error logging in: ' + error);
        return;
    } else
        console.log('Logged in. Token: ' + token);
}

client.on("message", function(message) {
    var username = message.author.username;
    var id = message.author.id;
    var argv = message.content.match(/(".*?")|(\S+)/g);
    if (argv !== null) {
        var argc = argv.length;
        if(argv[0] === "!tags") {
            if (argc === 1) {
                client.sendMessage(message, "Player tag storage for the UNSW PCSoc discord server.\n\n**`!tags`** `add` __`platform/game`__ __`tag`__\n    Adds/changes a player tag with associated platform/game to the list\n**`!tags`** `remove` __`platform/game`__\n    Removes a player tag from the list\n**`!tags`** `get` __`platform/game`__\n    Returns player tag for that discord user\n**`!tags`** __`platform/game`__\n    Displays all player tags stored for platform/game\n");
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
                client.sendMessage(message, "Unknown argument(s)...");
            }
        }
    }
});

function additem(id, app, tag, message, username) {
    db.findOne({ _id: id }, function (err, doc) {
        if (doc === null) {
            var temp = { _id: id, "user": username };
            temp[app] = tag;
            db.insert(temp);
        } else {
            db.update({ _id: id }, { $set: { [app]: tag, user: username } } );
        }
        client.sendMessage(message, tag + " added as " + app + " tag for " + username);
        console.log(tag + " added as " + app + " tag for " + username);
    });
}

function getitem(id, app, message, username) {
    db.findOne({ _id: id }, function (err, doc) {
        if (doc === null || !doc.hasOwnProperty("user")) {
            client.sendMessage(message, "User not found!");
        } else if (!doc.hasOwnProperty(app)) {
            client.sendMessage(message, "Platform/game not found!");
        } else {
            client.sendMessage(message, "The " + app + " tag of " + doc.user + " is " + doc[app]);
        }
    });
}

function removeitem(id, app, message, username) {
        db.findOne({ _id: id }, function (err, doc) {
        if (doc === null) {
            client.sendMessage(message, "User not found!");
        } else if (!doc.hasOwnProperty(app)) {
            client.sendMessage(message, "Platform/game not found!");
        } else {
            db.update({ _id: id }, { $unset: { [app]: true } });
            client.sendMessage(message, app + " tag for " + username + " removed");
            console.log(app + " tag for " + username + " removed");
        }
    });
}

function printtags(app, message) {
    db.find( { [app] : { $exists: true } }, function (err, docs) {
        if (docs.length === 0) {
            client.sendMessage(message, "Platform/game not found!");
        } else {
            var mymessage = "Tags stored for " + app + ":\n";
            for (var i = 0; i < docs.length; i++) {
                mymessage += docs[i][app] + " [" + docs[i].user + "]\n";
            }
            client.sendMessage(message, mymessage);
        }
    });
}
