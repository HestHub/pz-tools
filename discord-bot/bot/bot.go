package bot

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type Config struct {
	InstanceName string
	DiscordToken string
	Zone         string
	Api          instance.API
}

var helpMsg = `
Available commands:
!help   ->  Print this text
!start  ->  Boot up zomboid server
!stop   ->  Shut down zomboid server
!status ->  Check server status`

var Cfg Config

func Run() {
	discord, err := discordgo.New("Bot " + Cfg.DiscordToken)
	if err != nil {
		log.Fatal(err)
	}
	discord.AddHandler(newMessage)

	discord.Open()

	defer discord.Close()
	fmt.Println("Bot running....")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func newMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == discord.State.User.ID {
		return
	}

	switch {
	case strings.Contains(message.Content, "!start"):
		log.Println("start msg recieved")
		discord.ChannelMessageSend(message.ChannelID, "🧟 Starting server 🧟")
		err := sendAction(instance.ServerAction("poweron"))
		if err != nil {
			discord.ChannelMessageSend(message.ChannelID, "⚠️ Failed to start server ⚠️")
			discord.ChannelMessageSend(message.ChannelID, "Error: "+err.Error())
		} else {
			discord.ChannelMessageSend(message.ChannelID, "🧟 Server started. Happy surviving! 🧟")
		}

	case strings.Contains(message.Content, "!stop"):
		log.Println("stop msg recieved")
		discord.ChannelMessageSend(message.ChannelID, "🧟 Stopping server 🧟")
		err := sendAction(instance.ServerAction("poweroff"))
		if err != nil {
			discord.ChannelMessageSend(message.ChannelID, "⚠️ Failed to stop server ⚠️")
			discord.ChannelMessageSend(message.ChannelID, "Error: "+err.Error())
		} else {
			discord.ChannelMessageSend(message.ChannelID, "👋 Server shut down. Goodbye 👋")
		}

	case strings.Contains(message.Content, "!status"):
		log.Println("status msg recieved")
		state, err := checkStatus()
		if err != nil {
			discord.ChannelMessageSend(message.ChannelID, "⚠️ Failed to fetch server status ⚠️")
			discord.ChannelMessageSend(message.ChannelID, "Error: "+err.Error())
		} else {
			discord.ChannelMessageSend(message.ChannelID, "❓ State: "+state+" ❓")
		}

	case strings.Contains(message.Content, "!help"):
		log.Println("help msg recieved")
		discord.ChannelMessageSend(message.ChannelID, helpMsg)
	}
}

func checkStatus() (string, error) {
	api := Cfg.Api
	response, err := api.ListServers(&instance.ListServersRequest{
		Zone: scw.Zone(Cfg.Zone),
	})
	if err != nil {
		return "", err
	}

	if response.TotalCount != 1 || response.Servers[0].Name != Cfg.InstanceName {
		return "", errors.New("could not find the server instance, abort operation")
	}

	return string(response.Servers[0].State), nil
}

func sendAction(action instance.ServerAction) error {
	api := Cfg.Api
	response, err := api.ListServers(&instance.ListServersRequest{
		Zone: scw.Zone(Cfg.Zone),
	})
	if err != nil {
		return err
	}

	if response.TotalCount != 1 || response.Servers[0].Name != Cfg.InstanceName {
		return errors.New("could not find the server instance, abort operation")
	}

	server := response.Servers[0]
	if !slices.Contains(server.AllowedActions, action) {
		return errors.New("server action not available: " + string(action))
	}

	err = api.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
		ServerID: server.ID,
		Zone:     server.Zone,
		Action:   action,
	})
	if err != nil {
		return err
	}

	return nil
}
