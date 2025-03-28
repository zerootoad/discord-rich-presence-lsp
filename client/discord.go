package client

import (
	"net/http"
	"strings"
	"time"

	"github.com/hugolgst/rich-go/client"
	log "github.com/sirupsen/logrus"
	"github.com/zerootoad/discord-rpc-lsp/utils"
)

var (
	debouncer = utils.NewDebouncer(5 * time.Second)
)

func Login(applicationID string) error {
	return client.Login(applicationID)
}

func Logout() {
	client.Logout()
}

func replacePlaceholders(s string, placeholders map[string]string) string {
	for placeholder, value := range placeholders {
		s = strings.Replace(s, placeholder, value, -1)
	}
	return s
}

func updateActivityConfig(config *utils.Config, placeholders map[string]string) utils.ActivityConfig {
	newActivity := utils.ActivityConfig{
		State:      replacePlaceholders(config.Discord.Activity.State, placeholders),
		Details:    replacePlaceholders(config.Discord.Activity.Details, placeholders),
		LargeImage: replacePlaceholders(config.Discord.Activity.LargeImage, placeholders),
		LargeText:  replacePlaceholders(config.Discord.Activity.LargeText, placeholders),
		SmallImage: replacePlaceholders(config.Discord.Activity.SmallImage, placeholders),
		SmallText:  replacePlaceholders(config.Discord.Activity.SmallText, placeholders),
	}

	return newActivity
}

func getImageURL(url string, defaultURL string) string {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return defaultURL
	}
	defer resp.Body.Close()
	return url
}

func UpdateDiscordActivity(config *utils.Config, action, filename, workspace, currentLang, editor, gitRemoteURL, gitBranchName string, timestamp *time.Time) error {
	placeholders := map[string]string{
		"{action}":    action,
		"{filename}":  filename,
		"{workspace}": workspace,
		"{editor}":    editor,
		"{language}":  currentLang,
	}

	tempActivity := updateActivityConfig(config, placeholders)

	smallImage := getImageURL(tempActivity.SmallImage, "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/text.png")
	largeImage := getImageURL(tempActivity.LargeImage, "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/text.png")
	if editor == "neovim" {
		largeImage = "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/Nvemo.png"
	}

	if currentLang == "" {
		smallImage = ""
		tempActivity.SmallText = ""
	}

	activity := client.Activity{
		State:      tempActivity.State,
		Details:    tempActivity.Details,
		LargeImage: largeImage,
		LargeText:  tempActivity.LargeText,
		SmallImage: smallImage,
		SmallText:  tempActivity.SmallText,
	}

	switch config.Discord.LargeUse {
	case "language":
		activity.LargeImage = smallImage
		activity.LargeText = tempActivity.SmallText
	case "editor":
		activity.LargeImage = largeImage
		activity.LargeText = tempActivity.LargeText
	}

	switch config.Discord.SmallUse {
	case "language":
		activity.SmallImage = smallImage
		activity.SmallText = tempActivity.SmallText
	case "editor":
		activity.SmallImage = largeImage
		activity.SmallText = tempActivity.LargeText
	}

	if config.Discord.Activity.Timestamp {
		activity.Timestamps = &client.Timestamps{
			Start: timestamp,
		}
	}

	if gitRemoteURL != "" {
		activity.Buttons = []*client.Button{
			{
				Label: "View Repository",
				Url:   gitRemoteURL,
			},
		}
		activity.Details += " (" + gitBranchName + ")"
	}

	var err error
	debouncer.Debounce(func() {
		log.Info("Updating discord activity")
		err = client.SetActivity(activity)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to update Discord activity")
		}
	})
	return err
}

func ClearDiscordActivity(config *utils.Config, action, filename, workspace, editor, gitRemoteURL, gitBranchName string) error {
	placeholders := map[string]string{
		"{action}":    action,
		"{filename}":  filename,
		"{workspace}": workspace,
		"{editor}":    editor,
	}

	tempActivity := updateActivityConfig(config, placeholders)

	largeImage := getImageURL(tempActivity.LargeImage, "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/text.png")
	if editor == "neovim" {
		largeImage = "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/Nvemo.png"
	}

	now := time.Now()
	activity := client.Activity{
		State:      tempActivity.State,
		Details:    tempActivity.Details,
		LargeImage: largeImage,
		LargeText:  tempActivity.LargeText,
		Timestamps: &client.Timestamps{
			Start: &now,
		},
	}

	if gitRemoteURL != "" {
		activity.Buttons = []*client.Button{
			{
				Label: "View Repository",
				Url:   gitRemoteURL,
			},
		}
		activity.Details += " (" + gitBranchName + ")"
	}

	var err error
	debouncer.Debounce(func() {
		log.Info("Clear discord activity")
		err = client.SetActivity(activity)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to clear Discord activity")
		}
	})
	return err
}
