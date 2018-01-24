#!/bin/bash
touch $1
> $1
robots=(
    "github.com/zuercher/slackbot/robots/decide"
    "github.com/zuercher/slackbot/robots/bijin"
    "github.com/zuercher/slackbot/robots/nihongo"
    "github.com/zuercher/slackbot/robots/ping"
    "github.com/zuercher/slackbot/robots/roll"
    "github.com/zuercher/slackbot/robots/store"
    "github.com/zuercher/slackbot/robots/wiki"
    "github.com/zuercher/slackbot/robots/bot"
)

echo "package importer

import (" >> $1

for robot in "${robots[@]}"
do
    echo "    _ \"$robot\" // automatically generated import to register bot, do not change" >> $1
done
echo ")" >> $1

gofmt -w -s $1
