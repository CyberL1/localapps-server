package cmd

import (
	"fmt"
	"localapps-server/constants"
	"localapps-server/utils"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Check the server version",
	Run:   version,
}

func version(cmd *cobra.Command, args []string) {
	latestRelease, err := utils.GetLatestCliVersion()
	if err != nil {
		fmt.Println("Failed to get latest release", err)
		return
	}

	currentVersion, err := semver.NewVersion(constants.Version)
	if err != nil {
		fmt.Println("Failed to parse current version", err)
		return
	}

	newVersion, err := semver.NewVersion(latestRelease.TagName)
	if err != nil {
		fmt.Println("Failed to parse latest version", err)
		return
	}

	if currentVersion.LessThan(newVersion) {
		fmt.Println("A new update is available\nRun 'localapps-server upgrade' to upgrade")
	}

	fmt.Printf("Your server Version: %s\nLatest server version: %s\n", constants.Version, latestRelease.TagName)
}
