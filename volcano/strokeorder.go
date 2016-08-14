package main

import (
	"os"
	"os/exec"
	"io"

	"fmt"
	"strings"
	"strconv"
	"time"

	"archive/zip"
	"path/filepath"

	"github.com/bwmarrin/discordgo"
)

var (
	fileNameLength = 5;
	zipReader *zip.ReadCloser;
)

func registerStrokeOrder() {
	err := createCommand(
		"strokeorder",
		[]string{"strokes", "stroke"},
		"replies with an image of the stroke order for the provided kanji.",
		"<kanji>",
		"",
		strokeOrderCommand);

	if (err != nil) {
		fmt.Println(err);
	}

	zipReader, err = zip.OpenReader("data/kanji.zip");

	if (err != nil) {
		fmt.Println(err);
	}

	deferCallback(closeZipReader);
}

func closeZipReader() {
	zipReader.Close();
}

func isKanji(unicode string) bool {
	v,_ := strconv.ParseUint(strings.ToUpper(unicode), 16, 32);
	return (v >= 0x4E00 && v <= 0x9FC3) ||
		(v >= 0x3400 && v <= 0x4DBF) ||
		(v >= 0xF900 && v <= 0xFAD9) ||
		(v >= 0x2E80 && v <= 0x2EFF) ||
		(v >= 0x20000 && v <= 0x2A6DF);
}

func strokeOrderCommand(session *discordgo.Session, message *discordgo.MessageCreate, cmd string, args []string) error {
	
	if (len(args) < 1) {
		return invalid_args;
	}

	printDebug("Opening kanji.zip to search for the kanji '" + args[0] + "'");

	time0 := time.Now();

	for pos,char := range args[0] {
		if (pos == 0) {
			quotedcode := fmt.Sprintf("%+q", char);

			if (!strings.HasPrefix(quotedcode[1:], `\u`)) {
				break;
			}

			printDebug("Silicing code: " + quotedcode);
			kanjiFileName := quotedcode[3:len(quotedcode) - 1];

			for (len(kanjiFileName) < fileNameLength) {
				kanjiFileName = "0" + kanjiFileName;
			}

			if (!isKanji(kanjiFileName)) {
				break;
			}

			printDebug("Searching for kanji '" + args[0] + "' (" + kanjiFileName + ")");
			msg,_ := session.ChannelMessageSend(message.ChannelID, "Looking for that kanji, ちょっと待ってください。。。");

			for _, f := range zipReader.File {
				if (f.Name == kanjiFileName + ".svg") {

					printDebug("Kanji found!");

					fileReader, err := f.Open();

					if (err != nil) {
						printDebug(err.Error());
						return err;
					}
					defer fileReader.Close();

					path := filepath.Join("data/temp/", "kanji.svg");

					targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode());
					if (err != nil) {
						printDebug(err.Error());
						return err;
					}
					defer targetFile.Close();

					if _, err := io.Copy(targetFile, fileReader); err != nil {
						printDebug(err.Error());
						return err;
					}

					if _,err := os.Stat("data/temp/kanji.png"); err == nil {
						os.Remove("data/temp/kanji.png");
					}
					if _,err := os.Stat("data/temp/kanji.svg"); err == nil {
						os.Remove("data/temp/kanji.svg");
					}

					cmd := exec.Command("inkscape", "-z", "-e", "data/temp/kanji.png", "-w 870", "-h 870", "data/temp/kanji.svg");
					err = cmd.Run();
					if (err != nil) {
						printDebug(err.Error());
						return err;
					}

					png, err := os.Open("data/temp/kanji.png");
					if (err != nil) {
						printDebug(err.Error());
						return err;
					}

					session.ChannelMessageEdit(message.ChannelID, msg.ID, "どうぞ！");
					_,err = session.ChannelFileSend(message.ChannelID, "kanji.png", png);

					time1 := time.Now();
					duration := time1.Sub(time0);

					printDebug("It took " + strconv.FormatFloat(duration.Seconds(), 'f', -1, 64) + " seconds to grab find that kanji and upload it.");

					return nil;
				}
			}
		} else {
			break;
		}
	}

	printDebug("Couldn't find the kanji '" + args[0] + "', or it is invalid!");
	session.ChannelMessageSend(message.ChannelID, "'" + args[0] + "' isn't a valid Kanji!");

	return nil;
}