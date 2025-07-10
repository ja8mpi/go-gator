package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
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

func handleAgg(s *state, cmd command) error {
	// Option A: Use background context (simple, but no timeout or cancellation)
	ctx := context.Background()
	// Option B: With timeout
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	feed, err := fetchFeed(ctx, "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}

	// Do something with the fetched RSS feed
	fmt.Printf("Fetched feed: %+v\n", feed)

	return nil
}

func handleAddFeed(s *state, cmd command) error {
	if len(cmd.arguments) < 2 {
		return fmt.Errorf("Need name and url as well")
	}

	feedName := cmd.arguments[0]
	feedUrl := cmd.arguments[1]

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	currentUserName := s.cfg.CurrentUserName
	var user_id uuid.UUID

	for _, user := range users {
		if user.Name == currentUserName {
			user_id = user.ID
			break
		}
	}

	timeNow := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	feed_id := uuid.New()
	s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        feed_id,
		Name:      feedName,
		Url:       feedUrl,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		UserID:    user_id,
	})

	s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user_id,
		FeedID: feed_id,
	})

	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	var feed RSSFeed

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		fmt.Println("An error occured creating the request")
		return &feed, err
	}
	req.Header.Set("User-Agent", "gator")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&feed)
	if err != nil {
		panic(err)
	}
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for _, item := range feed.Channel.Item {
		item.Title = html.UnescapeString(feed.Channel.Title)
		item.Description = html.UnescapeString(feed.Channel.Description)
	}

	return &feed, nil
}

func handleFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		fmt.Println("Error getting feeds")
		os.Exit(1)
		return err
	}

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Println("Error getting users")
		os.Exit(1)
		return err
	}
	// First, make a map for quick user lookup
	userMap := make(map[uuid.UUID]string)
	for _, user := range users {
		userMap[user.ID] = user.Name
	}

	for _, feed := range feeds {
		fmt.Printf("* %v\n", feed.Name)
		fmt.Printf("* %v\n", feed.Name)
		fmt.Printf("* %v\n", feed.Name)
		fmt.Printf("* %v\n", userMap[feed.UserID])
	}

	os.Exit(0)
	return nil
}

func handleFollow(s *state, cmd command) error {
	if len(cmd.arguments) <= 0 {
		return fmt.Errorf("no arguments provided")
	}
	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		fmt.Println("User does not exists")
		os.Exit(1)
		return err
	}
	url := cmd.arguments[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		fmt.Println("User does not exists")
		os.Exit(1)
		return err
	}

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})

	if err != nil {
		fmt.Println("Error creating follow")
		os.Exit(1)
		return err
	}
	fmt.Println("%v", s.cfg.CurrentUserName)
	fmt.Println("%v", feed.Name)

	return nil
}

func handleFollowing(s *state, cmd command) error {
	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		fmt.Println("Error getting feeds for current user")
		os.Exit(1)
		return err
	}

	for _, feed := range feedFollows {
		fmt.Printf("%v\n", feed.FeedName)
	}
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
	coms.register("agg", handleAgg)
	coms.register("addfeed", handleAddFeed)
	coms.register("feeds", handleFeeds)
	coms.register("follow", handleFollow)
	coms.register("following", handleFollowing)

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
