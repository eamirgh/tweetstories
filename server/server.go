package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/brendanjryan/tweetstories/config"
	"github.com/dghubble/go-twitter/twitter"
)

// Server is the main app server.
type Server struct {
	config.IFace
	http   *http.Server
	tweets map[int64]twitter.Tweet // map of id => tweet
}

// New instantiates an instance of Server
func New() *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", ack)

	cfg := config.New()

	return &Server{
		IFace:  cfg,
		tweets: map[int64]twitter.Tweet{},
		http: &http.Server{
			Handler: mux,
			Addr:    cfg.Port(),
		},
	}

}

// Run runs the server process. It can be killed by sending a SIGTERM
// or any other interrupt signal.
func (s *Server) Run() error {
	s.Logger().Println("running server")
	min := time.Tick(1 * time.Minute)
	hour := time.Tick(1 * time.Hour)

	// kill the server on any interrupt signal
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)

	// start the background http server and fetch all tweets
	go s.http.ListenAndServe()
	s.fetch()

	s.delete()

	for true {
		select {
		case <-min:
			go s.ping()
			s.delete()
		case <-hour:
			s.fetch()
		case <-kill:
			s.Logger().Println("Server killed")
			return s.http.Shutdown(context.Background())
		}
	}

	return nil
}

func (s *Server) fetch() error {
	ts, _, err := s.Twitter().Timelines.UserTimeline(&twitter.UserTimelineParams{})
	if err != nil {
		s.Logger().Println(err)
		return err
	}

	for _, t := range ts {
		s.tweets[t.ID] = t
	}

	s.Logger().Printf("Fetched Tweets: %#v", s.tweets)

	return nil
}

// ping makes a get request against the local http server.
// this is to prevent heroku from putting the bot to sleep.
func (s *Server) ping() {
	_, err := http.Get(fmt.Sprintf("http://%s.herokuapp.com/", s.Name()))
	if err != nil {
		s.Logger().Println(err)
	}
}

// delete deletes all tweets which are over a day old
func (s *Server) delete() error {
	for id, t := range s.tweets {
		if time.Since(getTime(t)).Seconds() > float64(86400*7) { // 1 week
			s.Logger().Printf("Deleting %#v", t)
			_, _, err := s.Twitter().Statuses.Destroy(t.ID, &twitter.StatusDestroyParams{})
			if err != nil {
				// print error and continue
				fmt.Println(err)
				continue
			}

			delete(s.tweets, id)
		}
	}

	return nil
}

func ack(w http.ResponseWriter, r *http.Request) {
	// do nothing
}

func getTime(t twitter.Tweet) time.Time {
	// taken from https://dev.twitter.com/overview/api/tweets
	ti, _ := time.Parse(time.RubyDate, t.CreatedAt)
	return ti
}
