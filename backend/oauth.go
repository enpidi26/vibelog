package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// ---- GitHub OAuth ----

// oauthStateCookie name for the CSRF state cookie.
const oauthStateCookie = "vibelog_gh_state"

type githubUser struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
	ID    int64  `json:"id"`
}

func (s *server) handleGithubLogin(w http.ResponseWriter, r *http.Request) {
	if s.cfg.GitHubClientID == "" || s.cfg.GitHubClientSecret == "" {
		writeError(w, http.StatusServiceUnavailable, "oauth_disabled", "GitHub OAuth is not configured")
		return
	}

	state, err := randomToken()
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
		SameSite: http.SameSiteLaxMode,
	})

	redirectURI := s.cfg.FrontendURL + "/api/auth/github/callback"
	q := url.Values{}
	q.Set("client_id", s.cfg.GitHubClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", "user:email")
	q.Set("state", state)

	http.Redirect(w, r, "https://github.com/login/oauth/authorize?"+q.Encode(), http.StatusFound)
}

func (s *server) handleGithubCallback(w http.ResponseWriter, r *http.Request) {
	if s.cfg.GitHubClientID == "" || s.cfg.GitHubClientSecret == "" {
		writeError(w, http.StatusServiceUnavailable, "oauth_disabled", "GitHub OAuth is not configured")
		return
	}

	// Verify state to protect against CSRF.
	cookie, err := r.Cookie(oauthStateCookie)
	if err != nil || cookie.Value == "" || cookie.Value != r.URL.Query().Get("state") {
		redirectWithAuthError(w, r, s.cfg.FrontendURL, "OAuth state validation failed. Please try again.")
		return
	}
	// State is single-use.
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		redirectWithAuthError(w, r, s.cfg.FrontendURL, "GitHub did not return an authorization code.")
		return
	}

	accessToken, err := exchangeGithubCode(s.cfg.GitHubClientID, s.cfg.GitHubClientSecret, code, s.cfg.FrontendURL+"/api/auth/github/callback")
	if err != nil {
		fmt.Printf("github oauth: exchange code failed: %v\n", err)
		redirectWithAuthError(w, r, s.cfg.FrontendURL, "Failed to sign in with GitHub. Please try again.")
		return
	}

	ghUser, err := fetchGithubUser(accessToken)
	if err != nil {
		fmt.Printf("github oauth: fetch user failed: %v\n", err)
		redirectWithAuthError(w, r, s.cfg.FrontendURL, "Failed to load your GitHub profile. Please try again.")
		return
	}

	email, err := fetchGithubPrimaryEmail(accessToken)
	if err != nil || email == "" {
		fmt.Printf("github oauth: no verified email for %s: %v\n", ghUser.Login, err)
		redirectWithAuthError(w, r, s.cfg.FrontendURL, "Your GitHub account has no verified email. Please add one on GitHub, then try again.")
		return
	}

	name := ghUser.Login
	if ghUser.Name != "" {
		name = ghUser.Name
	}

	// Find or create the local user.
		user, err := GetUserByEmail(s.db, email)
	if errors.Is(err, ErrNotFound) {
		// Random secret as password: OAuth users never sign in with a password.
		secret, _ := randomToken()
		hash, herr := HashPassword(secret)
		if herr != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to create account")
			return
		}
		user, err = CreateUser(s.db, name, email, hash)
		if errors.Is(err, ErrEmailTaken) {
			user, err = GetUserByEmail(s.db, email)
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

	token, err := IssueToken(s.cfg.EffectiveJWTSecret(), user.ID, user.Email, user.Name, s.cfg.JWTExpiryHrs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create token")
		return
	}

	// Hand the JWT to the frontend callback page.
	q := url.Values{}
	q.Set("token", token)
	http.Redirect(w, r, s.cfg.FrontendURL+"/auth/github/callback?"+q.Encode(), http.StatusFound)
}

func redirectWithAuthError(w http.ResponseWriter, r *http.Request, base, msg string) {
	q := url.Values{}
	q.Set("error", msg)
	http.Redirect(w, r, base+"/auth/github/callback?"+q.Encode(), http.StatusFound)
}

func exchangeGithubCode(clientID, clientSecret, code, redirectURI string) (string, error) {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("client_secret", clientSecret)
	q.Set("code", code)
	q.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", nil)
	if err != nil {
		return "", err
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github status %d: %s", res.StatusCode, string(body))
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("github returned no access_token: %s", string(body))
	}
	return out.AccessToken, nil
}

func githubAPIGet(token, path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com"+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api status %d: %s", res.StatusCode, string(body))
	}
	return body, nil
}

func fetchGithubUser(token string) (githubUser, error) {
	body, err := githubAPIGet(token, "/user")
	if err != nil {
		return githubUser{}, err
	}
	var u githubUser
	if err := json.Unmarshal(body, &u); err != nil {
		return githubUser{}, err
	}
	return u, nil
}

// fetchGithubPrimaryEmail returns the user's primary verified email.
func fetchGithubPrimaryEmail(token string) (string, error) {
	emailsBody, err := githubAPIGet(token, "/user/emails")
	if err != nil {
		return "", err
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(emailsBody, &emails); err != nil {
		return "", err
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	return "", nil
}

func randomToken() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
