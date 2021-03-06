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
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
)

func (s *Server) setRoutes(mux *chi.Mux) {
	// unauthenticated endpoints
	mux.Get("/", s.RootHandler)
	mux.Get("/privacy", s.PrivacyHandler)
	mux.Get("/login", s.LoginHander)
	mux.Get("/auth", s.AuthHandler)

	mux.Get("/v3/token", s.TokenHandler)

	// authenticated endpoints
	mux.Route("/v3", func(r chi.Router) {
		r.With(s.authMiddleware).Get("/exercises", s.GetExercises)

		r.With(s.authMiddleware).Get("/stats", s.GetStats)
		r.With(s.authMiddleware).Post("/stats", s.PostStats)
		r.With(s.authMiddleware).Post("/stats/all", s.GetStatsAll)
		r.With(s.authMiddleware).Get("/stats/team/{teamID}", s.GetStatsForTeam)

		r.With(s.authMiddleware).Get("/teams", s.GetTeams)
		r.With(s.authMiddleware).Post("/teams", s.PostTeams)
		r.With(s.authMiddleware).Delete("/team/{teamID}", s.DeleteTeam)

		r.With(s.authMiddleware).Get("/myteams", s.GetMyTeams)
		r.With(s.authMiddleware).Post("/myteams/{teamID}", s.PostMyTeams)
		r.With(s.authMiddleware).Delete("/myteams/{teamID}", s.DeleteMyTeams)

	})

	filesDir := http.Dir(s.conf.FilesPath)
	FileServer(mux, "/files", filesDir)

}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
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

const ctxEmail = "ctxEmail"
const ctxUID = "ctxUID"

