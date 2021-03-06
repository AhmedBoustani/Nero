package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Command struct {
	Token	     string
	TeamID     string
	TeamDomain string

	ChannelID   string
	ChannelName string

	UserID   string
	UserName string

	Command string
	Text    string

	ResponseURL *url.URL
}

type Attachment struct {
	Pretext string `json:"pretext"`
	Text 		string `json:"text"`
}

func Give(w http.ResponseWriter, r *http.Request) {
	u, err := parseRequest(r)

	if err != nil {
			http.Error(w, err.Error(), 400)
			return
	}

	if !isLegit(u) {
		http.Error(w, "This request is not from an authorized Slack app", 403)
		return
	}

	s := strings.Split(u.Text, " ")
	user := s[0]

	if user[1:] == u.UserName {
    http.Error(w, "You can't give Nero to yourself!", 400)
    return
  }

	i := 1
	amount := 1
	if len(s) > 1 {
		amount, err = strconv.Atoi(s[1])
		if err != nil {
			amount = 1
			i = 1
		} else {
			i = 2
		}
	}

	if amount > NEROLIMIT {
		w.Write([]byte("You're giving more than what is allowed."))
		return
	}

	reason := strings.Join(s[i:], " ")
	dbUser := strings.Replace(user, "@", "", 1)
	go AddNero(dbUser, amount)

	msg := fmt.Sprintf("*@%s* gave you *%d Nero*", u.UserName, amount)
	go sendMsg(msg, user, reason)

	sMsg := fmt.Sprintf("You gave *%d Nero* to *%s*", amount, user)
	go sendMsg(sMsg, u.ChannelID, reason)

	w.Write([]byte(""))
}

func GetScore(w http.ResponseWriter, r *http.Request) {
	u, err := parseRequest(r)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if !isLegit(u) {
		http.Error(w, "This request is not from an authorized Slack app", 403)
		return
	}

	amount, err := GetNero(u.UserName)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte(fmt.Sprintf("You have *%d Nero*", amount)))
}

func GetAllScores(w http.ResponseWriter, r *http.Request) {
	res := db.FindAll()
	for e := range res {
		w.Write([]byte(fmt.Sprintf("%s %d\n", res[e].User, res[e].Amount)))
	}
}

func GetAllRemaining(w http.ResponseWriter, r *http.Request) {
	res := rem.FindAll()
	for e := range res {
		w.Write([]byte(fmt.Sprintf("%s %d\n", res[e].User, res[e].Amount)))
	}
}

func commandFromValues(v url.Values) (Command, error) {
	u, err := url.Parse(v.Get("response_url"))
	if err != nil {
		return Command{}, err
	}

	return Command{
		Token:       v.Get("token"),
		TeamID:      v.Get("team_id"),
		TeamDomain:  v.Get("team_domain"),
		ChannelID:   v.Get("channel_id"),
		ChannelName: v.Get("channel_name"),
		UserID:      v.Get("user_id"),
		UserName:    v.Get("user_name"),
		Command:     v.Get("command"),
		Text:        v.Get("text"),
		ResponseURL: u,
	}, nil
}

func parseRequest(r *http.Request) (Command, error) {
	err := r.ParseForm()
	if err != nil {
		return Command{}, err
	}
	return commandFromValues(r.Form)
}

func sendMsg(msg string, rec string, att string) {
	token := ENV.SlackAccessToken
	form := url.Values{}
	form.Add("text", msg)
	form.Add("channel", rec)
	form.Add("token", token)

	a := append([]Attachment{}, Attachment{ Text: att })
	out, err := json.Marshal(a)
	if err != nil {
		log.Fatal(err)
	}
	form.Add("attachments", string(out))

	http.Post("https://slack.com/api/chat.postMessage", "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(form.Encode())))
}

func isLegit(c Command) bool {
	return c.Token == ENV.SlackVerToken
}
