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

var openCmd = &cobra.Command{
	Use:     "open",
	Aliases: []string{"o"},
	Short:   "Open a project by name",

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projects, cobra.ShellCompDirectiveNoFileComp
	},
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		dirname := filepath.Join(projDir, name)

		var projFound bool

		for _, projname := range projects {
			if projname == name {
				projFound = true
				break
			}
		}

		if !projFound {
			fmt.Println("project not found:", name)
			os.Exit(1)
		}

		openProject(dirname)
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}
