package oauth2flow

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
)

// OAuth2Config respresents configuration for an OAuth2 provider
type OAuth2Config struct {
	AuthorizeURL string
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scope        string
}

// AccessToken represents an authorized access token
type AccessToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// AuthCodeFlow attempts the OAuth2 Authorization Code Flow with a browser
func AuthCodeFlow(settings OAuth2Config) (AccessToken, error) {
	responseType := "code"
	redirectURI := "http://localhost:5000"
	secretState := randomString(10)

	url, err := createAuthorizationURL(settings, responseType, redirectURI, secretState)
	if err != nil {
		return AccessToken{}, err
	}

	// Open the authorize url in the system web browser
	log.Printf("If a web browser window did not open, please visit: %v", url)
	err = openURLBrowser(url)
	if err != nil {
		return AccessToken{}, err
	}

	// Make a channel for the AccessToken to return (with buffer of 1 so we don't block)
	c := make(chan AccessToken, 1)

	// Create a listener which we can close
	l, err := net.Listen("tcp", ":5000")
	if err != nil {
		return AccessToken{}, err
	}

	// Start the callback http server
	err = http.Serve(l, callbackHandler(l, settings, redirectURI, secretState, c))
	if err != nil {
		return AccessToken{}, err
	}

	// Return the token from the channel
	return <-c, nil
}

func callbackHandler(l net.Listener, settings OAuth2Config, redirectURI string, secretState string, c chan AccessToken) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		// Check state is valid
		if state != secretState {
			log.Fatal("callbackHandler: state invalid")
			return
		}

		// Create a form
		form := url.Values{
			"grant_type":   {"authorization_code"},
			"code":         {code},
			"redirect_uri": {redirectURI},
		}

		// Create http client skipping SSL verify
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}

		// Create a POST request to the token url
		req, _ := http.NewRequest("POST", settings.TokenURL, bytes.NewBufferString(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		// Add the client id and secret in the Authorization header
		clientIDSecret := settings.ClientID + ":" + settings.ClientSecret
		encodedClientIDSecret := base64.StdEncoding.EncodeToString([]byte(clientIDSecret))
		req.Header.Add("Authorization", "Basic "+encodedClientIDSecret)

		// Make the POST request
		resp, _ := client.Do(req)

		// Decode the AccessToken response
		var token AccessToken
		defer resp.Body.Close()
		_ = json.NewDecoder(resp.Body).Decode(&token)

		// Close the browser window using JavaScript
		fmt.Fprint(w, `<script type="text/javascript">window.close()</script>`)

		// Send back AccessToken through the channel
		c <- token

		// Stop the HTTP server
		l.Close()
	})
}

func createAuthorizationURL(settings OAuth2Config, responseType string, redirectURL string, state string) (string, error) {
	u, err := url.Parse(settings.AuthorizeURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("client_id", settings.ClientID)
	q.Set("response_type", responseType)
	q.Set("redirect_uri", redirectURL)
	q.Set("state", state)
	q.Set("scope", settings.Scope)

	u.RawQuery = q.Encode()
	return u.String(), nil
}
