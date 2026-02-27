package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gyurisc/adocli/internal/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newAPIClient creates a new API client using config and keyring.
func newAPIClient() (*api.Client, error) {
	org := viper.GetString("organization")
	if org == "" {
		return nil, fmt.Errorf("organization not configured (run 'ado config set organization <org>')")
	}
	pat, err := GetPAT()
	if err != nil {
		return nil, err
	}
	return api.NewClient(org, pat), nil
}

// resolveProject returns the project from the flag or config default.
func resolveProject(cmd *cobra.Command) (string, error) {
	p, _ := cmd.Flags().GetString("project")
	if p != "" {
		return p, nil
	}
	p = viper.GetString("project")
	if p == "" {
		return "", fmt.Errorf("project not specified (use --project or 'ado config set project <name>')")
	}
	return p, nil
}

var workitemCmd = &cobra.Command{
	Use:     "workitem",
	Aliases: []string{"wi"},
	Short:   "Manage work items",
	Long:    "Create, list, show, and update Azure DevOps work items.",
}

// --- ado workitem list ---

var wiListCmd = &cobra.Command{
	Use:   "list",
	Short: "List work items",
	Long:  "List work items matching the given filters using WIQL.",
	RunE:  runWorkitemList,
}

// escapeWIQL escapes single quotes in a string for safe WIQL interpolation.
func escapeWIQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func buildWIQL(project, wiType, state, assignedTo string) string {
	q := "SELECT [System.Id], [System.Title], [System.State], [System.WorkItemType], [System.AssignedTo] FROM WorkItems"

	var conditions []string
	conditions = append(conditions, fmt.Sprintf("[System.TeamProject] = '%s'", escapeWIQL(project)))

	if wiType != "" {
		conditions = append(conditions, fmt.Sprintf("[System.WorkItemType] = '%s'", escapeWIQL(wiType)))
	}
	if state != "" {
		conditions = append(conditions, fmt.Sprintf("[System.State] = '%s'", escapeWIQL(state)))
	}
	if assignedTo != "" {
		if assignedTo == "@me" {
			conditions = append(conditions, "[System.AssignedTo] = @me")
		} else {
			conditions = append(conditions, fmt.Sprintf("[System.AssignedTo] = '%s'", escapeWIQL(assignedTo)))
		}
	}

	q += " WHERE " + strings.Join(conditions, " AND ")
	q += " ORDER BY [System.ChangedDate] DESC"

	return q
}

func runWorkitemList(cmd *cobra.Command, args []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}
	project, err := resolveProject(cmd)
	if err != nil {
		return err
	}

	wiType, _ := cmd.Flags().GetString("type")
	state, _ := cmd.Flags().GetString("state")
	assignedTo, _ := cmd.Flags().GetString("assigned-to")
	top, _ := cmd.Flags().GetInt("top")

	wiql := buildWIQL(project, wiType, state, assignedTo)

	result, err := client.QueryByWiql(project, wiql)
	if err != nil {
		return fmt.Errorf("querying work items: %w", err)
	}

	if len(result.WorkItems) == 0 {
		if OutputFormat() == "json" {
			fmt.Println("[]")
		} else {
			fmt.Fprintln(os.Stderr, "No work items found.")
		}
		return nil
	}

	// Collect IDs, respecting --top.
	ids := make([]int, 0, len(result.WorkItems))
	for _, ref := range result.WorkItems {
		ids = append(ids, ref.ID)
	}
	if top > 0 && len(ids) > top {
		ids = ids[:top]
	}

	items, err := client.GetWorkItems(project, ids)
	if err != nil {
		return fmt.Errorf("fetching work items: %w", err)
	}

	switch OutputFormat() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(items)
	case "plain":
		for _, wi := range items {
			title, _ := wi.Fields["System.Title"].(string)
			fmt.Printf("%d\t%s\n", wi.ID, title)
		}
	default: // table
		fmt.Fprintf(os.Stdout, "%-8s %-16s %-50s %-12s %-20s\n", "ID", "Type", "Title", "State", "Assigned To")
		fmt.Fprintln(os.Stdout, strings.Repeat("-", 110))
		for _, wi := range items {
			title := truncate(fieldStr(wi.Fields, "System.Title"), 50)
			wiT := fieldStr(wi.Fields, "System.WorkItemType")
			st := fieldStr(wi.Fields, "System.State")
			assigned := truncate(fieldStr(wi.Fields, "System.AssignedTo"), 20)
			fmt.Fprintf(os.Stdout, "%-8d %-16s %-50s %-12s %-20s\n", wi.ID, wiT, title, st, assigned)
		}
	}
	return nil
}