// TODO: rate limit? caddy server limits total requests at a time, but not by IP I think
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: allow auth bypass for local dev?
		Bearer := r.Header.Get("Authorization")
		if !strings.HasPrefix(strings.ToLower(Bearer), "bearer ") { // note the space is required
			http.Error(w, "Authorization: Bearer $token required in header", http.StatusBadRequest)
			return
		}
		// "Bearer $token"
		parts := strings.Split(Bearer, " ")
		t, ok := s.tokenCache.Get(parts[1])
		if !ok {
			http.Error(w, "invalid token", http.StatusBadRequest)
			return
		}
		token, ok := t.(Token)
		if !ok {
			log.Println("unexpected token failure - bad type")
			http.Error(w, "unexpected invalid token", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxEmail, token.email)
		ctx = context.WithValue(ctx, ctxUID, token.uid)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (s *Server) TokenHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing param: code", http.StatusBadRequest)
		return
	}

	oAuth, err := s.oAuthValidate(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uid, err := s.getOrCreateUser(oAuth.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := s.createAndStoreToken(uid, oAuth.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(token)
}

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

// oAuthValidate returns the email address of the signed in user via Google OAuthv2, or an error
func (s *Server) oAuthValidate(code string) (*googleAuthResp, error) {
	if s.DevMode {
		log.Printf("dev mode enabled, setting email as code value of %s", code)
		return &googleAuthResp{Email: code}, nil
	}

	tok, err := s.oAuthConf.Exchange(context.Background(), code)
	if err != nil {
		log.Println("error with exchange", err)
		return nil, fmt.Errorf("unalbe to oAuthValidate with exchange: %w", err)
	}

	client := s.oAuthConf.Client(context.Background(), tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		log.Println("error getting google api user info", err)
		return nil, fmt.Errorf("unalbe to oAuthValidate client get: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("error reading google api user info", err)
		return nil, fmt.Errorf("unalbe to oAuthValidate read body: %w", err)
	}

	var authResp googleAuthResp
	err = json.Unmarshal(body, &authResp)
	if err != nil {
		log.Println("error marshalling google api user info", err)
		return nil, fmt.Errorf("unable to oAuthValidate marshal: %w", err)
	}

	if !strings.HasSuffix(authResp.Email, "twilio.com") {
		return nil, fmt.Errorf("invalid email address: twilio.com required")
	}

	return &authResp, nil
}

func (s *Server) AuthHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: have this page serve SPA stuff.
	// For now, redirect to the token handler to make it easy to get the token for manual curl testing
	s.TokenHandler(w, r)
}

func (s *Server) GetStats(w http.ResponseWriter, r *http.Request) {
	uid, _ := r.Context().Value(ctxUID).(int)
	start, end := getStartAndEndTS(r)
	stats, err := s.getStats([]int{uid}, start, end)
	if err != nil {
		log.Printf("unable to getStats: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) GetStatsAll(w http.ResponseWriter, r *http.Request) {
	start, end := getStartAndEndTS(r)
	stats, err := s.getStats([]int{}, start, end)
	if err != nil {
		log.Printf("unable to getStats: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) GetStatsForTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := strconv.Atoi(chi.URLParam(r, "teamID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	start, end := getStartAndEndTS(r)
	stats, err := s.getStatsForTeam(teamID, start, end)
	if err != nil {
		log.Printf("unable to getStatsForTeam: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func getStartAndEndTS(r *http.Request) (int, int) {
	startDate := r.URL.Query().Get("startdate")
	endDate := r.URL.Query().Get("enddate")

	start, _ := strconv.Atoi(startDate)
	end, _ := strconv.Atoi(endDate)

	if start == 0 {
		start = int(time.Now().Add(-31 * 24 * time.Hour).Unix())
	}
	if end == 0 {
		end = int(time.Now().Add(24 * time.Hour).Unix())
	}

	return start, end
}

func (s *Server) PostStats(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var exs Exercises
	err = json.Unmarshal(reqBody, &exs)
	if err != nil {
		log.Printf("bad unmarshalling: %s, %s", err.Error(), string(reqBody))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uid, _ := r.Context().Value(ctxUID).(int)
	err = s.postStats(uid, exs)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) GetExercises(w http.ResponseWriter, r *http.Request) {
	data, err := s.getExercises()
	if err != nil {
		log.Println("error GetExercises ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("GetExercises marshal err ", err.Error())
	}
}

func (s *Server) GetTeams(w http.ResponseWriter, r *http.Request) {
	data, err := s.getAllTeams(-1)
	if err != nil {
		log.Println("error GetTeams ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("GetTeams marshal err ", err.Error())
	}
}
func (s *Server) PostTeams(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUID).(int)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("error reading body PostTeams: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	team := &Team{}

	err = json.Unmarshal(body, team)
	if err != nil {
		log.Println("error marshalling body PostTeams: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if team.Name == "" {
		http.Error(w, "team name required", http.StatusBadRequest)
		return
	}

	newTeam, err := s.postTeam(team.Name, uid)
	if err != nil {
		log.Printf("error postTeam %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(newTeam); err != nil {
		log.Println("PostTeams marshal err ", err.Error())
	}
}

func (s *Server) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUID).(int)
	teamID, err := strconv.Atoi(chi.URLParam(r, "teamID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if teamID <= 0 {
		http.Error(w, "teamid must be > 0", http.StatusBadRequest)
		return
	}

	err = s.deleteTeam(teamID, uid)
	if err != nil {
		log.Println("unable to DeleteTeam: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GetMyTeams(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUID).(int)
	data, err := s.getMyTeams(uid)
	if err != nil {
		log.Println("unable to GetMyTeams: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("GetMyTeams marshal err ", err.Error())
	}
}

func (s *Server) GetTeamsCreatedByUser(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUID).(int)
	data, err := s.getAllTeams(uid)
	if err != nil {
		log.Println("error GetMyTeams ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("GeTeamsCreatedByUser marshal err ", err.Error())
	}
}

func (s *Server) PostMyTeams(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUID).(int)
	teamID, err := strconv.Atoi(chi.URLParam(r, "teamID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.postMyTeams(teamID, uid)
	if err != nil {
		log.Println("unable to PostMyTeams: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) DeleteMyTeams(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUID).(int)
	teamID, err := strconv.Atoi(chi.URLParam(r, "teamID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.deleteMyTeams(teamID, uid)
	if err != nil {
		log.Println("unable to PostMyTeams: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) PrivacyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("stub page. TL;DR: your info is not shared with anyone for any reason. This service is just used internally by Twilio as a side project. We will store your email address in our db as to identify you and relate your exercise totals to you."))
}
