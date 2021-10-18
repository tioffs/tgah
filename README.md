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

import (
	"context"
	"fmt"
	"time"

	"github.com/tioffs/tgah"
)

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
	if confirm := tgah.SendPhoneTelegram(context.Background(), phone, nil);
		confirm.Error != nil || confirm.Status != tgah.Success {
		panic(confirm.Error)
	}
	// check accept user is auth you bot
	for {
		<-time.After(3 * time.Second)
		confirm := tgah.ChecksIsAcceptUserAuth(context.Background(), phone, nil)
		if confirm.Error != nil {
			panic(confirm.Error)
		}
		switch confirm.Status {
		case tgah.Success:
			fmt.Println(confirm.User)
			break
		case tgah.Cancel:
			panic(confirm.Error)
		case tgah.Pending:
			println(tgah.Pending)
		}
	}
}
```
