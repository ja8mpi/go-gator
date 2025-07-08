package main

import (
	"fmt"

	config "github.com/ja8mpi/go-gator-config"
)

func main() {
	var cfg config.Config
	cfg.Read()
	cfg.SetUser("daniel")
	cfg.Read()
	fmt.Printf("%v\n%v\n", cfg.DBUrl, cfg.CurrentUserName)
}
