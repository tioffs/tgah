package tgah

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	// regxConfirm regx parse hash.
	regxConfirm = `hash=([a-z0-9]+)`
	// regxUserJSON regx parse json object.
	regxUserJSON = `{[^;\s]*}`
	// oauthURL domain telegram oauth.
	oauthURL = "https://oauth.telegram.org/"
	// authURL endpoint.
	authURL = oauthURL + "auth?bot_id=%d&origin=%s&embed=0&request_access=write"
	// loginURL endpoint.
	loginURL = oauthURL + "auth/login?bot_id=%d&origin=%s&embed=1&request_access=write"
	// confirmHash endpoint.
	confirmHash = oauthURL + "auth/auth?bot_id=%d&origin=%s&request_access=write&confirm=1&%s"
	// confirmURL endpoint.
	confirmURL = oauthURL + "auth/push?bot_id=%d&origin=%s&request_access=write"
	// sendPhoneURL endpoint.
	sendPhoneURL = oauthURL + "auth/request?bot_id=%d&origin=%s&embed=1&request_access=write"
	// phone send params body.
	phone = "phone=%s"
	// cleanupInterval clean cache time second.
	cleanupInterval = 10
	// expires default time second cache delete.
	expires = 60
	// cookiesDELETE const status cookie header.
	cookiesDELETE = "DELETED"
	// declined User cancel auth.
	declined = "Declined by the User"
	// trueStr check string true.
	trueStr = "true"
)

// User struct telegram oauth info User.
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int    `json:"auth_date"`
	Hash      string `json:"hash"`
	Phone     string `json:"phone"`
}

type item struct {
	Cookies map[string]string
	Created int64
}

type confirmUser struct {
	PhoneNumber string
}

// store memory auth telegram User.
var store = make(map[string]item)

// domainUrl site connect bot telegram auth.
var domainURL string

// bot id bot telegram.
var bot int32

// statusCache is auto delete cache not auth User.
var statusCache = false

// cache goroutine checking and delete item cache memory.
func cleanup() {
	if statusCache {
		return
	}
	statusCache = true
	for {
		<-time.After(cleanupInterval * time.Second)
		if len(store) > 0 {
			for i := range store {
				if store[i].Created < time.Now().Unix() {
					delete(store, i)
				}
			}
		}
	}
}

func getCookies(key string) (cookie map[string]string, is bool) {
	i, is := store[key]
	if !is {
		cookie = make(map[string]string)

		return
	}
	cookie = i.Cookies

	return
}

func setCookies(key string, cookie map[string]string) {
	store[key] = item{
		Created: time.Now().Unix() + expires,
		Cookies: cookie,
	}
}

// Setting set variable bot id and domain.
func Setting(botID int32, domain string) {
	go cleanup()
	domainURL = fmt.Sprintf("https://%s", domain)
	bot = botID
}

// SendPhoneTelegram send push notify telegram User phone.
func SendPhoneTelegram(userPhone string) (err error) {
	confirm := &confirmUser{PhoneNumber: userPhone}
	phone := fmt.Sprintf(phone, userPhone)
	response, err := confirm.httpClient(fmt.Sprintf(sendPhoneURL, bot, domainURL), http.MethodPost, &phone)
	if err != nil {
		return
	}

	if string(response) != "true" {
		err = errors.New(string(response))

		return
	}

	return
}

// ChecksIsAcceptUserAuth Checks  Accept User Authentication.
func ChecksIsAcceptUserAuth(userPhone string) (userProfile *User, err error) {
	confirm := &confirmUser{PhoneNumber: userPhone}
	status, err := confirm.inAccept()
	if err != nil {
		return
	}
	if !status {
		return
	}
	hash, err := confirm.parseHash()
	if err != nil {
		return
	}

	if err = confirm.isConfirm(hash); err != nil {
		return
	}
	userProfile, err = confirm.parseUserProfile()

	return
}

func (c *confirmUser) inAccept() (status bool, err error) {
	status = false
	check, err := c.httpClient(fmt.Sprintf(loginURL, bot, domainURL), http.MethodPost, nil)
	if err != nil {
		return
	}
	switch string(check) {
	case declined:
		err = errors.New(declined)
	case trueStr:
		status = true
	}

	return
}

func (c *confirmUser) parseHash() (hash string, err error) {
	parseConfirm, err := c.httpClient(fmt.Sprintf(authURL, bot, domainURL), http.MethodGet, nil)
	if err != nil {
		return
	}
	confirm := regexp.MustCompile(regxConfirm)
	hash = confirm.FindString(string(parseConfirm))

	return
}

func (c *confirmUser) isConfirm(hash string) (err error) {
	_, err = c.httpClient(fmt.Sprintf(confirmHash, bot, domainURL, hash), http.MethodGet, nil)

	return
}

func (c *confirmUser) parseUserProfile() (u *User, err error) {
	responseUserRaw, err := c.httpClient(fmt.Sprintf(confirmURL, bot, domainURL), http.MethodGet, nil)
	if err != nil {
		return
	}
	RegxUserRaw := regexp.MustCompile(regxUserJSON)
	UserRaw := RegxUserRaw.FindString(string(responseUserRaw))
	err = json.Unmarshal([]byte(UserRaw), &u)
	if err == nil {
		u.Phone = c.PhoneNumber
	}

	return
}

func (c *confirmUser) httpClient(url, method string, body *string) (responseBody []byte, err error) {
	client := http.Client{}
	request, err := http.NewRequestWithContext(context.Background(), method, url, nil)
	if body != nil {
		request.Body = ioutil.NopCloser(strings.NewReader(*body))
	}
	if err != nil {
		return
	}
	cookie, is := getCookies(c.PhoneNumber)
	if is {
		for key, value := range cookie {
			request.AddCookie(&http.Cookie{
				Name:  key,
				Value: value,
			})
		}
	}

	request.Header.Add("origin", domainURL)
	request.Header.Add("referer", domainURL)

	request.Header.Add("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4606.81 Safari/537.36")
	if method == http.MethodPost {
		request.Header.Add("content-type", "application/x-www-form-urlencoded")
	}

	response, err := client.Do(request)
	if err == nil && response.StatusCode == 302 {
		return c.httpClient(oauthURL+response.Header.Get("Location"), method, body)
	}
	if err != nil {
		return
	}
	defer response.Body.Close() //nolint:errcheck

	for _, c := range response.Cookies() {
		if c.Value != cookiesDELETE {
			cookie[c.Name] = c.Value
		}
	}
	setCookies(c.PhoneNumber, cookie)

	responseBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	return
}
