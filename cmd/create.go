/*
Copyright © 2023 Wade Bourne wade@wbourne.dev
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Aliases: []string{"c", "new"},
	Short: "Create a new project",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(c *cobra.Command, args []string) {
		name := args[0]

		if name == "" {
			fmt.Println("expected non-empty name")
			os.Exit(1)
		}

		if strings.Contains(name, " ") {
			fmt.Println("expected name to not contain spaces")
			os.Exit(1)
		}

		ts := resolveTemplateFromArg(c)
		dirName := filepath.Join(projDir, name)

		ts.copyTo(dirName)
		initProject(dirName, name)

		openProject(dirName)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringP("template", "t", "", "template the project should be based on")
	createCmd.RegisterFlagCompletionFunc("template", templateCompletion)
}
