package main

import (
	"fmt"
	"os"
	"pz-discord-bot/bot"

	"github.com/joho/godotenv"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	envOrgID        = "SCW_DEFAULT_ORGANIZATION_ID"
	envRegion       = "SCW_DEFAULT_REGION"
	envZone         = "SCW_DEFAULT_ZONE"
	envAccessKey    = "SCW_ACCESS_KEY"
	envSecretKey    = "SCW_SECRET_KEY"
	envInstanceName = "INSTANCE_NAME"
	envDiscordToken = "DISCORD_TOKEN"
)

func main() {
	cfg := loadConfig()
	bot.Cfg = cfg
	bot.Run()
}

func loadConfig() bot.Config {
	mandatoryVariables := [...]string{envOrgID, envAccessKey, envSecretKey, envRegion, envZone, envInstanceName, envDiscordToken}

	err := godotenv.Load()
	if err != nil {
		fmt.Println("no .env found")
	}
	for idx := range mandatoryVariables {
		if os.Getenv(mandatoryVariables[idx]) == "" {
			panic("missing environment variable " + mandatoryVariables[idx])
		}
	}
	client, err := scw.NewClient(
		scw.WithDefaultOrganizationID(os.Getenv(envOrgID)),
		scw.WithAuth(os.Getenv(envAccessKey), os.Getenv(envSecretKey)),
		scw.WithDefaultRegion(scw.Region(os.Getenv(envRegion))),
	)
	if err != nil {
		panic(err)
	}

	return bot.Config{
		InstanceName: os.Getenv(envInstanceName),
		DiscordToken: os.Getenv(envDiscordToken),
		Zone:         os.Getenv(envZone),
		Api:          *instance.NewAPI(client),
	}
}
