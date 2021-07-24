package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
)

const (
	username = "username"
	password = "password"
	src      = "alpine:3.12"
	dst      = "your-registry/alpine:3.12"
)

type operation string

var (
	imagePull operation = "Pull"
	imagePush operation = "Push"
)

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	rc, err := cli.ImagePull(context.Background(), src,
		types.ImagePullOptions{})
	if rc != nil {
		defer rc.Close()
	}
	if err != nil {
		panic(err)
	}
	if err = detectErrorMessage(rc, imagePull); err != nil {
		panic(err)
	}

	err = cli.ImageTag(context.Background(), src, dst)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n[Tag Status]: '%s' successfully tagged as '%s'\n\n", src, dst)

	authConfig := types.AuthConfig{
		Username: username,
		Password: password,
	}

	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic(err)
	}

	authStr := base64.StdEncoding.EncodeToString(encodedJSON)
	rc, err = cli.ImagePush(context.Background(), dst,
		types.ImagePushOptions{RegistryAuth: authStr})
	if rc != nil {
		defer rc.Close()
	}
	if err != nil {
		panic(err)
	}
	if err = detectErrorMessage(rc, imagePush); err != nil {
		panic(err)
	}
}

func detectErrorMessage(in io.Reader, op operation) error {
	dec := json.NewDecoder(in)
	status := ""

	for {
		var jm jsonmessage.JSONMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if jm.Error != nil {
			return jm.Error
		}
		if len(jm.ErrorMessage) > 0 {
			return errors.New(jm.ErrorMessage)
		}

		if jm.Status != "" && !strings.EqualFold(status, jm.Status) {
			fmt.Printf("[%s Status]: %v\n", op, jm.Status)
			status = jm.Status
		}
	}
	return nil
}
