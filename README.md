go-oauth2flow
=============

This simple package allows you to authorize with an OAuth2 Authorization Code Flow
endpoint without copying and pasting codes.

The package opens the OAuth2 authorize url with the system browser with the `redirect_uri` set as
`http://localhost:5000` and listens for the callback. An AccessToken is then returned.

## Example

```go
package main

import "github.com/aaron7/go-oauth2flow"

func main() {
	config := oauth2flow.OAuth2Config{
		AuthorizeURL: "https://accounts.spotify.com/authorize",
		TokenURL:     "https://accounts.spotify.com/api/token",
		ClientID:     "a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0",
		ClientSecret: "b1b1b1b1b1b1b1b1b1b1b1b1b1b1b1b1",
		Scope:        "",
	}
	token := oauth2flow.AuthCodeFlow()
	log.Printf("AccessToken: %+v", token)
}
```

## Package oauth2flow

### func AuthCodeFlow
```go
func AuthCodeFlow(config OAuth2Config) AccessToken
```


### type OAuth2Config

```go
type OAuth2Config struct {
	AuthorizeURL string
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scope        string
}
```

### type AccessToken

```go
type AccessToken struct {
	AccessToken  string
	TokenType    string
	Scope        string
	ExpiresIn    int
	RefreshToken string
}
```

