package main

import (
	"log/slog"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func main() {
	config := LoadConfig()

	discord := config.Discord

	discord.Open()

	defer discord.Close()

	counter := 0
	for counter < 10 {
		output := rconCmd(config, "players")
		Log.Info(output, slog.Int("counter", counter))

		if strings.Contains(output, config.Rcon.state) {
			counter++
		} else {
			counter = 0
		}
		time.Sleep(60 * time.Second)
	}

	rconCmd(config, "save")

	err := powerOff(config)
	if err != nil {
		discord.ChannelMessageSend(config.ChannelID, "Failed to stop server: "+err.Error())
		panic("Failed to shut off server")
	}
}

func powerOff(cfg Config) error {
	discord := cfg.Discord
	response, err := cfg.Api.ListServers(&instance.ListServersRequest{
		Zone: scw.Zone(cfg.Zone),
	})
	if err != nil {
		discord.ChannelMessageSend(cfg.ChannelID, "Failed to list server: "+err.Error())
		Log.Error("Failed to list servers",
			"error", err,
			"operation", "poweroff")
		return err
	}

	if response.TotalCount != 1 || response.Servers[0].Name != cfg.InstanceName {
		discord.ChannelMessageSend(cfg.ChannelID, "Instance not found:")
		Log.Error("could not find the instance, abort operation",
			"operation", "poweroff")
		return err
	}

	server := response.Servers[0]

	if !slices.Contains(server.AllowedActions, "poweroff") {
		discord.ChannelMessageSend(cfg.ChannelID, "Failed perform action: "+err.Error())
		Log.Error("server action not available: poweroff",
			"operation", "poweroff")
		return err
	}

	Log.Info("Stopping server", "Name", server.Name)

	return cfg.Api.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
		ServerID: server.ID,
		Zone:     server.Zone,
		Action:   "poweroff",
	})
}

func rconCmd(cfg Config, cmd string) string {
	cmdCheck := exec.Command(cfg.Rcon.path, "-a", cfg.Rcon.addr, "-p", cfg.Rcon.pwd, cmd)
	output, err := cmdCheck.CombinedOutput()
	if err != nil {
		Log.Error("Failed to execute SSH command",
			"error", err,
			"output", output,
			"operation", "rconCmd")
		return ""
	}
	return strings.TrimSpace(string(output))
}
