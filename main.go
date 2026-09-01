package main

import (
	"log"
	"visuilizer/anilist"
)

func main() {
	bakemonogatariID := 5081

	client := anilist.NewClient()

	entry, rel, err := client.FetchMedia(bakemonogatariID)

	if err != nil {
		log.Fatalf("couldn't fetch Bakemonogatari data: %v", err)
	}

	log.Println(entry, "\n", "\n")
	for _, r := range rel {
		log.Println(r)
	}
}
