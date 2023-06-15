/*
Copyright © 2023 Wade Bourne wade@wbourne.dev
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	giturl "github.com/kubescape/go-git-url"
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:     "import",
	Aliases: []string{"i"},
	Short:   "Import a project from a github repository",
	Args:    cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")

		url, err := giturl.NewGitURL(args[0])

		if err != nil {
			fmt.Println("invalid git url:", err.Error())
		}

		if name == "" {
			name = url.GetRepoName()
		}

		for _, p := range projects {
			if p == name {
				fmt.Println("a project with this name already exists; please choose a different name")
				os.Exit(1)
			}
		}

		gitCmd := exec.Command("git", "clone", args[0], name)

		gitCmd.Stderr = os.Stderr
		gitCmd.Stdin = os.Stdin
		gitCmd.Stdout = os.Stdout
		gitCmd.Dir = projDir

		if err := gitCmd.Run(); err != nil {
			fmt.Println("error occurred while running git command:", err.Error())
			os.Exit(1)
		}

		openProject(filepath.Join(projDir, name))
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringP("name", "n", "", "Local project name (defaults to repo name)")
}
