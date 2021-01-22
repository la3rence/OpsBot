package handler

import (
	"fmt"
	"github.com/google/go-github/github"
	"io/ioutil"
	"log"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: err=%s\n", err)
		return
	}
	defer r.Body.Close()
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("could not parse webhook: err=%s\n", err)
		return
	}
	switch e := event.(type) {
	case *github.PushEvent:
		// this is a commit push, do something with it
	case *github.PullRequestEvent:
		// this is a pull request, do something with it
	case *github.WatchEvent:
		// https://developer.github.com/v3/activity/events/types/#watchevent
		// someone starred our repository
		if e.Action != nil && *e.Action == "starred" {
			fmt.Printf("%s starred repository %s\n",
				*e.Sender.Login, *e.Repo.FullName)
		}
	case *github.IssueCommentEvent:
		fmt.Println("IssueCommentEvent: ")
		fmt.Println(*e.Action)
		fmt.Println(*e.GetRepo().Name)
		fmt.Println(*e.GetIssue().Number)
		fmt.Println(*e.GetComment().Body)
		// todo: call GitHub api
	default:
		log.Printf("unknown event type %s\n", github.WebHookType(r))
		return
	}
	//fmt.Fprintf(w, "<h1>Hello from Go!</h1>")
}
