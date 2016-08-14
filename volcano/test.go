package main

import (
	"github.com/bwmarrin/discordgo"
)


func registerTest() {
	createCommand(
		"test",
		nil,
		"prints \"testing!\"",
		"",
		"",
		testCommandFunction);
}

func testCommandFunction(session *discordgo.Session, message *discordgo.MessageCreate, cmd string, args []string) error {
	session.ChannelMessageSend(message.ChannelID, "testing!");

	return nil;
}