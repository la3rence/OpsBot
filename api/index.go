package handler

import (
	"context"
	"fmt"
	"github.com/Lonor/OpsBot/utils"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"strings"
)

const botName = "k8s-ci-bot"

const (
	Label   = "/label"
	UnLabel = "/un-label"
	LGTM    = "/lgtm"
	Merge   = "/merge"
	Close   = "/close"
	Reopen  = "/reopen"
	ReOpen  = "/re-open"
	Approve = "/approve"
	Update  = "/update"
)

// https://www.conventionalcommits.org/zh-hans/v1.0.0/
var titleLabelMapping = map[string]string{
	"enhancement": "enhancement",
	"fix":         "enhancement",
	"ci":          "ci",
	"feat":        "feature",
	"bump":        "dependencies",
	"deps":        "dependencies",
	"dependency":  "dependencies",
	"release":     "release",
	"test":        "test",
	"doc":         "documentation",
	"readme":      "documentation",
	"wip":         "wip",
	"refactor":    "refactor",
	"bug":         "bug",
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
			// avoid recursion comment by bot
			if issueCommentEvent.GetSender().GetName() == botName {
				_, _ = fmt.Fprintf(w, "ok")
				return
			}
			if strings.Contains(commentBody, Label) {
				addLabelsToIssue(commentBody, githubClient, issueCommentEvent)
			}
			if strings.Contains(commentBody, UnLabel) {
				removeLabelFromIssue(commentBody, githubClient, issueCommentEvent)
			}
			if strings.Contains(commentBody, Approve) {
				approvePullRequest(githubClient, issueCommentEvent)
			}
			if strings.Contains(commentBody, LGTM) || strings.Contains(commentBody, Merge) {
				mergePullRequest(githubClient, issueCommentEvent)
			}
			if strings.Contains(commentBody, Close) {
				closeOrOpenIssue(githubClient, issueCommentEvent, false)
			}
			if strings.Contains(commentBody, Reopen) || strings.Contains(commentBody, ReOpen) {
				closeOrOpenIssue(githubClient, issueCommentEvent, true)
			}
			//if strings.Contains(commentBody, Update) {
			//	updatePullRequest(githubClient, issueCommentEvent)
			//}
		}
	default:
		log.Printf("unknown event type %s\n", github.WebHookType(r))
		_, _ = fmt.Fprintf(w, "ok")
		return
	}
	_, _ = fmt.Fprintf(w, "ok")
}

// ackByReaction ACK with reaction 👍
func ackByReaction(client *github.Client, issueCommentEvent github.IssueCommentEvent) {
	repo := *issueCommentEvent.GetRepo().Name
	owner := *issueCommentEvent.GetRepo().Owner.Login
	commentId := issueCommentEvent.GetComment().GetID()
	_, _, _ = client.Reactions.CreateIssueCommentReaction(ctx, owner, repo, commentId, "+1")
}

func updatePullRequest(client *github.Client, issueCommentEvent github.IssueCommentEvent) {
	ackByReaction(client, issueCommentEvent)
	repo := *issueCommentEvent.GetRepo().Name
	owner := *issueCommentEvent.GetRepo().Owner.Login
	number := *issueCommentEvent.GetIssue().Number
	pullRequest, _, _ := client.PullRequests.Get(ctx, owner, repo, number)
	sourceBranchSha := pullRequest.GetHead().GetSHA()
	// https://docs.github.com/cn/rest/reference/pulls#update-a-pull-request-branch
	_, res, err := client.PullRequests.UpdateBranch(ctx, owner, repo, number,
		&github.PullRequestBranchUpdateOptions{
			ExpectedHeadSHA: &sourceBranchSha,
		})
	if err != nil {
		log.Println("Update branch may has error " + err.Error())
		if res != nil && res.StatusCode == 202 {
			sendCommentWithDetailsDom(client, owner, repo, number, "Updating Accepted", err.Error()+"<br>"+res.Status)
		} else {
			sendCommentWithDetailsDom(client, owner, repo, number, "Error", err.Error())
		}
	}
}

func approvePullRequest(client *github.Client, event github.IssueCommentEvent) {
	ackByReaction(client, event)
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
	ackByReaction(githubClient, issueCommentEvent)
	params := utils.GetTagNextAllParams(commentBody, Label)
	issue, response, githubErr := githubClient.Issues.AddLabelsToIssue(ctx, *issueCommentEvent.GetRepo().Owner.Login,
		*issueCommentEvent.GetRepo().Name,
		*issueCommentEvent.GetIssue().Number, params)
	log.Println(response, issue, githubErr)
}

func removeLabelFromIssue(commentBody string, githubClient *github.Client, issueCommentEvent github.IssueCommentEvent) {
	ackByReaction(githubClient, issueCommentEvent)
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
	ackByReaction(githubClient, issueCommentEvent)
	owner := issueCommentEvent.GetRepo().GetOwner().GetLogin()
	senderName := issueCommentEvent.GetSender().GetLogin()
	repo := issueCommentEvent.GetRepo().GetName()
	number := issueCommentEvent.GetIssue().GetNumber()
	if owner != senderName {
		sendComment(githubClient, owner, repo, number,
			fmt.Sprintf("Sorry, @%s - This pull request can only be merged by the owner (@%s).", senderName, owner))
		return
	}
	mergedBefore, _, _ := githubClient.PullRequests.IsMerged(ctx, owner, repo, number)
	mergeComment := fmt.Sprintf("PR #%d was merged (with rebase). Thanks for your contribution.", number)
	commitMsg := fmt.Sprintf("merge: PR(#%d)", number)
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
			sendCommentWithDetailsDom(githubClient, owner, repo, number, "Error", err.Error())
		} else {
			log.Println(mergeResult)
			merged := mergeResult.GetMerged()
			if merged {
				log.Printf(mergeComment)
				sendComment(githubClient, owner, repo, number, mergeComment)
			} else {
				failMsg := fmt.Sprintf("Fail to merge this PR #%d", number)
				sendCommentWithDetailsDom(githubClient, owner, repo, number, "Debug", failMsg)
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

func sendCommentWithDetailsDom(githubClient *github.Client, owner string, repo string, number int,
	detailSummary string, detailBody string) *github.IssueComment {
	return sendComment(githubClient, owner, repo, number,
		`<details><summary>`+detailSummary+`</summary><p>`+detailBody+`</p></details>`)
}

func closeOrOpenIssue(githubClient *github.Client, issueCommentEvent github.IssueCommentEvent, open bool) {
	ackByReaction(githubClient, issueCommentEvent)
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
