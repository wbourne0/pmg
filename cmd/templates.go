/*
Copyright © 2023 Wade Bourne wade@wbourne.dev
*/
package cmd

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/hashicorp/go-envparse"
	"github.com/spf13/cobra"
)

//go:embed all:templates
var builtinTemplates embed.FS

var templates = map[string]string{}
var templateDirs []string
var projects []string
var templatePath string
var projDir string

const sublProjectFile = `{
	"folders":
	[
		{
			"path": "."
		}
	]
}
`

func templateCompletion(cmd *cobra.Command, args []string, toComplete string) (items []string, d cobra.ShellCompDirective) {
	items = make([]string, 0, len(templates))

	for n := range templates {
		items = append(items, n)
	}

	return items, cobra.ShellCompDirectiveNoFileComp
}

func listDirs(f fs.FS) (dirs []string) {
	dir, err := fs.ReadDir(f, ".")

	if err != nil {
		fmt.Println("unable to read template directory:", err.Error())
		os.Exit(1)
	}

	dirs = make([]string, 0, len(dir))

	for _, ent := range dir {
		if ent.Type().IsDir() {
			dirs = append(dirs, ent.Name())
		}
	}

	return
}

func init() {
	templatePath = os.Getenv("PROJMGR_TEMPLATE_PATH")

	projDir = os.Getenv("PROJMGR_PROJDIR")

	if projDir == "" {
		fmt.Println("expected PROJMGR_PROJDIR to be set")
		os.Exit(1)
	}

	info, err := os.Stat(projDir)

	if err != nil {
		err := os.MkdirAll(projDir, 0770)

		if err != nil {
			fmt.Println("failed to create project directory:", err.Error())
			os.Exit(1)
		}
	} else if !info.IsDir() {
		fmt.Println("expected PROJMGR_PROJDIR to be a directory")
		os.Exit(1)
	}

	projects = listDirs(os.DirFS(projDir))

	tmpdir, _ := fs.Sub(builtinTemplates, "templates")

	for _, tmpl := range listDirs(tmpdir) {
		templates[tmpl] = ""
	}

	tmplPaths := strings.Split(templatePath, ":")
	templateDirs = make([]string, 0, len(tmplPaths))

	for _, path := range tmplPaths {
		if path == "" {
			continue
		}

		templateDirs = append(templateDirs, path)

		for _, tmpl := range listDirs(os.DirFS(path)) {
			templates[tmpl] = filepath.Join(path, tmpl)
		}
	}
}

func (f folderSource) copyFile(from, to, toDir string, wg *sync.WaitGroup) {
	defer wg.Done()
	info, err := os.Lstat(from)

	if err != nil {
		fmt.Printf("failed to copy %s: %s\n", from, err.Error())
		return
	}

	if info.Mode().Type() == os.ModeSymlink {
		target, err := os.Readlink(from)
		if err != nil {
			fmt.Printf("failed to copy %s: %s\n", from, err.Error())
			return
		}

		if filepath.IsAbs(target) && strings.HasPrefix(target, string(f)+"/") {
			rel, _ := filepath.Rel(string(f), target)
			to = filepath.Join(toDir, rel)
		}

		if err := os.Symlink(target, to); err != nil {
			fmt.Printf("failed to copy %s: %s\n", from, err.Error())

			return
		}

		return
	}

	if !info.Mode().IsRegular() {
		fmt.Printf("unexpected irregular file: %s\n", from)
		return
	}

	src, err := os.Open(from)

	if err != nil {
		fmt.Printf("failed to copy %s: %s\n", from, err.Error())
		return
	}

	defer src.Close()

	dst, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY, info.Mode().Perm())

	if err != nil {
		fmt.Printf("failed to copy %s: %s\n", from, err.Error())
		return
	}

	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		fmt.Printf("failed to copy %s: %s\n", from, err.Error())
		return
	}

	return
}

func (f folderSource) copyTo(to string) {

	if err := os.Mkdir(to, 0770); err != nil {
		fmt.Println("Unable to create project directory:", err.Error())
		os.Exit(1)
	}

	var wg sync.WaitGroup

	filepath.WalkDir(string(f), func(path string, d fs.DirEntry, err error) error {
		if path == string(f) {
			return nil
		}

		relPath, _ := filepath.Rel(string(f), path)

		if d.IsDir() {
			err := os.Mkdir(filepath.Join(to, relPath), 0770)

			if err != nil {
				fmt.Println("failed to mkdir")
			}

			return err
		}

		if relPath == "project.sublime-workspace" {
			return nil
		}

		wg.Add(1)

		go f.copyFile(path, filepath.Join(to, relPath), to, &wg)

		return nil

	})

	wg.Wait()
}

