package main

import (
	"errors"
	"bytes"
	"fmt"
	"flag"
	"strings"
	"strconv"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/daviddengcn/go-colortext"
)

type CommandFunction func(*discordgo.Session, *discordgo.MessageCreate, string, []string) error;

type Command struct {
	prefix string
	aliases []string
	description string
	help string
	role string
	calledFunc CommandFunction
};

var (
	commands []Command;

	uname string;
	pword string;
	cmdPrefix string;
	debug bool;

	botname = "MemesToGO";

	invalid_args = errors.New("invalid args");

	kernel32, _ = syscall.LoadLibrary("kernel32.dll");
	getModuleHandle, _ = syscall.GetProcAddress(kernel32, "GetModuleHandleW");

	user32, _ = syscall.LoadLibrary("user32.dll");
	messageBox, _ = syscall.GetProcAddress(user32, "MessageBoxW");
)

func init() {
	flag.StringVar(&uname, "u", "", "Account Username");
	flag.StringVar(&pword, "p", "", "Account Password");
	flag.StringVar(&cmdPrefix, "f", ">", "Command Prefix");
	flag.BoolVar(&debug, "d", false, "Debug Mode");
	
	flag.StringVar(&uname, "username", "", "Account Username");
	flag.StringVar(&pword, "password", "", "Account Username");
	flag.StringVar(&cmdPrefix, "prefix", "", "Account Username");
	flag.BoolVar(&debug, "debug", false, "Debug Mode");

	flag.Parse();

	registerCommands();
}

func main() {

	defer syscall.FreeLibrary(kernel32);
	defer syscall.FreeLibrary(user32);

	ct.Foreground(ct.White, true);

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

	if (strings.ToLower(message.Author.Username) == strings.ToLower(botname)) {
		return;
	}

	printDebug("New message: " + message.Content);

	if (strings.HasPrefix(message.Content, cmdPrefix)) {

		cmd, args := spliceCommand(message.Content);

		for _,element := range commands {
			if (element.prefix == cmd || stringInSlice(cmd, element.aliases)) {

				if (!userCanUseCommand(session, message, element.role)) {
					session.ChannelMessageSend(message.ChannelID, "You do not have permission to use this command");
					return;
				}

				session.ChannelMessageDelete(message.ChannelID, message.ID);

				err := element.calledFunc(session, message, cmd, args);

				if (err != nil) {
					if (err == invalid_args) {
						printHelp(session, message, element, "Invalid arguments passed.");
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

	for _,char := range cmd {

		if (char == '"') {
			if (inDelim) {
				currentArg++;
				inDelim = false;
			} else {
				inDelim = true;
				continue;
			}
		} else if (char == ' ' && !inDelim) {
			currentArg++;
		}

		if (currentArg + 1 > len(args)) {
			args = append(args, "");
		} else {
			args[currentArg] = strcat(args[currentArg], string(char));
		}
	}

	return args[0], args[1:];
}

func getRolesFromMessage(session *discordgo.Session, message *discordgo.MessageCreate) ([]*discordgo.Role, *discordgo.Guild) {
	ch, err_ch := session.Channel(message.ChannelID);
	member, err_m := session.GuildMember(ch.GuildID, message.Author.ID);

	if (err_ch != nil || err_m != nil) {
		return nil, nil;
	}

	var roles []*discordgo.Role;

	for _,role := range member.Roles {
		r, err_r := session.State.Role(ch.GuildID, role);
		if (err_r != nil) {
			return nil, nil;
		}

		roles = append(roles, r);
	}

	guild, err_guild := session.State.Guild(ch.GuildID);

	if (err_guild != nil) {
		return nil, nil;
	}

	return roles, guild;
}

func userCanUseCommand(session *discordgo.Session, message *discordgo.MessageCreate, role string) bool {

	if (role == "") {
		return true;
	}

	userRoles, userGuild := getRolesFromMessage(session, message);

	if (userRoles != nil) {

		for _,userRole := range userRoles {
			for _,guildRole := range userGuild.Roles {
				if (guildRole.Name == role && userRole.Position >= guildRole.Position) {
					return true;
				}
			}
		}

	}

	return false;
}

func printHelp(session *discordgo.Session, message *discordgo.MessageCreate, cmd Command, reason string) {

	aliases := "";
	requires := "";

	if (cmd.aliases != nil && len(cmd.aliases) > 0) {
		aliases = "aliases: ";
	}

	for index,alias := range cmd.aliases {

		aliases += alias;

		if (index != len(cmd.aliases) - 1) {
			aliases += ", ";
		} else {
			aliases += "\n";
		}
	}

	if (cmd.role != "") {
		requires = "\nrequires: " + cmd.role + " or higher\n"
	}

	session.ChannelMessageSend(message.ChannelID, strcat(reason, "\n```xl\n", cmdPrefix, cmd.prefix, " ", cmd.help, "\n", aliases, "-    ", cmd.description, requires, "```"));
}

// Command generation
/*
prefix = the main command
aliases = a string array of alternate commands that do the same thing
description = what the command does
help = all of the arguments. [] = optional, <> = required
role = the minimum required role to use this command. Blank = no requirement
func = the function called
*/
func createCommand(prefixStr string, aliasList []string, description string, help string, role string, callFunc CommandFunction) error {

	printDebug("Registering command: " + prefixStr);

	pref := strings.ToLower(prefixStr);
	var aliases []string;

	for _,element := range aliasList {
		aliases = append(aliases, strings.ToLower(element));
	}

	for _,cmd := range commands {

		if (pref == cmd.prefix || stringInSlice(pref, cmd.aliases)) {

			printDebug("There is already a command with the prefix or alias '" + pref + "'");

			return errors.New("There is already a command with the prefix or alias '" + pref + "'");
		}

		for _,alias := range aliases {
			if (alias == cmd.prefix || stringInSlice(alias, cmd.aliases)) {

				printDebug("There is already a command with the prefix or alias '" + alias + "'");

				return errors.New(strcat("There is already a command with the prefix or alias '", alias, "'"));
			}
		}
	}

	commands = append(commands, Command{pref, aliases, description, help, role, callFunc});

	return nil;
}

func printDebug(text string) {
	if (debug) {
		ct.ChangeColor(ct.Yellow, true, ct.Black, false);
		fmt.Println("[debug] " + text + "\n");	
		ct.ResetColor();
	}
}

// Register commands here
func registerCommands() {

	printDebug("Registering commands");

	registerHelp();
	registerStrokeOrder();

	if (debug) {
		registerTest();	
	}
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