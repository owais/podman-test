package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/containers/podman/v3/pkg/bindings"
	"github.com/containers/podman/v3/pkg/domain/entities"
)

func main() {
	ctx, err := bindings.NewConnectionWithIdentity(context.Background(), "ssh://core@localhost:65174/run/user/1000/podman/podman.sock", "/Users/olone/.ssh/podman-machine-default")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Starting")
	ch := fromRest(ctx)
	for e := range ch {
		fmt.Println(e)
	}
}

func fromRest(ctx context.Context) chan entities.Event {
	ch := make(chan entities.Event)

	conn, err := bindings.GetClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	params := url.Values{"stream": []string{"true"}}

	response, err := conn.DoRequest(nil, http.MethodGet, "/events", params, nil)
	if err != nil {
		log.Fatalf("error making request to /events: %w", err)
	}

	go func() {
		dec := json.NewDecoder(response.Body)
		for {
			var event entities.Event
			if err := dec.Decode(&event); err != nil {
				if err == io.EOF {
					break
				}
				log.Fatalf("error decoding event: %w", err)
			}
			ch <- event
		}
		response.Body.Close()

	}()
	return ch
}
