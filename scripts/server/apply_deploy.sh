#!/bin/bash
set -e

sudo cp scripts/server/deploy.service /etc/systemd/system/manifold-trading-bot.service
sudo systemctl daemon-reload
sudo systemctl restart manifold-trading-bot
sudo systemctl enable manifold-trading-bot
