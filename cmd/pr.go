package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gyurisc/adocli/internal/api"
	"github.com/spf13/cobra"
)

var prCmd = &cobra.Command{
	Use:     "pr",
	Aliases: []string{"pullrequest"},
	Short:   "Manage pull requests",
	Long:    "Create, list, show, approve, and reject Azure DevOps pull requests.",
}

// --- ado pr list ---

var prListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pull requests",
	Long:  "List pull requests matching the given filters.",
	RunE:  runPRList,
}

func runPRList(cmd *cobra.Command, args []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}
	project, err := resolveProject(cmd)
	if err != nil {
		return err
	}

	status, _ := cmd.Flags().GetString("status")
	creator, _ := cmd.Flags().GetString("creator")
	reviewer, _ := cmd.Flags().GetString("reviewer")
	repo, _ := cmd.Flags().GetString("repo")
	top, _ := cmd.Flags().GetInt("top")

	var repoID string
	if repo != "" {
		repoID, err = resolveRepoID(client, project, repo)
		if err != nil {
			return err
		}
	}

	prs, err := client.ListPullRequests(project, repoID, api.PullRequestQuery{
		Status:   status,
		Creator:  creator,
		Reviewer: reviewer,
		Top:      top,
	})
	if err != nil {
		return fmt.Errorf("listing pull requests: %w", err)
	}

	if len(prs) == 0 {
		if OutputFormat() == "json" {
			fmt.Println("[]")
		} else {
			fmt.Fprintln(os.Stderr, "No pull requests found.")
		}
		return nil
	}

	switch OutputFormat() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(prs)
	case "plain":
		for _, pr := range prs {
			fmt.Printf("%d\t%s\n", pr.ID, pr.Title)
		}
	default: // table
		fmt.Fprintf(os.Stdout, "%-8s %-50s %-20s %-20s %-12s %-20s\n",
			"ID", "Title", "Source", "Target", "Status", "Creator")
		fmt.Fprintln(os.Stdout, strings.Repeat("-", 134))
		for _, pr := range prs {
			fmt.Fprintf(os.Stdout, "%-8d %-50s %-20s %-20s %-12s %-20s\n",
				pr.ID,
				truncate(pr.Title, 50),
				truncate(shortBranch(pr.SourceBranch), 20),
				truncate(shortBranch(pr.TargetBranch), 20),
				pr.Status,
				truncate(pr.CreatedBy.DisplayName, 20),
			)
		}
	}
	return nil
}

// --- ado pr show ---

var prShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a pull request",
	Long:  "Show details of a single pull request.",
	Args:  cobra.ExactArgs(1),
	RunE:  runPRShow,
}

func runPRShow(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid pull request ID: %s", args[0])
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}
	project, err := resolveProject(cmd)
	if err != nil {
		return err
	}

	pr, err := client.GetPullRequest(project, id)
	if err != nil {
		return fmt.Errorf("fetching pull request %d: %w", id, err)
	}

	switch OutputFormat() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(pr)
	case "plain":
		fmt.Printf("%d\t%s\n", pr.ID, pr.Title)
	default: // table
		fmt.Printf("ID:           %d\n", pr.ID)
		fmt.Printf("Title:        %s\n", pr.Title)
		fmt.Printf("Status:       %s\n", pr.Status)
		fmt.Printf("Draft:        %v\n", pr.IsDraft)
		fmt.Printf("Source:       %s\n", shortBranch(pr.SourceBranch))
		fmt.Printf("Target:       %s\n", shortBranch(pr.TargetBranch))
		fmt.Printf("Creator:      %s\n", pr.CreatedBy.DisplayName)
		fmt.Printf("Merge Status: %s\n", pr.MergeStatus)
		fmt.Printf("Repository:   %s\n", pr.Repository.Name)
		if len(pr.Reviewers) > 0 {
			fmt.Println("\nReviewers:")
			for _, r := range pr.Reviewers {
				fmt.Printf("  - %s (%s)\n", r.DisplayName, voteString(r.Vote))
			}
		}
		if pr.Description != "" {
			fmt.Printf("\nDescription:\n%s\n", pr.Description)
		}
	}
	return nil
}

// --- ado pr create ---

var prCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a pull request",
	Long:  "Create a new pull request in Azure DevOps.",
	RunE:  runPRCreate,
}

func runPRCreate(cmd *cobra.Command, args []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}
	project, err := resolveProject(cmd)
	if err != nil {
		return err
	}

	repo, _ := cmd.Flags().GetString("repo")
	title, _ := cmd.Flags().GetString("title")
	source, _ := cmd.Flags().GetString("source")
	target, _ := cmd.Flags().GetString("target")
	desc, _ := cmd.Flags().GetString("description")
	reviewersStr, _ := cmd.Flags().GetString("reviewers")
	draft, _ := cmd.Flags().GetBool("draft")

	if repo == "" {
		return fmt.Errorf("--repo is required")
	}
	if title == "" {
		return fmt.Errorf("--title is required")
	}
	if source == "" {
		return fmt.Errorf("--source is required")
	}
	if target == "" {
		return fmt.Errorf("--target is required")
	}

	repoID, err := resolveRepoID(client, project, repo)
	if err != nil {
		return err
	}

	input := api.CreatePRInput{
		SourceRefName: ensureRef(source),
		TargetRefName: ensureRef(target),
		Title:         title,
		Description:   desc,
		IsDraft:       draft,
	}

	if reviewersStr != "" {
		for _, r := range strings.Split(reviewersStr, ",") {
			r = strings.TrimSpace(r)
			if r != "" {
				input.Reviewers = append(input.Reviewers, api.IdentityRef{ID: r})
			}
		}
	}

	pr, err := client.CreatePullRequest(project, repoID, input)
	if err != nil {
		return fmt.Errorf("creating pull request: %w", err)
	}

	switch OutputFormat() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(pr)
	case "plain":
		fmt.Printf("%d\t%s\n", pr.ID, pr.Title)
	default:
		fmt.Printf("Created pull request %d: %s\n", pr.ID, pr.Title)
	}
	return nil
}