// --- ado workitem show ---

var wiShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a work item",
	Long:  "Show details of a single work item.",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkitemShow,
}

func runWorkitemShow(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid work item ID: %s", args[0])
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	project, err := resolveProject(cmd)
	if err != nil {
		return err
	}

	wi, err := client.GetWorkItem(project, id)
	if err != nil {
		return fmt.Errorf("fetching work item %d: %w", id, err)
	}

	switch OutputFormat() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(wi)
	case "plain":
		title := fieldStr(wi.Fields, "System.Title")
		fmt.Printf("%d\t%s\n", wi.ID, title)
	default: // table
		fmt.Printf("ID:           %d\n", wi.ID)
		fmt.Printf("Type:         %s\n", fieldStr(wi.Fields, "System.WorkItemType"))
		fmt.Printf("Title:        %s\n", fieldStr(wi.Fields, "System.Title"))
		fmt.Printf("State:        %s\n", fieldStr(wi.Fields, "System.State"))
		fmt.Printf("Assigned To:  %s\n", fieldStr(wi.Fields, "System.AssignedTo"))
		fmt.Printf("Area Path:    %s\n", fieldStr(wi.Fields, "System.AreaPath"))
		fmt.Printf("Iteration:    %s\n", fieldStr(wi.Fields, "System.IterationPath"))
		desc := fieldStr(wi.Fields, "System.Description")
		if desc != "" {
			fmt.Printf("\nDescription:\n%s\n", desc)
		}
	}
	return nil
}

// --- ado workitem create ---

var wiCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a work item",
	Long:  "Create a new work item in Azure DevOps.",
	RunE:  runWorkitemCreate,
}

func runWorkitemCreate(cmd *cobra.Command, args []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}
	project, err := resolveProject(cmd)
	if err != nil {
		return err
	}

	wiType, _ := cmd.Flags().GetString("type")
	title, _ := cmd.Flags().GetString("title")
	desc, _ := cmd.Flags().GetString("description")
	assignedTo, _ := cmd.Flags().GetString("assigned-to")
	areaPath, _ := cmd.Flags().GetString("area-path")
	iterationPath, _ := cmd.Flags().GetString("iteration-path")

	if wiType == "" {
		return fmt.Errorf("--type is required")
	}
	if title == "" {
		return fmt.Errorf("--title is required")
	}

	fields := []api.PatchField{
		{Op: "add", Path: "/fields/System.Title", Value: title},
	}
	if desc != "" {
		fields = append(fields, api.PatchField{Op: "add", Path: "/fields/System.Description", Value: desc})
	}
	if assignedTo != "" {
		fields = append(fields, api.PatchField{Op: "add", Path: "/fields/System.AssignedTo", Value: assignedTo})
	}
	if areaPath != "" {
		fields = append(fields, api.PatchField{Op: "add", Path: "/fields/System.AreaPath", Value: areaPath})
	}
	if iterationPath != "" {
		fields = append(fields, api.PatchField{Op: "add", Path: "/fields/System.IterationPath", Value: iterationPath})
	}

	wi, err := client.CreateWorkItem(project, wiType, fields)
	if err != nil {
		return fmt.Errorf("creating work item: %w", err)
	}

	switch OutputFormat() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(wi)
	case "plain":
		fmt.Printf("%d\t%s\n", wi.ID, fieldStr(wi.Fields, "System.Title"))
	default:
		fmt.Printf("Created work item %d: %s\n", wi.ID, fieldStr(wi.Fields, "System.Title"))
	}
	return nil
}

