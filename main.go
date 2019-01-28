package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var filterEnv string
var tickerWait time.Duration
var filterMap map[string]struct{}

func main() {
	log.Println("Starting image cleaner")
	log.Printf("Running with filter \"%s\", and time interval of %v", filterEnv, tickerWait)

	ticker := time.NewTicker(tickerWait)
	for range ticker.C {
		runtime()
	}
}

func init() {
	// Docker Api version
	_, ok := os.LookupEnv("DOCKER_API_VERSION")
	if !ok {
		os.Setenv("DOCKER_API_VERSION", "1.39")
	}
	// image filter
	filterEnv, _ = os.LookupEnv("FILTER")
	filterSlice := strings.Split(filterEnv, ",")
	filterMap = make(map[string]struct{})

	for _, f := range filterSlice {
		filterMap[f] = struct{}{}
	}

	// ticker
	tickerEnvWait, ok := os.LookupEnv("TIME_INTERVAL")
	if !ok {
		tickerEnvWait = "60s"
	}
	var err error
	tickerWait, err = time.ParseDuration(tickerEnvWait)
	if err != nil {
		log.Printf("Ticking with %d", tickerWait)
		log.Println(err)
	}
}

func runtime() {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	log.Println("Listing images on host")
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	removeOptions := types.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	}

	var reducedSpace int64

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if _, ok := filterMap[tag]; ok {
				log.Printf("Image %v, filttered out and will not be deleted", tag)
			} else {
				log.Printf("Cleaning image %v, with tags %v", image.ID, image.RepoTags)
				_, err := cli.ImageRemove(context.Background(), image.ID, removeOptions)
				if err != nil {
					log.Printf("Can not removing image %v, Image is in use by a running container", image.ID)
					// log.Printf("Reason: %v", err)
				} else {
					reducedSpace = reducedSpace + image.Size
				}
			}
		}
	}

	if reducedSpace > 0 {
		printSize(reducedSpace)
	}
}

func printSize(reducedSpace int64) {
	sizeInMb := float64(reducedSpace) / 1000 / 1000
	log.Printf("Cleaned %.2fMB from disk", sizeInMb)
}
