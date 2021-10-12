# TGAH - telegram Authorization
<a href="https://opensource.org/licenses/MIT" style="text-decoration: none">
<img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square" alt="License: MIT">
</a>

**Example of authorization in telegram without using a [widget](https://core.telegram.org/widgets/login)**

### Installation
- ```bash
    go get -d github.com/tioffs/tgah@master
    ```
- **[Setting up a bot](https://core.telegram.org/widgets/login#setting-up-a-bot)**
### Easy to use
**example http.Server: [`example`](./example)**

```go
package main

import "github.com/tioffs/tgah"

func main() {
	// phone number no +
	phone := "79000000000"
	// you bot ID
	botID := 1234567899
	// you bot domain
	domain := "yousite.com"
	// set setting
	tgah.Setting(int32(botID), domain)
	// send push notify user (Auth)
	if err := tgah.SendPhoneTelegram(phone); err != nil {
		panic(err)
	}
	// check accept user is auth you bot
	user, err := tgah.ChecksIsAcceptUserAuth(phone)
	if err != nil {
		panic(err)
	}
	println(user)
}
```

#### User profile
```go
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int    `json:"auth_date"`
	Hash      string `json:"hash"`
	Phone     string `json:"phone"`
}
```
