/*
Copyright © 2023 Wade Bourne wade@wbourne.dev
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// openTemplateCmd represents the open-template command
var openTemplateCmd = &cobra.Command{
	Use:   "open-template",
	Aliases: []string{"ot", "opentemplate"},
	Short: "Open a template for editing",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		items := make([]string, 0, len(templates))

		for n, v := range templates {
			if v == "" {
				continue
			}

			items = append(items, n)
		}

		return items, cobra.ShellCompDirectiveNoFileComp
	},
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if name == "" {
			fmt.Println("expected a non-empty template name")
			os.Exit(1)
		}


		ts, isFolderSource := resolveTemplate(name).(folderSource)

		if !isFolderSource {
			fmt.Println("cannot open builtin templates")
			os.Exit(1)
		}

		var tmplFound bool

		for projname := range templates {
			fmt.Println(projname, name)
			if projname == name {
				tmplFound = true
				break
			}
		}

		if !tmplFound {
			fmt.Println("template not found:", name)
			os.Exit(1)
		}

		openProject(string(ts))
	},
}

func init() {
	rootCmd.AddCommand(openTemplateCmd)
}
