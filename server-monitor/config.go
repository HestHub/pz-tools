package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	envOrgID        = "SCW_DEFAULT_ORGANIZATION_ID"
	envAccessKey    = "SCW_ACCESS_KEY"
	envSecretKey    = "SCW_SECRET_KEY"
	envRegion       = "SCW_DEFAULT_REGION"
	envZone         = "SCW_DEFAULT_ZONE"
	envInstanceName = "INSTANCE_NAME"
	envRconPwd      = "RCON_PWD"
	envRconAddr     = "RCON_ADDR"
	envRconState    = "RCON_STATE"
	envRconPath     = "RCON_PATH"
)

type RCON struct {
	path  string
	state string
	addr  string
	pwd   string
}

type Config struct {
	Rcon         RCON
	InstanceName string
	Zone         string
	Api          instance.API
}

var log slog.Logger

func loadConfig() Config {
	godotenv.Load()
	setLog()

	mandatoryVariables := [...]string{
		envOrgID,
		envAccessKey,
		envSecretKey,
		envRegion,
		envZone,
		envRconPwd,
		envRconPwd,
		envRconState,
		envRconState,
		envInstanceName,
	}

	for idx := range mandatoryVariables {
		if os.Getenv(mandatoryVariables[idx]) == "" {
			log.Error("Failed to execute SSH command",
				"missing", mandatoryVariables[idx],
				"operation", "loadConfig")
			os.Exit(1)
		}
	}
	client, err := scw.NewClient(
		scw.WithDefaultOrganizationID(os.Getenv(envOrgID)),
		scw.WithAuth(os.Getenv(envAccessKey), os.Getenv(envSecretKey)),
		scw.WithDefaultRegion(scw.Region(os.Getenv(envRegion))),
	)
	if err != nil {
		log.Error("Failed to create scw client",
			"error", err,
			"operation", "loadConfig")
		os.Exit(1)
	}

	return Config{
		Rcon: RCON{
			path:  os.Getenv(envRconPath),
			state: os.Getenv(envRconState),
			addr:  os.Getenv(envRconAddr),
			pwd:   os.Getenv(envRconPwd),
		},
		InstanceName: os.Getenv(envInstanceName),
		Zone:         os.Getenv(envZone),
		Api:          *instance.NewAPI(client),
	}
}

func setLog() {
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		panic(err)
	}
	today := time.Now().Format("01-02")
	logFileName := fmt.Sprintf("logs/%s.log", today)

	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	handler := slog.NewTextHandler(logFile, nil)
	log := slog.New(handler)

	slog.SetDefault(log)
}
