/*
Copyright © 2023 Wade Bourne wade@wbourne.dev
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// createTemplateCmd represents the create-template command
var createTemplateCmd = &cobra.Command{
	Use:     "create-template",
	Aliases: []string{"ct", "createtemplate"},
	Short:   "Create a new template",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if name == "" {
			fmt.Println("expected template name to be a non-empty string")
			os.Exit(1)
		}

		if templatePath == "" {
			fmt.Println("expected template directory to be a non-empty string")
			os.Exit(1)
		}

		templateDir := templateDirs[len(templateDirs)-1]

		for _, tmpl := range templates {
			if tmpl == name {
				fmt.Println("template with that name already exists")
				os.Exit(1)
			}
		}

		ts := resolveTemplateFromArg(cmd)

		newDir := filepath.Join(templateDir, name)

		ts.copyTo(newDir)

		openProject(newDir)
	},
}

func init() {
	rootCmd.AddCommand(createTemplateCmd)
	createTemplateCmd.Flags().StringP("template", "t", "", "base template for the new template")
	createTemplateCmd.RegisterFlagCompletionFunc("template", templateCompletion)
}
