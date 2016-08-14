package main

import (
	"encoding/json"
	"net/http"

	"strconv"
	"fmt"
	"io/ioutil"

	"github.com/bwmarrin/discordgo"
)

var (
	jishoApi = "http://jisho.org/api/v1/search/words?keyword="
)

type JishoWord struct {
	Word string `json:"word"`;
	Reading string `json:"reading"`;
}

type JishoSense struct {
	EnglishDefinitions []string `json:"english_definitions"`;
	PartsOfSpeech []string `json:"parts_of_speech"`;
}

type JishoMeta struct {
	Status int64 `json:"status"`;
}

type JishoData struct {
	IsCommon bool `json:"is_common"`;
	Tags []string `json:"tags"`;
	Japanese []JishoWord `json:"japanese"`;
	Senses []JishoSense `json:"senses"`;
}

type JishoBody struct {
	Meta JishoMeta `json:"meta"`;
	Data []JishoData `json:"data"`;
}

func getJishoBody(body []byte) (*JishoBody, error) {
	var s = new(JishoBody);
	err := json.Unmarshal(body, &s);
	return s, err;
}

func registerJisho() {
	err := createCommand(
		"jisho",
		[]string{"define"},
		"replies with an image of the stroke order for the provided kanji.",
		"<keyword> [result] [definition]",
		"",
		jishoCommand);

	if (err != nil) {
		fmt.Println(err);
	}
}

func jishoCommand(session *discordgo.Session, message *discordgo.MessageCreate, cmd string, args []string) error {

	if (len(args) < 1) {
		return invalid_args;
	}

	var item int64 = 0;
	var defined int64 = 0;
	var err error;

	if (len(args) > 1) {
		item, err = strconv.ParseInt(args[1], 10, 64);
		if (err != nil) {
			return invalid_args;
		}

		item--;
	}

	if (len(args) > 2) {
		defined, err = strconv.ParseInt(args[2], 10, 64);
		if (err != nil) {
			return invalid_args;
		}

		defined--;
	}

	waitingMessage, _ := session.ChannelMessageSend(message.ChannelID, "Waiting for jisho.org, ちょっと待ってください。。。");

	response, err := http.Get(jishoApi + args[0]);

	if (err != nil) {
		printDebug(err.Error());
		return err;
	}

	responseBody, err := ioutil.ReadAll(response.Body);
	if (err != nil) {
		printDebug(err.Error());
		return err;
	}

	body, err := getJishoBody([]byte(responseBody));
	if (err != nil) {
		printDebug(err.Error());
		return err;
	}

	if (int(item) > len(body.Data)) {
		item = 0;
	}

	word := body.Data[int(item)];

	if (int(defined) > len(word.Senses)) {
		defined = 0;
	}

	definition := word.Senses[int(defined)];
	readings := word.Japanese;

	/*
	printed like:

	「何」 on Jisho.org:
		Readings: 「なに」, 「なん」.
		In English: "what".
		Parts of Speech: pronoun, no-adjective.

		There are 5 other definitions of 「何」.
	*/

	readingsText := "";
	inEnglishText := "";
	partsText := "";

	for i := 0; i < len(readings); i++ {
		jword := readings[i];
		if (i == len(readings) - 1) {
			readingsText += "「" + jword.Reading + "」.";
		} else {
			readingsText += "「" + jword.Reading + "」, ";
		}
	}

	for i := 0; i < len(definition.EnglishDefinitions); i++ {
		edef := definition.EnglishDefinitions[i];
		if (i == len(readings) - 1) {
			inEnglishText += "\"" + edef + "\".";
		} else {
			inEnglishText += "\"" + edef + "\", ";
		}
	}

	for i := 0; i < len(definition.PartsOfSpeech); i++ {
		part := definition.PartsOfSpeech[i];
		if (i == len(readings) - 1) {
			partsText += "\"" + part + "\".";
		} else {
			partsText += "\"" + part + "\", ";
		}
	}

	chosenReading := readings[0].Word;

	if (chosenReading == "") {
		chosenReading = args[0];
	}

	finalText := "\n```xl\n「" + chosenReading + "」 on jisho.org:\n" +
		" -  readings: " + readingsText + "\n" +
		" -  english: " + inEnglishText + "\n" +
		" -  part-of-speech: " + partsText + "\n" +
		"\n" +
		" -  there are " + strconv.Itoa(len(body.Data) - 1) + " other results for 「" + args[0] + "」\n```";

	session.ChannelMessageEdit(message.ChannelID, waitingMessage.ID, finalText);

	return nil;
}