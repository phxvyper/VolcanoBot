package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)


func registerHelp() {
	createCommand(
		"help",
		[]string{"?"},
		"prints help info about [cmd], or shows all commands.",
		"[cmd]",
		"",
		helpCommand);
}

func helpCommand(session *discordgo.Session, message *discordgo.MessageCreate, cmd string, args []string) error {

	if (len(args) < 1) {
		info := "\n```xl\nCommands:\n";

		for _,item := range commands {
			info += "	" + cmdPrefix + item.prefix + " " + item.help + ": " + item.description + "\n\n";
		}

		info += "\n```";

		session.ChannelMessageSend(message.ChannelID, info);

		return nil;

	} else {
		for _,item := range commands {
			if (item.prefix == strings.ToLower(args[0]) || stringInSlice(strings.ToLower(args[0]), item.aliases)) {
				
				printHelp(session, message, item, "");

				return nil;
			}
		}
	}

	return nil;
}