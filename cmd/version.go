package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

type versionOutput struct {
	Version  string `json:"version"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	GoVer    string `json:"go_version"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		info := versionOutput{
			Version:  appVersion,
			OS:       runtime.GOOS,
			Arch:     runtime.GOARCH,
			GoVer:    runtime.Version(),
		}

		switch OutputFormat() {
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(info)
		default:
			fmt.Printf("ado %s (%s/%s, %s)\n", info.Version, info.OS, info.Arch, info.GoVer)
			return nil
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
