package countmyreps

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	_ "github.com/mattn/go-sqlite3"
	"github.com/patrickmn/go-cache"
	"github.com/sethgrid/countmyreps/v2/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Credentials struct {
	CID     string `json:"cid"`
	CSecret string `json:"csecret"`
}

type Server struct {
	DB *sql.DB

	conf        *config.Config
	httpSrv     *http.Server
	googleCreds Credentials
	oAuthConf   *oauth2.Config
	tokenCache  *cache.Cache
	rand        *rand.Rand
}

func NewServer(c *config.Config) (*Server, error) {
	s := &Server{conf: c}
	s.tokenCache = cache.New(60*time.Minute, 15*time.Minute)
	s.rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	if err := c.Sanitize(); err != nil {
		return nil, err
	}

	var creds Credentials
	file, err := ioutil.ReadFile(c.GoogleCredsPath)
	if err != nil {
		log.Fatalf("cred file error: %v for path %q", err, c.GoogleCredsPath)
	}
	err = json.Unmarshal(file, &creds)
	if err != nil {
		log.Fatalf("unmarshal creds error: %v", err)
	}

	s.oAuthConf = &oauth2.Config{
		ClientID:     creds.CID,
		ClientSecret: creds.CSecret,
		RedirectURL:  fmt.Sprintf("%s/auth", c.FullAddr),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email", // You have to select your own scope from here -> https://developers.google.com/identity/protocols/googlescopes#google_sign-in
		},
		Endpoint: google.Endpoint,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		s.Close()
	}()

	if err := s.InitDB(); err != nil {
		return nil, err
	}

	mux := chi.NewMux()
	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Port),
		Handler: mux,
	}

	s.setRoutes(mux)

	return s, nil
}

func (s *Server) Serve() error {
	log.Printf("serving on :%d", s.conf.Port)
	if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// don't overwrite the existing error needed to show why the server failed, but still capture any problems with the server shutting down
		if err2 := s.Close(); err2 != nil {
			log.Println(err2.Error())
		}
		log.Fatal(err)
	}
	return nil
}

func (s *Server) Close() error {
	defer s.DB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := s.httpSrv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	if s.conf.RemoveDBOnShutdown {
		if err := os.Remove(s.conf.DBPath); err != nil {
			return fmt.Errorf("unable to remove db - %w", err)
		}
	}
	return nil
}

type Token struct {
	Token string
	email string
	uid   int
}

// createAndStoreToken returns a random string of valid characters for a bearer token. It can return an error to be future proof if we change the token storage mechanism
func (s *Server) createAndStoreToken(uid int, email string) (Token, error) {
	t := Token{email: email, uid: uid, Token: s.RandStringRunes(48)}
	// stores with default timeout (60 min)
	s.tokenCache.Add(t.Token, t, -1)

	return t, nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+/-_0123456789")

func (s *Server) RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[s.rand.Intn(len(letterRunes))]
	}
	return string(b)
}
