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
	config := loadConfig()

	counter := 0
	for counter < 10 {
		output := rconCmd(config, "players")

		log.Info(output, slog.Int("counter", counter))

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
		panic("Failed to shut off server")
	}
}

func powerOff(cfg Config) error {
	response, err := cfg.Api.ListServers(&instance.ListServersRequest{
		Zone: scw.Zone(cfg.Zone),
	})
	if err != nil {
		log.Error("Failed to list servers",
			"error", err,
			"operation", "poweroff")
		return err
	}

	if response.TotalCount != 1 || response.Servers[0].Name != cfg.InstanceName {
		log.Error("could not find the instance, abort operation",
			"operation", "poweroff")
		return err
	}

	server := response.Servers[0]

	if !slices.Contains(server.AllowedActions, "poweroff") {
		log.Error("server action not available: poweroff",
			"operation", "poweroff")
		return err
	}

	log.Info("Stopping server", "Name", server.Name)

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
		log.Error("Failed to execute SSH command",
			"error", err,
			"output", output,
			"operation", "rconCmd")
		return ""
	}
	return strings.TrimSpace(string(output))
}
