package tgah

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type status string

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
	// cleanupInterval clean cache time second.
	cleanupInterval = 10
	// expires default time second cache delete.
	expires = 60
	// cookiesDELETE const status cookie header.
	cookiesDELETE = "DELETED"
	// declined user Cancel auth.
	declined = "Declined by the user"
	// trueStr check string true.
	trueStr = "true"
	// Pending status.
	Pending status = "pending"
	// Cancel status.
	Cancel status = "cancel"
	// Success status.
	Success status = "success"
)

// user struct telegram oauth info user.
type user struct {
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

// Confirm struct status, error checks.
type Confirm struct {
	Phone   string  `json:"-"`
	User    *user   `json:"user,omitempty"`
	Status  status  `json:"status"`
	Error   *string `json:"error,omitempty"`
	hash    string
	context context.Context
}

// store memory auth telegram user.
var store = make(map[string]item)

// domainUrl site connect bot telegram auth.
var domainURL string

// bot id bot telegram.
var bot int32

// statusCache is auto delete cache not auth user.
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

// SendPhoneTelegram send push notify telegram user phone.
func SendPhoneTelegram(ctx context.Context, userPhone string) (confirm *Confirm) {
	confirm = &Confirm{Phone: userPhone, context: ctx, Status: Cancel}
	response, err := confirm.httpClient(fmt.Sprintf(sendPhoneURL, bot, domainURL), http.MethodPost, confirm.pointerString(fmt.Sprintf("phone=%s", userPhone)))
	if err != nil {
		confirm.Error = confirm.pointerString(err.Error())

		return
	}
	statusSendNotify, err := strconv.ParseBool(string(response))
	if err != nil {
		confirm.Error = confirm.pointerString(string(response))

		return
	}
	if !statusSendNotify {
		confirm.Error = confirm.pointerString(string(response))

		return
	}

	confirm.Status = Success

	return
}

// ChecksIsAcceptUserAuth Checks  Accept user Authentication.
func ChecksIsAcceptUserAuth(ctx context.Context, userPhone string) (confirm *Confirm) {
	confirm = &Confirm{Phone: userPhone, context: ctx}
	confirm.inAccept()
	if confirm.Status != Success {
		return
	}
	confirm.parseHash()
	if confirm.Status == Cancel {
		return
	}
	confirm.isConfirm()
	if confirm.Status == Cancel {
		return
	}
	confirm.parseUserProfile()

	return
}

func (c *Confirm) inAccept() {
	response, err := c.httpClient(fmt.Sprintf(loginURL, bot, domainURL), http.MethodPost, nil)
	if err != nil {
		c.Error = c.pointerString(err.Error())
		c.Status = Cancel

		return
	}
	switch string(response) {
	case declined:
		c.Status = Cancel
		c.Error = c.pointerString(string(response))
	case trueStr:
		c.Status = Success
	default:
		c.Status = Pending
	}
}

func (c *Confirm) parseHash() {
	parseConfirm, err := c.httpClient(fmt.Sprintf(authURL, bot, domainURL), http.MethodGet, nil)
	if err != nil {
		c.Error = c.pointerString(err.Error())
		c.Status = Cancel

		return
	}
	confirm := regexp.MustCompile(regxConfirm)
	c.hash = confirm.FindString(string(parseConfirm))
	if c.hash == "" {
		c.Status = Cancel
	}
	c.Status = Success
}

func (c *Confirm) isConfirm() {
	_, err := c.httpClient(fmt.Sprintf(confirmHash, bot, domainURL, c.hash), http.MethodGet, nil)
	if err != nil {
		c.Status = Cancel
		c.Error = c.pointerString(err.Error())
	}
	c.Status = Success
}

func (c *Confirm) parseUserProfile() {
	responseUserRaw, err := c.httpClient(fmt.Sprintf(confirmURL, bot, domainURL), http.MethodGet, nil)
	if err != nil {
		c.Error = c.pointerString(err.Error())
		c.Status = Cancel

		return
	}
	RegxUserRaw := regexp.MustCompile(regxUserJSON)
	UserRaw := RegxUserRaw.FindString(string(responseUserRaw))
	err = json.Unmarshal([]byte(UserRaw), &c.User)
	if err != nil {
		c.Error = c.pointerString(err.Error())
		c.Status = Cancel

		return
	}
	c.User.Phone = c.Phone
	c.Status = Success
}

func (c *Confirm) pointerString(s string) *string {
	return &s
}

func (c *Confirm) httpClient(url, method string, body *string) (responseBody []byte, err error) { //nolint:cyclop
	client := http.Client{}
	request, err := http.NewRequestWithContext(c.context, method, url, nil)
	if body != nil {
		request.Body = ioutil.NopCloser(strings.NewReader(*body))
	}
	if err != nil {
		return
	}
	cookie, is := getCookies(c.Phone)
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

	request.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4606.81 Safari/537.36")
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
	setCookies(c.Phone, cookie)

	responseBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	return
}
