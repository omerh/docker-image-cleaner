package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

var filterEnv string
var freshness int
var tickerWait time.Duration
var filterMap map[string]struct{}

func main() {
	log.Println("Starting image cleaner")
	log.Printf("Running with filter \"%s\", and time interval of %v and freshness of %v min", filterEnv, tickerWait, freshness)

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
	// for errors
	var err error

	// image freshness
	freshString, _ := os.LookupEnv("FRESHNESS")
	freshness, err = strconv.Atoi(freshString)
	if err != nil {
		freshness, _ = strconv.Atoi("-30")
	}
	if freshness > 0 {
		freshness = freshness * -1
	}

	// ticker
	tickerEnvWait, ok := os.LookupEnv("TIME_INTERVAL")
	if !ok {
		tickerEnvWait = "24h"
	}
	// var err error
	tickerWait, err = time.ParseDuration(tickerEnvWait)
	if err != nil {
		log.Printf("Ticking with %d", tickerWait)
		log.Println(err)
	}
}

func runtime() {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Println("Failed to initiate client")
		panic(err)
	}

	log.Println("Listing images on host")
	images := listDockerImages(cli)

	var TotalSpaceReduced int64

	freshImageTime := time.Now().Add(time.Minute * time.Duration(freshness))
	log.Printf("Will skip images that were created after %v", freshness)

	// prun containers
	log.Println("Pruning unused containers")
	pruneContainers(cli)

	// Prune volumes
	log.Println("Prunning unused volumes")
	pruneUnusedVolumes(cli)

	for _, image := range images {
		if image.Created < freshImageTime.Unix() {
			// Image are older than 30 minutes and can be deleted
			for _, tag := range image.RepoTags {
				// check if a tag is set on the image to skip delete
				if _, ok := filterMap[tag]; ok {
					log.Printf("Image %v, filttered out and will not be deleted", tag)
				} else {
					log.Printf("Cleaning image %v, with tags %v", image.ID, image.RepoTags)
					reducedSpace := deleteDockerImage(image, cli)
					TotalSpaceReduced = TotalSpaceReduced + reducedSpace
				}
			}
		} else {
			log.Printf("Skipping deletion of image %v, its to fresh", image.ID)
		}
	}

	if TotalSpaceReduced > 0 {
		printSize(TotalSpaceReduced)
	}
}

func listDockerImages(cli *client.Client) []types.ImageSummary {
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		log.Println("Failed to list docker images on host")
		log.Println(err)
	}
	return images
}

func deleteDockerImage(image types.ImageSummary, cli *client.Client) int64 {
	// Set docker remove options
	removeOptions := types.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	}
	// Delete image
	_, err := cli.ImageRemove(context.Background(), image.ID, removeOptions)
	if err != nil {
		log.Printf("Can't remove image %v, image is being used by a running container", image.ID)
	} else {
		return image.Size
	}
	return 0
}

func printSize(reducedSpace int64) {
	sizeInMb := float64(reducedSpace) / 1000 / 1000
	log.Printf("Cleaned %.2fMB from disk", sizeInMb)
}

func pruneUnusedVolumes(cli *client.Client) {
	_, volErr := cli.VolumesPrune(context.Background(), filters.Args{})
	if volErr != nil {
		log.Println("Failed to delete unused volumes")
	}
}

func pruneContainers(cli *client.Client) {
	_, pruneErr := cli.ContainersPrune(context.Background(), filters.Args{})
	if pruneErr != nil {
		log.Println("Failed to prune unused containers")
	}
}
