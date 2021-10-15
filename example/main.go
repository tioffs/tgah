package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/tioffs/tgah"
)

const templatesTest = `
<html><head></head><body>
<div>
<input type="text" name="phone" placeholder="79000000000"/>
</div>
<div>
<button type="button" onclick="send()">Submit</button>
</div>
<div id="user"></div>
<script>
var phone = ""
var button = document.querySelector('button');
var input = document.querySelector('input[name="phone"]');
function send(){
phone = input.value
fetch("/telegram?act=send&phone="+phone, {}).then(r=> r.json().then(data => {
if (data.status == "success") {
	button.setAttribute("disabled", true)
	input.setAttribute("disabled", true)
	check()
}
if (data.error) {alert(data.error)}
})).catch(error => { alert(error) })}

function check(){
fetch("/telegram?act=check&phone="+phone, {}).then(r=> r.json().then(data => {
if (data.error) {alert(data.error)}
if (data.status == "cancel") {alert(data.error + data.status)}
if (data.status == "pending") {setTimeout(check, 3000)}
if (data.status == "success") {document.querySelector('#user').innerHTML = JSON.stringify(data)}
})).catch(error => { alert(error) })}

</script>

</body></html>
`

func main() {
	// replace botID and domain
	// tgah.Setting(botID, domain)
	tgah.Setting(12345678, "youdomain.com")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(templatesTest))
	})
	http.HandleFunc("/telegram", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("act") {
		case "send":
			confirm := tgah.SendPhoneTelegram(r.Context(), r.URL.Query().Get("phone"))
			ResponseJSON(w, confirm, http.StatusOK)
		case "check":
			confirm := tgah.ChecksIsAcceptUserAuth(r.Context(), r.URL.Query().Get("phone"))
			ResponseJSON(w, confirm, http.StatusOK)
		}
	})

	_ = http.ListenAndServe("localhost:3000", nil)
}

// ResponseJSON encode struct to json string.
func ResponseJSON(w http.ResponseWriter, v interface{}, status ...int) {
	code := http.StatusOK

	if len(status) > 0 {
		code = status[0]
	}

	if err, ok := v.(error); ok || code >= 400 {
		v = map[string]interface{}{
			"error": err.Error(),
		}

		if code == http.StatusOK {
			code = http.StatusInternalServerError
		}
	}

	strJSON, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
	w.Header().Set(`Content-Type`, `application/json`)
	w.WriteHeader(code)
	if _, err = w.Write(strJSON); err != nil {
		log.Println(err.Error())
	}
}
