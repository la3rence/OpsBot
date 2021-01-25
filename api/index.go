package handler

import (
	"OpsBot/utils"
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	LABEL   = "/label"
	UNLABEL = "/un-label"
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
	githubClient := getGitHubClient()
	switch e := event.(type) {
	case *github.PushEvent:
		// this is a commit push, do something with it
	case *github.PullRequestEvent:
		pullRequestEvent := *e
		requestReview(githubClient, pullRequestEvent)
	case *github.WatchEvent:
		if e.Action != nil && *e.Action == "starred" {
			fmt.Printf("%s starred repository %s\n",
				*e.Sender.Login, *e.Repo.FullName)
		}
	case *github.IssueCommentEvent:
		fmt.Printf("IssueCommentEvent: %s\n", *e.Action)
		commentBody := *e.GetComment().Body
		action := *e.Action
		if action == "edited" || action == "created" {
			issueCommentEvent := *e
			if strings.Contains(commentBody, LABEL) {
				addLabelsToIssue(commentBody, githubClient, issueCommentEvent)
			}
			if strings.Contains(commentBody, UNLABEL) {
				removeLabelFromIssue(commentBody, githubClient, issueCommentEvent)
			}
		}
	default:
		log.Printf("unknown event type %s\n", github.WebHookType(r))
		return
	}
	_, err = fmt.Fprintf(w, "ok")
	if err != nil {
		log.Fatal("Write response error ", err)
	}
}

func addLabelsToIssue(commentBody string, githubClient *github.Client, issueCommentEvent github.IssueCommentEvent) {
	// 获取 /label 后的 字符串，注意越界问题
	wordArray := strings.Fields(commentBody)
	labelIndex := utils.StringIndexOf(wordArray, LABEL)
	if len(wordArray) > labelIndex+1 {
		labelName := wordArray[labelIndex+1]
		if labelName != "" {
			labels := []string{labelName}
			issue, response, err := githubClient.Issues.AddLabelsToIssue(context.Background(), *issueCommentEvent.GetRepo().Owner.Login,
				*issueCommentEvent.GetRepo().Name,
				*issueCommentEvent.GetIssue().Number, labels)
			if err != nil {
				log.Print(err)
			}
			log.Println(response, issue)
		}
	}
}

func removeLabelFromIssue(commentBody string, githubClient *github.Client, issueCommentEvent github.IssueCommentEvent) {
	wordArray := strings.Fields(commentBody)
	unLabelIndex := utils.StringIndexOf(wordArray, UNLABEL)
	if len(wordArray) > unLabelIndex+1 {
		unLabelName := wordArray[unLabelIndex+1]
		if unLabelName != "" {
			response, err := githubClient.Issues.RemoveLabelForIssue(context.Background(),
				*issueCommentEvent.GetRepo().Owner.Login,
				*issueCommentEvent.GetRepo().Name,
				*issueCommentEvent.GetIssue().Number,
				unLabelName)
			if err != nil {
				log.Print(err)
			}
			log.Println(response)
		}
	}
}

func requestReview(githubClient *github.Client, pullRequestEvent github.PullRequestEvent) {
	action := *pullRequestEvent.Action
	if action == "opened" || action == "reopened" {
		reviewers, response, err := githubClient.PullRequests.RequestReviewers(context.Background(),
			*pullRequestEvent.GetRepo().Owner.Login,
			*pullRequestEvent.GetRepo().Name,
			*pullRequestEvent.GetPullRequest().Number,
			github.ReviewersRequest{
				Reviewers: []string{*pullRequestEvent.GetRepo().Owner.Login},
			},
		)
		if err != nil {
			log.Print(err)
		}
		log.Println(response, reviewers)
	}

}
