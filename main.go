package main

import (
	_ "github.com/zuercher/slackbot/importer"
	"github.com/zuercher/slackbot/robots"
	"github.com/zuercher/slackbot/server"
)

func main() {
	server.Main(robots.Robots)
}
