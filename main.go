package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println(`
  ___       _____        _       ___
 / __|___  |_   _|__  __| |___  |   \ ___ _ __  ___
| (_ / _ \   | |/ _ \/ _| / _ \ | |) / -_) '  \/ _ \
 \___\___/   |_|\___/\__,_\___/ |___/\___|_|_|_\___/
 `)

	err := godotenv.Load()
	check(err)

	var store storage

	useMongo := parseFlags()
	if useMongo {
		store = &mongoStore{}
		fmt.Println("(using mongodb datastore)")
	} else {
		store = &inMemoryStore{}
		fmt.Println("(using in-memory datastore)")
	}

	err = store.initialize()
	check(err)

	server := NewServer(os.Getenv("GO_TODO_DEMO_ADDR"), store)
	server.start()

	err = store.shutdown()
	check(err)
}

func parseFlags() bool {
	useMongo := flag.Bool("m", false, "if true, use mongo db instead of in-memory store")
	flag.Parse()
	return *useMongo
}
