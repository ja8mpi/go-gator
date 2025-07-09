package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	config "github.com/ja8mpi/go-gator-config"
	"github.com/ja8mpi/go-gator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

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

	name := cmd.arguments[0]
	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		fmt.Println("User does not exists")
		os.Exit(1)
		return err
	}

	s.cfg.SetUser(name)
	fmt.Println("The user has been set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) <= 0 {
		return fmt.Errorf("no arguments provided")
	}
	name := cmd.arguments[0]
	_, err := s.db.GetUser(context.Background(), name)
	if err == nil {
		fmt.Println("User already exists")
		os.Exit(1)
		return err
	}

	timeNow := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	})

	s.cfg.SetUser(name)

	fmt.Println("The user has been set")
	os.Exit(0)
	return nil
}
func handleReset(s *state, cmd command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		fmt.Println("Could not delete users")
		os.Exit(1)
		return err
	}
	fmt.Println("Deleted all users")
	os.Exit(0)
	return nil
}

func handleGetAllUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Println("Error getting users")
		os.Exit(1)
		return err
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %v (current)\n", user.Name)
		} else {
			fmt.Printf("* %v\n", user.Name)
		}
	}

	os.Exit(0)
	return nil
}

func main() {

	var cfg config.Config
	cfg.Read()

	// connect to db
	db, err := sql.Open("postgres", cfg.DBUrl)
	dbQueries := database.New(db)

	configState := state{
		cfg: &cfg,
		db:  dbQueries,
	}

	coms := commands{
		cmds: make(map[string]func(*state, command) error),
	}
	coms.register("login", handlerLogin)
	coms.register("register", handlerRegister)
	coms.register("reset", handleReset)
	coms.register("users", handleGetAllUsers)
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

	err = coms.run(&configState, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
