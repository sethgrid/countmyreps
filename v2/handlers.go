package countmyreps

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

func (s *Server) setRoutes(mux *chi.Mux) {
	// unauthenticated endpoints
	mux.Get("/", s.RootHandler)
	mux.Get("/privacy", s.PrivacyHandler)
	mux.Get("/login", s.LoginHander)
	mux.Get("/auth", s.AuthHandler)

	// authenticated endpoints
	mux.Get("/stats", s.StatsHandler)
	mux.Get("/exercises", s.GetExercises)
}

// on the front end, after a user signs in:
/*
function onSignIn(googleUser) {
  var id_token = googleUser.getAuthResponse().id_token;
  ...
}
*/

// and now that token can be passed to the backend. We will want to store this token as X-GOOGLE-
/*
var xhr = new XMLHttpRequest();
xhr.open('POST', 'https://yourbackend.example.com/tokensignin');
xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
xhr.onload = function() {
  console.log('Signed in as: ' + xhr.responseText);
};
xhr.send('idtoken=' + id_token);
*/

func (s *Server) RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
<html>
<head>
 <script src="https://apis.google.com/js/platform.js" async defer></script>
 <meta name="google-signin-client_id" content="156533058127-78t7m5d1vlac3k0sk5ls0a6hk4nj2ius.apps.googleusercontent.com">
</head>
<body>
  <div class="g-signin2" data-onsuccess="onSignIn"></div>
  <a href="#" onclick="signOut();">Sign out</a>
  <script>
    function signOut() {
      var auth2 = gapi.auth2.getAuthInstance();
      auth2.signOut().then(function () {
        console.log('User signed out.');
      });
    }
  </script>
</body>
</html>

`))
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func (s *Server) LoginHander(w http.ResponseWriter, r *http.Request) {
	// TODO: store csrfState for validating against callback from oAuth (cookie?)
	csrfState := randToken()
	w.Write([]byte("<html><title>Golang Google</title> <body> <a href='" + s.oAuthConf.AuthCodeURL(csrfState) + "'><button>Login with Google!</button> </a> </body></html>"))
}

type googleAuthResp struct {
	Sub           string `json:"sub"` // numerical
	PicURL        string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	HD            string `json:"hd"`
}

func (s *Server) AuthHandler(w http.ResponseWriter, r *http.Request) {
	// TODO - retreive csrfState
	// compare it to query state
	csrfStateVerify := r.URL.Query().Get("state")
	_ = csrfStateVerify

	code := r.URL.Query().Get("code")
	tok, err := s.oAuthConf.Exchange(context.Background(), code)
	if err != nil {
		log.Println("error with exchange", err)
		http.Error(w, "unable to validate exchange", http.StatusBadRequest)
		return
	}

	client := s.oAuthConf.Client(context.Background(), tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		log.Println("error getting google api user info", err)
		http.Error(w, "unable to get user info from google", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("error reading google api user info", err)
		http.Error(w, "unable to read user info from google", http.StatusBadRequest)
		return
	}
	log.Println("google api resp body: ", string(body))

	var authResp googleAuthResp
	err = json.Unmarshal(body, &authResp)
	if err != nil {
		log.Println("error marshalling google api user info", err)
		http.Error(w, "unable to marshall user info from google", http.StatusBadRequest)
		return
	}

	// store authResp onto context or state?
	// redirect to app data page
	http.Redirect(w, r, "/stats", http.StatusTemporaryRedirect)
}

func (s *Server) StatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("you must be logged in to view this page. <a href='https://www.google.com/accounts/Logout?continue=https://appengine.google.com/_ah/logout?continue=%s'>Click to log out</a>", s.conf.FullAddr)))
}

func (s *Server) GetExercises(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) PrivacyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("stub page. TL;DR: your info is not shared with anyone for any reason. This service is just used internally by Twilio as a side project. We will store your email address in our db as to identify you and relate your exercise totals to you."))
}
