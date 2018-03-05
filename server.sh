while `true`
do
    echo "starting server"
    timeout 24h python3 bot.py
    echo "killing server to restart"
    git pull
    sleep 1
done
