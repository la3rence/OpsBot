package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v85/github"
	"github.com/la3rence/OpsBot/utils"
	"golang.org/x/oauth2"
)

const botName = "k8s-ci-bot"

const (
	Label   = "/label"
	UnLabel = "/un-label"
	LGTM    = "/lgtm" // rebase
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
	"renovate":    "dependencies",
	"release":     "release",
	"test":        "test",
	"doc":         "documentation",
	"readme":      "documentation",
	"wip":         "wip",
	"refactor":    "refactor",
	"bug":         "bug",
	"功能":          "feature",
	"报错":          "bug",
	"优化":          "enhancement",
	"请求":          "feature",
	"测试":          "test",
	"依赖":          "dependencies",
	"升级":          "dependencies",
	"国际化":         "i18n",
	"i18n":        "i18n",
	"性能":          "enhancement",
	"?":           "question",
	"？":           "question",
}

var labelColorMapping = map[string]string{
	"feature":       "5319E7",
	"bug":           "ff0000",
	"test":          "006B75",
	"ci":            "006B75",
	"release":       "0075ca",
	"wip":           "FBCA04",
	"refactor":      "5319E7",
	"documentation": "0075ca",
	"dependencies":  "0075ca",
	"enhancement":   "a2eeef",
	"fixed":         "0E8A16",
	"approved":      "0E8A16",
	"todo":          "FBCA04",
	"label":         "5319E7",
	"default":       "FBCA04",
}

var ctx = context.Background()
var secret = os.Getenv("WEBHOOK_SECRET")

func getGitHubClient() (*github.Client, error) {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("BOT_TOKEN environment variable is not set")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc), nil
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
	githubClient, err := getGitHubClient()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("error getting GitHub client: %s\n", err)
		return
	}

	// Add proper error handling for all operations
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}()

	switch e := event.(type) {
	case *github.PushEvent:
		// this is a commit push, do something with it
	case *github.PullRequestEvent:
		pullRequestEvent := *e
		if err := addLabelIfPROpen(githubClient, pullRequestEvent); err != nil {
			log.Printf("Error adding label to PR: %v", err)
		}
		if err := requestReviewIfPROpen(githubClient, pullRequestEvent); err != nil {
			log.Printf("Error requesting review for PR: %v", err)
		}
	case *github.PullRequestReviewEvent:
		if err := addApproveLabelIfApproved(githubClient, *e); err != nil {
			log.Printf("Error adding approve label: %v", err)
		}
	case *github.IssuesEvent:
		issueEvent := *e
		if err := addLabelIfIssueOpen(githubClient, issueEvent); err != nil {
			log.Printf("Error adding label to issue: %v", err)
		}
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
			if issueCommentEvent.GetSender().GetLogin() == botName {
				_, _ = fmt.Fprintf(w, "ok")
				return
			}
			if strings.Contains(commentBody, Label) {
				if err := addLabelsByComment(commentBody, githubClient, issueCommentEvent); err != nil {
					log.Printf("Error adding labels: %v", err)
				}
			}
			if strings.Contains(commentBody, UnLabel) {
				if err := removeLabelFromIssue(commentBody, githubClient, issueCommentEvent); err != nil {
					log.Printf("Error removing labels: %v", err)
				}
			}
			if strings.Contains(commentBody, Approve) {
				if err := approvePullRequest(githubClient, issueCommentEvent); err != nil {
					log.Printf("Error approving PR: %v", err)
				}
			}
			if strings.Contains(commentBody, LGTM) {
				if err := mergePullRequest(githubClient, issueCommentEvent, "rebase"); err != nil {
					log.Printf("Error merging PR with rebase: %v", err)
				}
			}
			if strings.Contains(commentBody, Merge) {
				if err := mergePullRequest(githubClient, issueCommentEvent, "merge"); err != nil {
					log.Printf("Error merging PR: %v", err)
				}
			}
			if strings.Contains(commentBody, Close) {
				if err := closeOrOpenIssue(githubClient, issueCommentEvent, false); err != nil {
					log.Printf("Error closing issue: %v", err)
				}
			}
			if strings.Contains(commentBody, Reopen) || strings.Contains(commentBody, ReOpen) {
				if err := closeOrOpenIssue(githubClient, issueCommentEvent, true); err != nil {
					log.Printf("Error reopening issue: %v", err)
				}
			}
			if strings.Contains(commentBody, Update) {
				if err := updatePullRequest(githubClient, issueCommentEvent); err != nil {
					log.Printf("Error updating PR: %v", err)
				}
			}
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

func updatePullRequest(client *github.Client, issueCommentEvent github.IssueCommentEvent) error {
	ackByReaction(client, issueCommentEvent)
	repo := *issueCommentEvent.GetRepo().Name
	owner := *issueCommentEvent.GetRepo().Owner.Login
	number := *issueCommentEvent.GetIssue().Number
	pullRequest, _, err := client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return fmt.Errorf("failed to get pull request: %w", err)
	}
	sourceBranchSha := pullRequest.GetHead().GetSHA()
	// https://docs.github.com/cn/rest/reference/pulls#update-a-pull-request-branch
	_, res, err := client.PullRequests.UpdateBranch(ctx, owner, repo, number,
		&github.PullRequestBranchUpdateOptions{
			ExpectedHeadSHA: &sourceBranchSha,
		})
	if err != nil {
		if res != nil && res.StatusCode == 202 {
			// sendCommentWithDetailsDom(client, owner, repo, number, "Updating Accepted", err.Error()+"<br>"+res.Status)
			fmt.Println("Updating Accepted")
		} else {
			sendCommentWithDetailsDom(client, owner, repo, number, "Error", err.Error())
		}
		return fmt.Errorf("failed to update PR branch: %w", err)
	}
	return nil
}