type templateSource interface {
	copyTo(path string)
}

type folderSource string

type embedSource struct {
	fs.FS
}

func (es embedSource) copyTo(to string) {
	if err := os.Mkdir(to, 0770); err != nil {
		fmt.Println("Unable to create project directory:", err.Error())
		os.Exit(1)
	}

	fs.WalkDir(es, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." || path == "project.sublime-workspace" {
			return nil
		}

		newPath := filepath.Join(to, path)

		if d.IsDir() {
			err := os.Mkdir(newPath, 0770)

			if err != nil {
				fmt.Printf("failed to make directory %s: %s\n", path, err.Error())
				return fs.SkipDir
			}

			return nil
		}

		file, err := os.Create(newPath)

		if err != nil {
			fmt.Printf("failed to create file %s: %s\n", path, err)
			return nil
		}

		defer file.Close()

		src, _ := es.Open(path)
		defer src.Close()

		if _, err = io.Copy(file, src); err != nil {
			fmt.Printf("failed to write file %s: %s\n", path, err)
			return nil
		}

		return nil
	})
}

func resolveTemplateFromArg(cmd *cobra.Command) templateSource {
	templateName, _ := cmd.Flags().GetString("template")

	return resolveTemplate(templateName)
}

func resolveTemplate(templateName string) templateSource {
	if templateName == "" {
		templateName = "default"
	}

	path, ok := templates[templateName]

	if !ok {
		fmt.Println("template not found:", templateName)
		os.Exit(1)
	}

	if path == "" {
		f, err := fs.Sub(builtinTemplates, filepath.Join("templates", templateName))

		if err != nil {
			fmt.Println("failed to load builtin template:", err.Error())
			os.Exit(1)
		}

		return embedSource{f}
	}

	return folderSource(path)
}

func initProject(projectDir, name string) {
	autorunPath := filepath.Join(projectDir, ".pmg/setup")

	if _, err := os.Stat(autorunPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return
		}

		fmt.Println("unable to read autoinit file:", err.Error())
		os.Exit(1)
	}

	cmd := exec.Command(autorunPath, name)

	cmd.Dir = projectDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Println("error occurred while running setup:", err.Error())
	}

	if err := os.Remove(autorunPath); err != nil {
		fmt.Println("failed to remove setup file:", err.Error())
		os.Exit(1)
	}
}

func openProject(dirName string) {
	var err error

	isHeadless, _ := rootCmd.PersistentFlags().GetBool("no-editor")

	if !isHeadless {
		editor := os.Getenv("EDITOR")

		if editor == "" {
			fmt.Println("missing EDITOR env var; unable to open editor")
		} else if editor == "subl" {
			projPath := filepath.Join(dirName, "project.sublime-project")
			_, err := os.Stat(projPath)

			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					fmt.Println("unable to stat project.sublime-project file")
					os.Exit(1)
				}

				if err := ioutil.WriteFile(projPath, []byte(sublProjectFile), 0660); err != nil {
					fmt.Println("unable to write project.sublime-project file:", err.Error())
					os.Exit(1)
				}
			}

			err = exec.Command("subl", "-p", projPath).Run()
			if err != nil {
				fmt.Println("unable to start subl:", err.Error())
				os.Exit(1)
			}
		} else if editor == "code" || editor == "code-oss" {
			err = exec.Command(editor, dirName).Run()
			if err != nil {
				fmt.Println("unable to start vscode:", err.Error())
				os.Exit(1)
			}
		} else {
			// this doesn't support vim as vim would have to share the tty with the shell
			// which doesn't make sense
			fmt.Println("Unsupported editor, please make a support ticket (or update EDITOR env var)")
		}
	}

	shellPath := os.Getenv("SHELL")

	envvars, err := os.Open(filepath.Join(dirName, ".pmg/env"))

	if err == nil {
		defer envvars.Close()
		parsed, err := envparse.Parse(envvars)

		if err != nil {
			os.Exit(1)
		}

		for name, val := range parsed {
			val = os.Expand(val, func(s string) string {
				if s == "projdir" {
					return dirName
				}

				return os.Getenv(s)
			})

			os.Setenv(name, val)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		fmt.Println("failed to read envvar file:", err.Error())
		os.Exit(1)
	}

	if shellPath == "" {
		shellPath, err = exec.LookPath("sh")
		if err != nil {
			fmt.Println("unable to resolve shell:", err.Error())
			os.Exit(1)
		}
	}

	os.Chdir(dirName)
	if err = syscall.Exec(shellPath, []string{shellPath}, os.Environ()); err != nil {
		fmt.Println("unable to exec shell:", err.Error())
		os.Exit(1)
	}
}