// --- ado workitem update ---

var wiUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a work item",
	Long:  "Update fields on an existing work item.",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkitemUpdate,
}

func runWorkitemUpdate(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid work item ID: %s", args[0])
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	project, err := resolveProject(cmd)
	if err != nil {
		return err
	}

	title, _ := cmd.Flags().GetString("title")
	state, _ := cmd.Flags().GetString("state")
	assignedTo, _ := cmd.Flags().GetString("assigned-to")

	var fields []api.PatchField
	if title != "" {
		fields = append(fields, api.PatchField{Op: "replace", Path: "/fields/System.Title", Value: title})
	}
	if state != "" {
		fields = append(fields, api.PatchField{Op: "replace", Path: "/fields/System.State", Value: state})
	}
	if assignedTo != "" {
		fields = append(fields, api.PatchField{Op: "replace", Path: "/fields/System.AssignedTo", Value: assignedTo})
	}

	if len(fields) == 0 {
		return fmt.Errorf("no fields to update (use --title, --state, or --assigned-to)")
	}

	wi, err := client.UpdateWorkItem(project, id, fields)
	if err != nil {
		return fmt.Errorf("updating work item %d: %w", id, err)
	}

	switch OutputFormat() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(wi)
	case "plain":
		fmt.Printf("%d\t%s\n", wi.ID, fieldStr(wi.Fields, "System.Title"))
	default:
		fmt.Printf("Updated work item %d: %s\n", wi.ID, fieldStr(wi.Fields, "System.Title"))
	}
	return nil
}

// --- helpers ---

func fieldStr(fields map[string]interface{}, key string) string {
	v, ok := fields[key]
	if !ok {
		return ""
	}
	// AssignedTo may come as a nested object with displayName.
	if m, ok := v.(map[string]interface{}); ok {
		if name, ok := m["displayName"].(string); ok {
			return name
		}
	}
	return fmt.Sprintf("%v", v)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func init() {
	// List flags
	wiListCmd.Flags().StringP("project", "p", "", "Project name")
	wiListCmd.Flags().String("type", "", "Work item type (Bug, Task, User Story, etc.)")
	wiListCmd.Flags().String("state", "", "Filter by state (New, Active, Closed, etc.)")
	wiListCmd.Flags().String("assigned-to", "", "Filter by assigned user (@me for current user)")
	wiListCmd.Flags().Int("top", 20, "Maximum number of results")

	// Show flags
	wiShowCmd.Flags().StringP("project", "p", "", "Project name")

	// Create flags
	wiCreateCmd.Flags().StringP("project", "p", "", "Project name")
	wiCreateCmd.Flags().String("type", "", "Work item type (required)")
	wiCreateCmd.Flags().String("title", "", "Title (required)")
	wiCreateCmd.Flags().String("description", "", "Description")
	wiCreateCmd.Flags().String("assigned-to", "", "Assigned to user")
	wiCreateCmd.Flags().String("area-path", "", "Area path")
	wiCreateCmd.Flags().String("iteration-path", "", "Iteration path")

	// Update flags
	wiUpdateCmd.Flags().StringP("project", "p", "", "Project name")
	wiUpdateCmd.Flags().String("title", "", "New title")
	wiUpdateCmd.Flags().String("state", "", "New state")
	wiUpdateCmd.Flags().String("assigned-to", "", "New assigned user")

	workitemCmd.AddCommand(wiListCmd)
	workitemCmd.AddCommand(wiShowCmd)
	workitemCmd.AddCommand(wiCreateCmd)
	workitemCmd.AddCommand(wiUpdateCmd)

	rootCmd.AddCommand(workitemCmd)
}
