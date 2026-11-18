package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/enpidi26/vibelog/internal/auth"
	"github.com/enpidi26/vibelog/internal/database"
	"github.com/enpidi26/vibelog/internal/domain"
)

const oauthStateCookie = "vibelog_gh_state"

func (s *Server) handleGithubLogin(w http.ResponseWriter, r *http.Request) {
	if s.cfg.GitHubClientID == "" || s.cfg.GitHubClientSecret == "" {
		writeError(w, http.StatusServiceUnavailable, "oauth_disabled", "GitHub OAuth is not configured")
		return
	}
	state, err := auth.RandomToken(20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create OAuth state")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   s.cfg.BaseURLIsHTTPS(),
		SameSite: http.SameSiteLaxMode,
	})

	q := url.Values{}
	q.Set("client_id", s.cfg.GitHubClientID)
	q.Set("redirect_uri", s.cfg.BaseURL+"/api/auth/github/callback")
	q.Set("scope", "user:email")
	q.Set("state", state)
	http.Redirect(w, r, "https://github.com/login/oauth/authorize?"+q.Encode(), http.StatusFound)
}

func (s *Server) handleGithubCallback(w http.ResponseWriter, r *http.Request) {
	if s.cfg.GitHubClientID == "" || s.cfg.GitHubClientSecret == "" {
		writeError(w, http.StatusServiceUnavailable, "oauth_disabled", "GitHub OAuth is not configured")
		return
	}

	// Verify state to protect against CSRF.
	cookie, err := r.Cookie(oauthStateCookie)
	if err != nil || cookie.Value == "" || cookie.Value != r.URL.Query().Get("state") {
		redirectWithAuthError(w, r, s.cfg.BaseURL, "OAuth state validation failed. Please try again.")
		return
	}
	// State is single-use.
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.BaseURLIsHTTPS(),
		SameSite: http.SameSiteLaxMode,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		redirectWithAuthError(w, r, s.cfg.BaseURL, "GitHub did not return an authorization code.")
		return
	}

	accessToken, err := exchangeGithubCode(s.cfg.GitHubClientID, s.cfg.GitHubClientSecret, code, s.cfg.BaseURL+"/api/auth/github/callback")
	if err != nil {
		fmt.Printf("github oauth: exchange code failed: %v\n", err)
		redirectWithAuthError(w, r, s.cfg.BaseURL, "Failed to sign in with GitHub. Please try again.")
		return
	}

	ghUser, err := fetchGithubUser(accessToken)
	if err != nil {
		fmt.Printf("github oauth: fetch user failed: %v\n", err)
		redirectWithAuthError(w, r, s.cfg.BaseURL, "Failed to load your GitHub profile. Please try again.")
		return
	}

	emailAddr, err := fetchGithubPrimaryEmail(accessToken)
	if err != nil || emailAddr == "" {
		fmt.Printf("github oauth: no verified email for %s: %v\n", ghUser.Login, err)
		redirectWithAuthError(w, r, s.cfg.BaseURL, "Your GitHub account has no verified email. Please add one on GitHub, then try again.")
		return
	}

	name := ghUser.Login
	if ghUser.Name != "" {
		name = ghUser.Name
	}

	user, err := database.GetUserByEmail(s.db, emailAddr)
	if errors.Is(err, domain.ErrNotFound) {
		if !s.cfg.AllowRegistration {
			redirectWithAuthError(w, r, s.cfg.BaseURL, "Registration is closed on this instance.")
			return
		}
		// Random secret as password: OAuth users never sign in with a password.
		secret, _ := auth.RandomToken(20)
		hash, herr := auth.HashPassword(secret)
		if herr != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to create account")
			return
		}
		user, err = database.CreateUser(s.db, name, emailAddr, hash)
		if errors.Is(err, domain.ErrEmailTaken) {
			user, err = database.GetUserByEmail(s.db, emailAddr)
		}
		if err != nil {
			fmt.Printf("github oauth: create user failed: %v\n", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to create account")
			return
		}
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load account")
		return
	}

	// Email GitHub sudah terverifikasi, jadi tandai terverifikasi juga di sini.
	if !user.EmailVerified {
		if err := database.MarkEmailVerified(s.db, user.ID); err == nil {
			user.EmailVerified = true
		}
	}

	token, err := s.issueToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create token")
		return
	}

	// Hand the JWT to the frontend callback page.
	q := url.Values{}
	q.Set("token", token)
	http.Redirect(w, r, s.cfg.BaseURL+"/auth/github/callback?"+q.Encode(), http.StatusFound)
}

func redirectWithAuthError(w http.ResponseWriter, r *http.Request, base, msg string) {
	q := url.Values{}
	q.Set("error", msg)
	http.Redirect(w, r, base+"/auth/github/callback?"+q.Encode(), http.StatusFound)
}
