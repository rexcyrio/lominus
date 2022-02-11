// Package telegram provides functions that facilitates the integration with Telegram.
package telegram

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/beebeeoii/lominus/internal/file"
	"github.com/beebeeoii/lominus/pkg/api"
)

// TelegramInfo struct is the datapack that holds the data required for Telegram integration.
type TelegramInfo struct {
	BotApi string
	UserId string
}

// TelegramError struct contains the TelegramError which will be thrown when an error is returned by Telegram servers.
type TelegramError struct {
	Description string
}

const SEND_MSG_URL = "https://api.telegram.org/bot%s/sendMessage"
const CONTENT_TYPE = "application/x-www-form-urlencoded"
const POST = "POST"

// SendMessage is a wrapper function that sends a message to the user using the Bot (specified by the botApi) created by the user.
func SendMessage(botApi string, userId string, message string) error {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	reqBody := url.Values{}
	reqBody.Set("chat_id", userId)
	reqBody.Set("text", message)
	reqBody.Set("parse_mode", "HTML")
	sendMsgReq, sendMsgErr := http.NewRequest(POST, fmt.Sprintf(SEND_MSG_URL, botApi), strings.NewReader(reqBody.Encode()))

	if sendMsgErr != nil {
		return sendMsgErr
	}

	sendMsgReq.Header.Add("Content-Type", CONTENT_TYPE)
	sendMsgRes, sendMsgResErr := client.Do(sendMsgReq)

	if sendMsgResErr != nil {
		return sendMsgResErr
	}

	if sendMsgRes.StatusCode != 200 {
		bodyBytes, err := io.ReadAll(sendMsgRes.Body)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)
		return &TelegramError{Description: bodyString}
	}

	return nil
}

// GenerateGradeMessageFormat creates a message text for grade notifications.
func GenerateGradeMessageFormat(grade api.Grade) string {
	var gradeMessage string

	if grade.Comments != "" {
		gradeMessage = fmt.Sprintf("<b><u>Grades</u></b>\n<b>%s</b>: <i>%s</i>\n<b>Grade</b>: <i><tg-spoiler>%f</tg-spoiler>/%f</i>\n\n<b>Comments</b>: %s", grade.Module.ModuleCode, grade.Name, grade.Marks, grade.MaxMarks, grade.Comments)
	} else {
		gradeMessage = fmt.Sprintf("<b><u>Grades</u></b>\n<b>%s</b>: <i>%s</i>\n<b>Grade</b>: <i><tg-spoiler>%f</tg-spoiler>/%f</i>", grade.Module.ModuleCode, grade.Name, grade.Marks, grade.MaxMarks)
	}

	return gradeMessage
}

// GenerateFileUpdatedMessageFormat creates a message text for file update notifications.
func GenerateFileUpdatedMessageFormat(files []api.File) string {
	nFilesUpdated := len(files)
	updatedFilesModulesNames := []string{}

	for _, file := range files {
		updatedFilesModulesNames = append(updatedFilesModulesNames, fmt.Sprintf("[%s] %s ", file.Ancestors[0], file.Name))
	}

	var updatedFileNamesString string

	if nFilesUpdated > 4 {
		updatedFileNamesString = strings.Join(append(updatedFilesModulesNames[:3], "..."), "\n")
	} else {
		updatedFileNamesString = strings.Join(updatedFilesModulesNames, "\n")
	}

	return fmt.Sprintf("🆕 Files\n%s", updatedFileNamesString)
}

// SaveTelegramData saves the user's Telegram data onto local storage.
func SaveTelegramData(telegramDataPath string, telegramInfo TelegramInfo) error {
	return file.EncodeStructToFile(telegramDataPath, telegramInfo)
}

// LoadTelegramData loads the user's Telegram data from local storage.
func LoadTelegramData(telegramDataPath string) (TelegramInfo, error) {
	telegramInfo := TelegramInfo{}
	if !file.Exists(telegramDataPath) {
		return telegramInfo, &file.FileNotFoundError{FileName: telegramDataPath}
	}
	err := file.DecodeStructFromFile(telegramDataPath, &telegramInfo)

	return telegramInfo, err
}

// TelegramError error will be thrown when an error is returned by Telegram servers.
func (e *TelegramError) Error() string {
	return fmt.Sprintf("TelegramError: %s", e.Description)
}
