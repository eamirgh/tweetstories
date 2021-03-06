package config

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// IFace defines all functilnality provided by the config package.
type IFace interface {
	Logger() *log.Logger
	Twitter() *twitter.Client
	Port() string
	Name() string
}

// Config holds and exposes all configurable and global objects and
// variables.
type Config struct {
	log *log.Logger

	client  *http.Client
	port    string
	name    string
	twitter *twitter.Client
}

// New instantiates an instance of Config.
func New() *Config {

	client := Twitter{}.Parse().Client()

	addr, err := determineListenAddress()
	if err != nil {
		log.Fatal(err)
	}

	name := os.Getenv("HEROKU_NAME")
	if name == "" {
		log.Fatalln("!! $HEROKU_NAME not set !!")
	}

	return &Config{
		log:     log.New(os.Stdout, "", log.LstdFlags),
		client:  client,
		port:    addr,
		name:    name,
		twitter: twitter.NewClient(client),
	}
}

// Logger exposes the app-wide logger.
func (c *Config) Logger() *log.Logger {
	return c.log
}

// Twitter exposes a Twitter client.
func (c *Config) Twitter() *twitter.Client {
	return c.twitter
}

// Port exposes the HTTP port of the server.
func (c *Config) Port() string {
	return c.port
}

// Name exposes the name of the app
func (c *Config) Name() string {
	return c.name
}

func determineListenAddress() (string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return "", fmt.Errorf("!! $PORT not set !!")
	}
	return ":" + port, nil
}

// Twitter holds configuration variables pertaining
// to the Twitter API.
type Twitter struct {
	conKey    string
	conSecret string
	token     string
	secret    string
}

// Parse loads the twitter config variables from the local
// environment.
func (t Twitter) Parse() Twitter {
	t.conKey = os.Getenv("TWITTER_CONSUMER_KEY")
	if t.conKey == "" {
		log.Fatalln("!! $TWITTER_CONSUMER_KEY not set !!")
	}

	t.conSecret = os.Getenv("TWITTER_CONSUMER_SECRET")
	if t.conSecret == "" {
		log.Fatalln("!! $TWITTER_CONSUMER_SECRET not set!!")
	}

	t.token = os.Getenv("TWITTER_ACCESS_TOKEN")
	if t.token == "" {
		log.Fatalln("!! $TWITTER_ACCESS_TOKEN not set !!")
	}

	t.secret = os.Getenv("TWITTER_ACCESS_SECRET")
	if t.secret == "" {
		log.Fatalln("!! $TWITTER_ACCESS_SECRET not set !!")
	}

	return t
}

// Client instantiates an HTTP client for
// interacting with the twitter API.
func (t Twitter) Client() *http.Client {
	c := oauth1.NewConfig(t.conKey, t.conSecret)
	tt := oauth1.NewToken(t.token, t.secret)

	return c.Client(oauth1.NoContext, tt)
}
