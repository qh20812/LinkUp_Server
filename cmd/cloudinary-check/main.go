package main

import (
	"fmt"
	"log"
	"os"

	"linkup/config"
)

func main() {
	env, err := config.LoadCloudinaryEnv()
	if err != nil {
		log.Printf("Cloudinary connection: failed (%v)", err)
		os.Exit(1)
	}

	plan, err := config.CheckCloudinaryConnection(env)
	if err != nil {
		log.Printf("Cloudinary connection: failed (%v)", err)
		os.Exit(1)
	}

	if plan != "" {
		fmt.Printf("Cloudinary connection: success (plan: %s)\n", plan)
		return
	}

	fmt.Println("Cloudinary connection: success")
}
