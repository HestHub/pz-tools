package pzfunc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"

	"github.com/go-playground/validator/v10"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type Data struct {
	Action string `json:"action"      validate:"required,oneof=poweroff poweron"`
}

// based on https://github.com/scaleway/serverless-examples/blob/main/functions/go-mail/handler.go

func Handler(respWriter http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		respWriter.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body Data

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		respWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := validator.New().Struct(body); err != nil {
		respWriter.WriteHeader(http.StatusBadRequest)
		_, _ = respWriter.Write([]byte(err.Error()))
		return
	}

	action := instance.ServerAction(body.Action)

	if err := sendAction(action); err != nil {
		respWriter.WriteHeader(http.StatusInternalServerError)
		_, _ = respWriter.Write([]byte(err.Error()))
		return
	}
}

func sendAction(action instance.ServerAction) error {
	client, err := scw.NewClient(
		scw.WithDefaultOrganizationID(os.Getenv("SCW_DEFAULT_ORGANIZATION_ID")),
		scw.WithAuth(os.Getenv("SCW_ACCESS_KEY"), os.Getenv("SCW_SECRET_KEY")),
		scw.WithDefaultRegion(scw.Region(os.Getenv("SCW_DEFAULT_REGION"))),
	)
	if err != nil {
		return fmt.Errorf("error creating scaleway client with sdk %w", err)
	}

	api := instance.NewAPI(client)
	response, err := api.ListServers(&instance.ListServersRequest{
		Zone: scw.Zone(os.Getenv("SCW_DEFAULT_ZONE")),
	})
	if err != nil {
		return err
	}

	if response.TotalCount != 1 || response.Servers[0].Name != os.Getenv("INSTANCE_NAME") {
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