// --- ado pr approve ---

var prApproveCmd = &cobra.Command{
	Use:   "approve <id>",
	Short: "Approve a pull request",
	Long:  "Approve a pull request (vote = 10).",
	Args:  cobra.ExactArgs(1),
	RunE:  runPRApprove,
}

func runPRApprove(cmd *cobra.Command, args []string) error {
	return votePR(cmd, args, 10, "Approved")
}

// --- ado pr reject ---

var prRejectCmd = &cobra.Command{
	Use:   "reject <id>",
	Short: "Reject a pull request",
	Long:  "Reject a pull request (vote = -10).",
	Args:  cobra.ExactArgs(1),
	RunE:  runPRReject,
}

func runPRReject(cmd *cobra.Command, args []string) error {
	return votePR(cmd, args, -10, "Rejected")
}

func votePR(cmd *cobra.Command, args []string, vote int, label string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid pull request ID: %s", args[0])
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}
	project, err := resolveProject(cmd)
	if err != nil {
		return err
	}

	// Get the PR to find the repository ID.
	pr, err := client.GetPullRequest(project, id)
	if err != nil {
		return fmt.Errorf("fetching pull request %d: %w", id, err)
	}

	// Get the authenticated user's ID.
	conn, err := client.GetConnectionData()
	if err != nil {
		return fmt.Errorf("getting authenticated user: %w", err)
	}

	if err := client.VotePullRequest(project, pr.Repository.ID, id, conn.AuthenticatedUser.ID, vote); err != nil {
		return fmt.Errorf("voting on pull request %d: %w", id, err)
	}

	switch OutputFormat() {
	case "json":
		out := map[string]interface{}{
			"pullRequestId": id,
			"vote":          vote,
			"status":        label,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	case "plain":
		fmt.Printf("%d\t%s\n", id, label)
	default:
		fmt.Printf("%s pull request %d\n", label, id)
	}
	return nil
}

// --- helpers ---

func resolveRepoID(client *api.Client, project, repoName string) (string, error) {
	repos, err := client.ListRepositories(project)
	if err != nil {
		return "", fmt.Errorf("listing repositories: %w", err)
	}
	for _, r := range repos {
		if strings.EqualFold(r.Name, repoName) {
			return r.ID, nil
		}
	}
	return "", fmt.Errorf("repository %q not found in project %q", repoName, project)
}

func shortBranch(ref string) string {
	return strings.TrimPrefix(ref, "refs/heads/")
}

func ensureRef(branch string) string {
	if strings.HasPrefix(branch, "refs/") {
		return branch
	}
	return "refs/heads/" + branch
}

func voteString(vote int) string {
	switch vote {
	case 10:
		return "Approved"
	case 5:
		return "Approved with suggestions"
	case 0:
		return "No vote"
	case -5:
		return "Waiting for author"
	case -10:
		return "Rejected"
	default:
		return fmt.Sprintf("%d", vote)
	}
}

func init() {
	// List flags
	prListCmd.Flags().StringP("project", "p", "", "Project name")
	prListCmd.Flags().String("status", "", "Filter by status (active, completed, abandoned, all)")
	prListCmd.Flags().String("creator", "", "Filter by creator ID")
	prListCmd.Flags().String("reviewer", "", "Filter by reviewer ID")
	prListCmd.Flags().String("repo", "", "Repository name")
	prListCmd.Flags().Int("top", 20, "Maximum number of results")

	// Show flags
	prShowCmd.Flags().StringP("project", "p", "", "Project name")

	// Create flags
	prCreateCmd.Flags().StringP("project", "p", "", "Project name")
	prCreateCmd.Flags().String("repo", "", "Repository name (required)")
	prCreateCmd.Flags().String("title", "", "Pull request title (required)")
	prCreateCmd.Flags().String("source", "", "Source branch (required)")
	prCreateCmd.Flags().String("target", "", "Target branch (required)")
	prCreateCmd.Flags().String("description", "", "Pull request description")
	prCreateCmd.Flags().String("reviewers", "", "Comma-separated reviewer IDs")
	prCreateCmd.Flags().Bool("draft", false, "Create as draft pull request")

	// Approve flags
	prApproveCmd.Flags().StringP("project", "p", "", "Project name")

	// Reject flags
	prRejectCmd.Flags().StringP("project", "p", "", "Project name")

	prCmd.AddCommand(prListCmd)
	prCmd.AddCommand(prShowCmd)
	prCmd.AddCommand(prCreateCmd)
	prCmd.AddCommand(prApproveCmd)
	prCmd.AddCommand(prRejectCmd)

	rootCmd.AddCommand(prCmd)
}
