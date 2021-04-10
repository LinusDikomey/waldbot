package oauth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const REDIRECT = "https://wald.mbehrmann.de/auth"
const SCOPE = "identify"

var (
	ClientId string
	ClientSecret string
)

type AccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    uint64 `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"identify"`
	TokenType    string `json:"token_type"`
}

func request(method string, endpoint string, data io.Reader, auth *string, oauth bool) ([]byte, int, error) {
	req, err := http.NewRequest(method, "https://discord.com/api/v8" + endpoint, data)

	if err != nil {
		fmt.Println("could not create request:", err)
		return nil, -1, err
	}

	if oauth {
		req.SetBasicAuth(ClientId, ClientSecret)
	}

	if method != "GET" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	if auth != nil {
		req.Header.Add("Authorization", "Bearer " + *auth)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error while executing request:", err)
		return nil, -1, err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	resp.Body.Close()
	return buf.Bytes(), resp.StatusCode, nil
}

func oauth2Token(data url.Values) (*AccessTokenResponse, error) {
	strings.NewReader(data.Encode())

	bytes, status, err := request("POST", "/oauth2/token", strings.NewReader(data.Encode()), nil, true)
	if err != nil {
		return nil, err
	}

	if status != 200 {
		return nil, errors.New("Invalid status code")
	}
	body := AccessTokenResponse{}
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		fmt.Println("error while parsing auth response:", err)
		return nil, err
	}
	return &body, nil
}

func AuthCode(code string) (*AccessTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", REDIRECT)
	data.Set("scope", SCOPE)
	return oauth2Token(data)
}

func RefreshToken(refresh string) (*AccessTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refresh)
	data.Set("redirect_uri", REDIRECT)
	data.Set("scope", SCOPE)
	return oauth2Token(data)
}

type AuthorizationResponse struct {
	User discordgo.User		`json:"user"`
}

func Me(access string) (*AuthorizationResponse, error) {
	resp, status, err := request("GET", "/oauth2/@me", nil, &access, false)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		fmt.Println("Received invalid status code while requesting: '/oauth2/@me': ", string(resp))
		return nil, errors.New("Invalid status code")
	}
	body := AuthorizationResponse {}
	err = json.Unmarshal(resp, &body)
	if err != nil {
		return nil, err
	}
	return &body, nil
}