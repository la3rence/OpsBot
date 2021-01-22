package handler

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func getGitHubClient() *github.Client {
	token := os.Getenv("BOT_TOKEN")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

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
		action := *e.Action

		if action == "edited" || action == "created" {
			client := getGitHubClient()
			labelName := *e.GetComment().Body
			labels := []string{labelName}
			issue, response, err := client.Issues.AddLabelsToIssue(context.Background(), *e.GetRepo().Owner.Login,
				*e.GetRepo().Name,
				*e.GetIssue().Number, labels)
			if err != nil {
				log.Print(err)
			}
			log.Println(response, issue)
			fmt.Fprintf(w, "ok")
		}

	default:
		log.Printf("unknown event type %s\n", github.WebHookType(r))
		return
	}
	//fmt.Fprintf(w, "<h1>Hello from Go!</h1>")
}
