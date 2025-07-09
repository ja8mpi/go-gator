package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	config "github.com/ja8mpi/go-gator-config"
	"github.com/ja8mpi/go-gator/internal/database"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	cmds map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	com, exists := c.cmds[cmd.name]
	if !exists {
		return fmt.Errorf("command does not exist")
	}
	return com(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	if c.cmds == nil {
		c.cmds = make(map[string]func(*state, command) error)
	}
	c.cmds[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) <= 0 {
		return fmt.Errorf("no arguments provided")
	}
	s.cfg.SetUser(cmd.arguments[0])
	fmt.Println("The user has been set")
	return nil
}

func main() {

	var cfg config.Config
	cfg.Read()
	configState := state{
		cfg: &cfg,
	}

	// connect to db
	db, err := sql.Open("postgres", cfg.DBUrl)
	dbQueries := database.New(db)

	coms := commands{
		cmds: make(map[string]func(*state, command) error),
	}
	coms.register("login", handlerLogin)

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage: <command> [arguments...]")
		os.Exit(1)
		return
	}

	commandName := args[0]
	commandsArgs := args[1:]

	cmd := command{
		name:      commandName,
		arguments: commandsArgs,
	}

	err := coms.run(&configState, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
