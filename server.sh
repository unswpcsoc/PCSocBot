echo "Updating bot..."
git pull
echo
echo "Starting bot..."
python3 bot.py "$1"