func approvePullRequest(client *github.Client, event github.IssueCommentEvent) error {
	approveEventName := "APPROVE"
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	issueNumber := event.GetIssue().GetNumber()
	review, _, err := client.PullRequests.CreateReview(ctx, owner, repo, issueNumber,
		&github.PullRequestReviewRequest{
			Event: &approveEventName,
		})
	if err != nil {
		log.Println("CreateReview Error" + err.Error())
		return fmt.Errorf("failed to create review: %w", err)
	}

	submitReview, _, err := client.PullRequests.SubmitReview(ctx, owner, repo, issueNumber,
		review.GetID(),
		&github.PullRequestReviewRequest{
			Event: &approveEventName,
		},
	)
	if err != nil {
		log.Println("SubmitReview Error" + err.Error())
		return fmt.Errorf("failed to submit review: %w", err)
	}

	log.Println(submitReview)
	// labels := []string{"approved"}
	// addLabelsToIssue(labels, client, owner, repo, issueNumber)
	return nil
}

func addLabelsByComment(commentBody string, githubClient *github.Client, issueCommentEvent github.IssueCommentEvent) error {
	ackByReaction(githubClient, issueCommentEvent)
	labelsToAdd := utils.GetTagNextAllParams(commentBody, Label)
	err := addLabelsToIssue(labelsToAdd, githubClient,
		issueCommentEvent.GetRepo().GetOwner().GetLogin(),
		issueCommentEvent.GetRepo().GetName(),
		issueCommentEvent.GetIssue().GetNumber())
	if err != nil {
		return fmt.Errorf("failed to add labels by comment: %w", err)
	}
	return nil
}

func addLabelsToIssue(labelsToAdd []string, githubClient *github.Client, owner string, repo string, issueNumber int) error {
	// check if label exists, if yes, add it
	labels, _, err := githubClient.Issues.ListLabelsByIssue(ctx, owner, repo, issueNumber, nil)
	if err != nil {
		return fmt.Errorf("failed to list labels for issue: %w", err)
	}
	for _, param := range labelsToAdd {
		labelExists := false
		for _, label := range labels {
			if label.GetName() == param {
				labelExists = true
				break
			}
		}
		// if not, create by api and add it.
		if !labelExists {
			color := labelColorMapping[param]
			if color == "" {
				color = labelColorMapping["default"]
			}
			_, _, err := githubClient.Issues.CreateLabel(ctx, owner, repo, &github.Label{
				Name:  &param,
				Color: &color,
			})
			if err != nil {
				log.Printf("Failed to create label %s: %v", param, err)
			}
		}
	}
	issue, response, githubErr := githubClient.Issues.AddLabelsToIssue(ctx, owner, repo, issueNumber, labelsToAdd)
	if githubErr != nil {
		return fmt.Errorf("failed to add labels to issue: %w", githubErr)
	}
	log.Println(response, issue, githubErr)
	return nil
}

func removeLabelFromIssue(commentBody string, githubClient *github.Client, issueCommentEvent github.IssueCommentEvent) error {
	ackByReaction(githubClient, issueCommentEvent)
	params := utils.GetTagNextAllParams(commentBody, UnLabel)
	for _, param := range params {
		response, githubErr := githubClient.Issues.RemoveLabelForIssue(ctx,
			*issueCommentEvent.GetRepo().Owner.Login,
			*issueCommentEvent.GetRepo().Name,
			*issueCommentEvent.GetIssue().Number,
			param)
		if githubErr != nil {
			log.Printf("Error removing label %s: %v", param, githubErr)
			return fmt.Errorf("failed to remove label %s: %w", param, githubErr)
		}
		log.Println(response)
	}
	return nil
}

