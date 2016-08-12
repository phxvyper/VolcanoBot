package main

import (
	"errors"
	"bytes"
	"fmt"
	"flag"
	"strings"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

type CommandFunction func(*discordgo.Session, *discordgo.MessageCreate, string, []string) error;

type Command struct {
	prefix string
	aliases []string
	help string
	role string
	calledFunc CommandFunction
};

var uname string;
var pword string;
var cmdPrefix string;

var commands []Command;

func init() {
	flag.StringVar(&uname, "u", "", "Account Username");
	flag.StringVar(&pword, "p", "", "Account Password");
	flag.StringVar(&cmdPrefix, "f", ">", "Command Prefix");
	flag.Parse();

	registerCommands();
}

func main() {
	if (uname == "" || pword == "") {
		fmt.Println("No user info provided. Please run: volcano -u <email> -p <password>");
		return;
	}

	// Discord sessions init
	dg, err := discordgo.New(uname, pword);
	if (err != nil) {
		fmt.Println("Error creating Discord session: ", err);
		return;
	}

	dg.AddHandler(messageCreate);

	err = dg.Open();
	if (err != nil) {
		fmt.Println("Error opening Discord session: ", err);
	}

	fmt.Println("Volcano is now running. Press Ctrl-C to exit.");
	// Ctrl+C to quit
	<-make(chan struct{});
	return;
}

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if (strings.HasPrefix(message.Content, cmdPrefix)) {

		cmd, args := spliceCommand(message.Content);

		for _,element := range commands {
			if (element.prefix == cmd || stringInSlice(cmd, element.aliases)) {
				if (!stringInSlice(element.role, getRolesFromMessage(message))) {
					session.ChannelMessageSend(message.ChannelID, "You do not have permission to use this command");
					return;
				}
				err := element.calledFunc(session, message, cmd, args);

				if (err != nil) {
					if (err.Error() == "invalid args") {
						session.ChannelMessageSend(message.ChannelID, strcat("Invalid arguments passed\n", element.help));
					} else {
						session.ChannelMessageSend(message.ChannelID, err.Error());
					}
				}
			}
		}
	}
}

// Helpful utility functions
func stringInSlice(n string, list []string) bool {
	for _,b := range list {
		if (b == n) {
			return true;
		}
	}

	return false;
}

func strcat(a string, b ...string) string {
	var buffer bytes.Buffer

	buffer.WriteString(a);

	for _,element := range b {
		buffer.WriteString(element);
	}

	return buffer.String();
}

func spliceCommand(cmd string) (string, []string) {
	var args []string;

	inDelim := false;
	currentArg := 0;

	for i := 0; i < len(cmd); i++ {
		ch := cmd[i];

		if (ch == '"') {
			if (inDelim) {
				currentArg++;
				inDelim = false;
			} else {
				inDelim = true;
				continue;
			}
		} else if (ch == ' ' && !inDelim) {
			currentArg++;
		}

		if (currentArg + 1 > len(args)) {
			args = append(args, "");
		} else {
			args[currentArg] = strcat(args[currentArg], string(ch));
		}
	}

	return args[0], args[1:];
}

func getRolesFromMessage(session *discordgo.Session, message *discordgo.MessageCreate) []string {
	ch, err_ch := session.Channel(message.ChannelID);
	member, err_m := session.GuildMember(ch, message.Author.ID);

	return member.Roles;
}

// Command generation
func createCommand(prefixStr string, aliasList []string, helpStr string, role string, callFunc CommandFunction) error {

	pref := strings.ToLower(prefixStr);
	var aliases []string;
	help := strings.ToLower(helpStr);

	for _,element := range aliasList {
		aliases = append(aliases, strings.ToLower(element));
	}

	for _,cmd := range commands {

		if (pref == cmd.prefix || stringInSlice(pref, cmd.aliases)) {
			return errors.New(strcat("There is already a command with the prefix or alias '", pref, "'"));
		}

		for _,alias := range aliases {
			if (alias == cmd.prefix || stringInSlice(alias, cmd.aliases)) {
				return errors.New(strcat("There is already a command with the prefix or alias '", alias, "'"));
			}
		}
	}

	commands = append(commands, Command{pref, aliases, help, role, callFunc});

	return nil;
}

// Register commands here
func registerCommands() {
	createCommand("test", []string{"test"}, strcat(cmdPrefix, "test\n- prints \"testing!\""), "Cancer", testCommandFunction);

	createCommand("permissions",
		[]string{"perms"},
		strcat(cmdPrefix, "permissions\n- prints \"the permissions of the user who sent the command.\""),
		"Owner",
		showUserPermissions);
}

// Command functions go down here (events that occur when a function is called)
func testCommandFunction(session *discordgo.Session, message *discordgo.MessageCreate, cmd string, args []string) error {
	session.ChannelMessageSend(message.ChannelID, "testing!");

	return nil;
}

func showUserPermissions(session *discordgo.Session, message *discordgo.MessageCreate, cmd string, args []string) error {

	i, err := session.State.UserChannelPermissions(message.Author.ID, message.ChannelID);

	if (err != nil) {
		return err;
	}

	session.ChannelMessageSend(message.ChannelID, strcat("Your permissions: \n",
		strconv.FormatInt(int64(i), 2), "\n",
		strconv.FormatInt(int64(i), 10), "\n",
		strconv.FormatInt(int64(i), 16)));

	return nil;
}