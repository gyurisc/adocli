package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

const (
	keyringService = "adocli"
	keyringUser    = "pat"
)

// GetPAT retrieves the stored PAT from the OS keyring.
func GetPAT() (string, error) {
	pat, err := keyring.Get(keyringService, keyringUser)
	if err != nil {
		return "", fmt.Errorf("no PAT found in keyring (run 'ado auth login'): %w", err)
	}
	return pat, nil
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  "Manage Azure DevOps authentication credentials.",
}

// --- ado auth login ---

var patFlag string

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store a Personal Access Token (PAT)",
	Long: `Authenticate with Azure DevOps by storing a PAT in the OS keyring.

Provide the token via --pat flag or enter it interactively:
  ado auth login --pat <token>
  ado auth login`,
	RunE: runAuthLogin,
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	pat := patFlag
	if pat == "" {
		fmt.Fprint(os.Stderr, "Enter PAT: ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading PAT: %w", err)
		}
		pat = strings.TrimSpace(line)
	}
	if pat == "" {
		return fmt.Errorf("PAT cannot be empty")
	}

	if err := keyring.Set(keyringService, keyringUser, pat); err != nil {
		return fmt.Errorf("storing PAT in keyring: %w", err)
	}

	fmt.Fprintln(os.Stderr, "PAT stored successfully.")
	return nil
}

// --- ado auth logout ---

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	Long:  "Delete the stored PAT from the OS keyring.",
	RunE:  runAuthLogout,
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	if err := keyring.Delete(keyringService, keyringUser); err != nil {
		return fmt.Errorf("removing PAT from keyring: %w", err)
	}
	fmt.Fprintln(os.Stderr, "PAT removed from keyring.")
	return nil
}

// --- ado auth status ---

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long:  "Check whether a PAT is stored in the OS keyring.",
	RunE:  runAuthStatus,
}

type authStatusOutput struct {
	Authenticated bool   `json:"authenticated"`
	TokenStored   bool   `json:"token_stored"`
	TokenPrefix   string `json:"token_prefix,omitempty"`
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	pat, err := keyring.Get(keyringService, keyringUser)
	authenticated := err == nil && pat != ""

	status := authStatusOutput{
		Authenticated: authenticated,
		TokenStored:   authenticated,
	}
	if authenticated && len(pat) >= 4 {
		status.TokenPrefix = pat[:4] + "..."
	}

	switch OutputFormat() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(status)
	default:
		if authenticated {
			fmt.Printf("Authenticated: yes (token: %s)\n", status.TokenPrefix)
		} else {
			fmt.Println("Authenticated: no")
			fmt.Println("Run 'ado auth login' to authenticate.")
		}
		return nil
	}
}

func init() {
	authLoginCmd.Flags().StringVar(&patFlag, "pat", "", "Personal Access Token (non-interactive)")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)

	rootCmd.AddCommand(authCmd)
}