func requestReviewIfPROpen(githubClient *github.Client, pullRequestEvent github.PullRequestEvent) error {
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
			return fmt.Errorf("failed to request reviewers: %w", err)
		}
		log.Println(response, reviewers)
	}
	return nil
}

func addLabelIfPROpen(githubClient *github.Client, pullRequestEvent github.PullRequestEvent) error {
	action := *pullRequestEvent.Action
	title := pullRequestEvent.GetPullRequest().GetTitle()
	if action == "edited" || action == "opened" {
		for titleKey, labelValue := range titleLabelMapping {
			if strings.Contains(strings.ToLower(title), strings.ToLower(titleKey)) {
				err := addLabelsToIssue([]string{labelValue}, githubClient,
					*pullRequestEvent.GetRepo().Owner.Login,
					*pullRequestEvent.GetRepo().Name,
					*pullRequestEvent.GetPullRequest().Number)
				if err != nil {
					return fmt.Errorf("failed to add label to PR: %w", err)
				}
			}
		}
	}
	return nil
}

func addApproveLabelIfApproved(githubClient *github.Client, reviewEvent github.PullRequestReviewEvent) error {
	review := *reviewEvent.Review
	state := review.GetState()
	approvedString := "approved"
	if state == approvedString {
		labels := []string{approvedString}
		owner := reviewEvent.GetRepo().GetOwner().GetLogin()
		repo := reviewEvent.GetRepo().GetName()
		issueNumber := reviewEvent.GetPullRequest().GetNumber()
		err := addLabelsToIssue(labels, githubClient, owner, repo, issueNumber)
		if err != nil {
			return fmt.Errorf("failed to add approve label: %w", err)
		}
	}
	return nil
}

func addLabelIfIssueOpen(githubClient *github.Client, issuesEvent github.IssuesEvent) error {
	action := *issuesEvent.Action
	title := issuesEvent.GetIssue().GetTitle()
	if action == "edited" || action == "opened" {
		for titleKey, labelValue := range titleLabelMapping {
			if strings.Contains(strings.ToLower(title), strings.ToLower(titleKey)) {
				err := addLabelsToIssue([]string{labelValue}, githubClient,
					*issuesEvent.GetRepo().Owner.Login,
					*issuesEvent.GetRepo().Name,
					*issuesEvent.GetIssue().Number)
				if err != nil {
					return fmt.Errorf("failed to add label to issue: %w", err)
				}
			}
		}
	}
	return nil
}

func mergePullRequest(githubClient *github.Client, issueCommentEvent github.IssueCommentEvent, mergeMethod string) error {
	ackByReaction(githubClient, issueCommentEvent)
	owner := issueCommentEvent.GetRepo().GetOwner().GetLogin()
	senderName := issueCommentEvent.GetSender().GetLogin()
	repo := issueCommentEvent.GetRepo().GetName()
	number := issueCommentEvent.GetIssue().GetNumber()
	if owner != senderName {
		sendComment(githubClient, owner, repo, number,
			fmt.Sprintf("Sorry, @%s - This pull request can only be merged by the owner (@%s).", senderName, owner))
		return nil
	}
	mergedBefore, _, err := githubClient.PullRequests.IsMerged(ctx, owner, repo, number)
	if err != nil {
		return fmt.Errorf("failed to check if PR is merged: %w", err)
	}
	mergeComment := fmt.Sprintf("PR #%d was merged. Thanks for your contribution.", number)
	commitMsg := fmt.Sprintf("merge: PR(#%d)", number)
	if mergedBefore {
		log.Println(mergeComment)
		sendComment(githubClient, owner, repo, number, mergeComment)
	} else {
		log.Printf("start to %s\n", commitMsg)
		mergeResult, _, err := githubClient.PullRequests.Merge(ctx, owner, repo, number, commitMsg, &github.PullRequestOptions{
			MergeMethod: mergeMethod, // optional with string: "merge", "squash", and "rebase"
		})
		if err != nil {
			log.Println(err)
			sendCommentWithDetailsDom(githubClient, owner, repo, number, "Error", err.Error())
			return fmt.Errorf("failed to merge PR: %w", err)
		} else {
			log.Println(mergeResult)
			merged := mergeResult.GetMerged()
			if merged {
				log.Println(mergeComment)
				sendComment(githubClient, owner, repo, number, mergeComment)
			} else {
				failMsg := fmt.Sprintf("Fail to merge this PR #%d", number)
				sendCommentWithDetailsDom(githubClient, owner, repo, number, "Debug", failMsg)
				log.Println(failMsg)
				return fmt.Errorf("PR merge result was not successful: %s", failMsg)
			}
		}
	}
	return nil
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

func closeOrOpenIssue(githubClient *github.Client, issueCommentEvent github.IssueCommentEvent, open bool) error {
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
	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to close or open issue: %w", err)
	}
	log.Println(response)
	log.Println(edit)
	return nil
}
