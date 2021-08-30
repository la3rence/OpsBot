package handler

import (
	"OpsBot/utils"
	"context"
	"fmt"
	"github.com/google/go-github/v38/github"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	Label   = "/label"
	UnLabel = "/un-label"
	LGTM    = "/lgtm"
	Close   = "/close"
	Reopen  = "/reopen"
	Approve = "/approve"
)

// https://www.conventionalcommits.org/zh-hans/v1.0.0/
var titleLabelMapping = map[string]string{
	"fix":        "enhancement",
	"ci":         "ci",
	"feat":       "feature",
	"bump":       "dependencies",
	"deps":       "dependencies",
	"dependency": "dependencies",
	"release":    "release",
	"test":       "ci",
	"doc":        "documentation",
	"readme":     "documentation",
	"wip":        "wip",
	"refactor":   "refactor",
}

var ctx = context.Background()
var secret = os.Getenv("WEBHOOK_SECRET")

func getGitHubClient() *github.Client {
	token := os.Getenv("BOT_TOKEN")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// this Handler used to be the serverless function
func Handler(w http.ResponseWriter, r *http.Request) {
	payload, validateErr := github.ValidatePayload(r, []byte(secret))
	if validateErr != nil {
		http.Error(w, "The GitHub signature header is invalid.", http.StatusUnauthorized)
		log.Printf("invalid signature: %s\n", validateErr.Error())
		return
	}
	event, parseErr := github.ParseWebHook(github.WebHookType(r), payload)
	if parseErr != nil {
		http.Error(w, "The payload parsed failed", http.StatusInternalServerError)
		log.Printf("could not parse webhook: %s\n", parseErr)
		return
	}
	githubClient := getGitHubClient()
	switch e := event.(type) {
	case *github.PushEvent:
		// this is a commit push, do something with it
	case *github.PullRequestEvent:
		pullRequestEvent := *e
		addLabelIfPROpen(githubClient, pullRequestEvent)
		requestReviewIfPROpen(githubClient, pullRequestEvent)
	case *github.IssuesEvent:
		issueEvent := *e
		addLabelIfIssueOpen(githubClient, issueEvent)
	case *github.WatchEvent:
		if e.Action != nil && *e.Action == "starred" {
			fmt.Printf("%s starred repository %s\n",
				*e.Sender.Login, *e.Repo.FullName)
		}
	case *github.IssueCommentEvent:
		action := e.GetAction()
		fmt.Printf("IssueCommentEvent: %s\n", action)
		commentBody := e.GetComment().GetBody()
		if action == "edited" || action == "created" {
			issueCommentEvent := *e
			if strings.Contains(commentBody, Label) {
				addLabelsToIssue(commentBody, githubClient, issueCommentEvent)
			}
			if strings.Contains(commentBody, UnLabel) {
				removeLabelFromIssue(commentBody, githubClient, issueCommentEvent)
			}
			if strings.Contains(commentBody, Approve) {
				approvePullRequest(githubClient, issueCommentEvent)
			}
			if strings.Contains(commentBody, LGTM) {
				mergePullRequest(githubClient, issueCommentEvent)
			}
			if strings.Contains(commentBody, Close) {
				closeOrOpenIssue(githubClient, issueCommentEvent, false)
			}
			if strings.Contains(commentBody, Reopen) {
				closeOrOpenIssue(githubClient, issueCommentEvent, true)
			}
		}
	default:
		log.Printf("unknown event type %s\n", github.WebHookType(r))
		_, _ = fmt.Fprintf(w, "ok")
		return
	}
	_, _ = fmt.Fprintf(w, "ok")
}

func approvePullRequest(client *github.Client, event github.IssueCommentEvent) {
	approveEventName := "APPROVE"
	loginOwner := event.GetRepo().GetOwner().GetLogin()
	repoName := event.GetRepo().GetName()
	issueNumber := event.GetIssue().GetNumber()
	review, _, err := client.PullRequests.CreateReview(ctx, loginOwner, repoName, issueNumber,
		&github.PullRequestReviewRequest{
			Event: &approveEventName,
		})
	if err == nil {
		submitReview, _, _ := client.PullRequests.SubmitReview(ctx, loginOwner, repoName, issueNumber,
			review.GetID(),
			&github.PullRequestReviewRequest{
				Event: &approveEventName,
			},
		)
		log.Println(submitReview)
		labels := []string{"approved"}
		_, _, _ = client.Issues.AddLabelsToIssue(ctx, loginOwner, repoName, issueNumber, labels)
	} else {
		log.Println("CreateReview Error" + err.Error())
	}
}

func addLabelsToIssue(commentBody string, githubClient *github.Client, issueCommentEvent github.IssueCommentEvent) {
	params := utils.GetTagNextAllParams(commentBody, Label)
	issue, response, githubErr := githubClient.Issues.AddLabelsToIssue(ctx, *issueCommentEvent.GetRepo().Owner.Login,
		*issueCommentEvent.GetRepo().Name,
		*issueCommentEvent.GetIssue().Number, params)
	log.Println(response, issue, githubErr)
}

func removeLabelFromIssue(commentBody string, githubClient *github.Client, issueCommentEvent github.IssueCommentEvent) {
	params := utils.GetTagNextAllParams(commentBody, UnLabel)
	for _, param := range params {
		response, githubErr := githubClient.Issues.RemoveLabelForIssue(ctx,
			*issueCommentEvent.GetRepo().Owner.Login,
			*issueCommentEvent.GetRepo().Name,
			*issueCommentEvent.GetIssue().Number,
			param)
		if githubErr != nil {
			log.Print(githubErr)
		}
		log.Println(response)
	}
}

func requestReviewIfPROpen(githubClient *github.Client, pullRequestEvent github.PullRequestEvent) {
	action := *pullRequestEvent.Action
	if action == "opened" || action == "reopened" {
		reviewers, response, err := githubClient.PullRequests.RequestReviewers(ctx,
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

func addLabelIfPROpen(githubClient *github.Client, pullRequestEvent github.PullRequestEvent) {
	action := *pullRequestEvent.Action
	title := pullRequestEvent.GetPullRequest().GetTitle()
	if action == "edited" || action == "opened" {
		for titleKey, labelValue := range titleLabelMapping {
			if strings.Contains(strings.ToLower(title), strings.ToLower(titleKey)) {
				labels, response, err := githubClient.Issues.AddLabelsToIssue(ctx, *pullRequestEvent.GetRepo().Owner.Login,
					*pullRequestEvent.GetRepo().Name, *pullRequestEvent.GetPullRequest().Number, []string{labelValue})
				log.Println(response, labels, err)
			}
		}
	}
}

func addLabelIfIssueOpen(githubClient *github.Client, issuesEvent github.IssuesEvent) {
	action := *issuesEvent.Action
	title := issuesEvent.GetIssue().GetTitle()
	if action == "edited" || action == "opened" {
		for titleKey, labelValue := range titleLabelMapping {
			if strings.Contains(strings.ToLower(title), strings.ToLower(titleKey)) {
				labels, response, err := githubClient.Issues.AddLabelsToIssue(ctx, *issuesEvent.GetRepo().Owner.Login,
					*issuesEvent.GetRepo().Name, *issuesEvent.GetIssue().Number, []string{labelValue})
				log.Println(response, labels, err)
			}
		}
	}
}

func mergePullRequest(githubClient *github.Client, issueCommentEvent github.IssueCommentEvent) {
	owner := issueCommentEvent.GetRepo().GetOwner().GetLogin()
	repo := issueCommentEvent.GetRepo().GetName()
	number := issueCommentEvent.GetIssue().GetNumber()
	mergedBefore, _, _ := githubClient.PullRequests.IsMerged(ctx, owner, repo, number)
	mergeComment := fmt.Sprintf("PR #%d was merged (with rebase). Thanks for your contribution.", number)
	commitMsg := fmt.Sprintf("merge: PR(#%d)", number)
	failMsg := fmt.Sprintf("Fail to merge this PR #%d", number)
	senderName := issueCommentEvent.GetSender().GetLogin()
	if owner != senderName {
		sendComment(githubClient, owner, repo, number,
			fmt.Sprintf("Sorry. This pull request can only be merged by its owner (@%s).", owner))
		return
	}
	if mergedBefore {
		log.Printf(mergeComment)
		sendComment(githubClient, owner, repo, number, mergeComment)
	} else {
		log.Printf("start to " + commitMsg + "\n")
		mergeResult, _, err := githubClient.PullRequests.Merge(ctx, owner, repo, number, commitMsg, &github.PullRequestOptions{
			MergeMethod: "rebase",
		})
		if err != nil {
			log.Println(err)
			sendComment(githubClient, owner, repo, number, err.Error())
		} else {
			log.Println(mergeResult)
			merged := mergeResult.GetMerged()
			if merged {
				log.Printf(mergeComment)
				sendComment(githubClient, owner, repo, number, mergeComment)
			} else {
				sendComment(githubClient, owner, repo, number, failMsg)
				log.Printf(failMsg)
			}
		}
	}
}

func sendComment(githubClient *github.Client, owner string, repo string, number int, comment string) *github.IssueComment {
	log.Printf("send comment to %s/%s #%d : %s\n", owner, repo, number, comment)
	createdComment, _, err := githubClient.Issues.CreateComment(
		ctx, owner, repo, number, &github.IssueComment{
			Body: &comment,
		})
	if err == nil {
		return createdComment
	}
	return nil
}

func closeOrOpenIssue(githubClient *github.Client, issueCommentEvent github.IssueCommentEvent, open bool) {
	owner := issueCommentEvent.GetRepo().GetOwner().GetLogin()
	repo := issueCommentEvent.GetRepo().GetName()
	number := issueCommentEvent.GetIssue().GetNumber()
	var state string
	if open {
		state = "open"
	} else {
		state = "closed"
	}
	edit, response, err := githubClient.Issues.Edit(ctx, owner, repo, number, &github.IssueRequest{
		State: &state,
	})
	if err == nil {
		log.Println(response)
		log.Println(edit)
	}
}
