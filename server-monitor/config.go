package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/bwmarrin/discordgo"

	"github.com/joho/godotenv"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	envOrgID            = "SCW_DEFAULT_ORGANIZATION_ID"
	envAccessKey        = "SCW_ACCESS_KEY"
	envSecretKey        = "SCW_SECRET_KEY"
	envRegion           = "SCW_DEFAULT_REGION"
	envZone             = "SCW_DEFAULT_ZONE"
	envInstanceName     = "INSTANCE_NAME"
	envRconPwd          = "RCON_PWD"
	envRconAddr         = "RCON_ADDR"
	envRconState        = "RCON_STATE"
	envRconPath         = "RCON_PATH"
	envDiscordToken     = "DISCORD_TOKEN"
	envDiscordChannelId = "DISCORD_PRIVATE_CHANNELD"
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
	Discord      *discordgo.Session
	ChannelID    string
}

var Log *slog.Logger

func LoadConfig() Config {
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
		envDiscordToken,
		envDiscordChannelId,
	}

	for idx := range mandatoryVariables {
		if os.Getenv(mandatoryVariables[idx]) == "" {
			Log.Error("Failed to execute SSH command",
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
		Log.Error("Failed to create scw client",
			"error", err,
			"operation", "loadConfig")
		os.Exit(1)
	}

	discord, err := discordgo.New("Bot " + os.Getenv(envDiscordToken))
	if err != nil {
		log.Fatal(err)
	}

	return Config{
		Rcon: RCON{
			path:  os.Getenv(envRconPath),
			state: os.Getenv(envRconState),
			addr:  os.Getenv(envRconAddr),
			pwd:   os.Getenv(envRconPwd),
		},
		Discord:      discord,
		ChannelID:    os.Getenv(envDiscordChannelId),
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

	logFile, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	wrt := io.MultiWriter(os.Stdout, logFile)

	handler := slog.NewTextHandler(wrt, nil)
	Log = slog.New(handler)

	slog.SetDefault(Log)
}
