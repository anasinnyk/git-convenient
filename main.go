package main

import (
	"os"
	"fmt"
	"regexp"
	"strings"
	"os/exec"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"github.com/spf13/cobra"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
)

var (
	version  string
	revision string
)

type OptionalPart struct {
	Skip bool
}

type Commit struct {
	ValidatePattern string `git:"validate-pattern"`
	Scope struct {
		OptionalPart
		Pattern string
		Replace string
	}
	Body OptionalPart
	Footer OptionalPart
}

type Config struct {
	Commit Commit `git:"convenient"`
	User struct {
		Email string
		Name  string
	}
}

type App struct {
	Config *Config
}

type CommitType struct {
	Value       string
	Description string
}

func (a *App) SelectCommitType() string	 {
	types := []CommitType{
		{Value: "feat",        Description: "A new feature"},
		{Value: "fix",         Description: "A bug fix"},
		{Value: "breaking",    Description: "A commit with contain breaking changes"},
		{Value: "chore",       Description: "Any other changes not ralated to another categories (empty commit for trigger CI for example)"},
		{Value: "ci",          Description: "Changes to our CI configuration files and scripts"},
		{Value: "docs",        Description: "Documentation only changes"},
		{Value: "build",       Description: "Changes that affect the build system or external dependencies"},
		{Value: "pref",        Description: "A code change that improves performance"},
		{Value: "refactor",    Description: "A code change that neither fixes a bug nor adds a feature"},
		{Value: "revert",      Description: "When you revert some commit"},
		{Value: "style",       Description: "Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)"},
		{Value: "test",        Description: "Adding missing tests or correcting existing tests"},
		{Value: "improvement", Description: "A improve a current implementation without adding a new feature or fixing a bug"},
	}

	templatesType := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   `{{ "\u27A1" | yellow }} {{ .Value | cyan }}: {{ .Description | yellow }}`,
		Inactive: "  {{ .Value | cyan }}: {{ .Description | yellow }}",
		Selected: `{{ "\u2714" | green }} {{ "Select Commit Type" | cyan | bold }} {{ .Value }}`,
	}

	promptType := promptui.Select{
		Label:     "Select Commit Type",
		Items:     types,
		Templates: templatesType,
		Size:      len(types),
	}

	i, _, err := promptType.Run()
	if err != nil {
		log.Panicf("Incorrect commit type\nDetails: %v\n", err)
	}
	return types[i].Value
}

func (a *App) DetectScope() []string {
	diff, err := exec.Command("git", "diff", "--staged", "--name-status").Output()
	if err != nil {
		log.Printf("Git diff %v\n", err)
		return nil
	}
	
	var scopes []string
	re, err := regexp.Compile(a.Config.Commit.Scope.Pattern)
	if err != nil {
		log.Error(err)
	}

	files := strings.Split(strings.Trim(string(diff), "\n"), "\n")
	for _, f := range files {
		f = strings.Trim(f, " ")
		if re.MatchString(f) {
			scopes = append(scopes, re.ReplaceAllString(f, a.Config.Commit.Scope.Replace))
		}
	}

	func (xs *[]string) {
		found := make(map[string]bool)
		j := 0
		for i, x := range *xs {
			if !found[x] {
				found[x] = true
				(*xs)[j] = (*xs)[i]
				j++
			}
		}
		*xs = (*xs)[:j]
	}(&scopes)

	return scopes
}

func (a *App) prompt(label string, def string) string {
	prompt := promptui.Prompt{
		Label:     label,
		Default:   def,
		Templates: &promptui.PromptTemplates{
			Prompt:  "{{ . }} ",
			Success: "{{ \"\u2714\" | green }} {{ . | cyan | bold }} ",
		},
	}

	msg, _ := prompt.Run()
	return msg
}

func (a *App) Commit() {
	typeName := a.SelectCommitType()
	var scopes string
	if !a.Config.Commit.Scope.Skip {
		var dScopes []string
		if a.Config.Commit.Scope.Pattern != "" && a.Config.Commit.Scope.Replace != "" {
			dScopes = a.DetectScope()
		}
		scopes = a.prompt("Enter your scope without breket or press enter for apply this", strings.Join(dScopes, ","))
	}
	msg := a.prompt("Enter your comment", "")
	
	var body string
	if !a.Config.Commit.Body.Skip {
		body = a.prompt("Enter your body part", "")
	}

	var footer string
	if !a.Config.Commit.Body.Skip {
		footer = a.prompt("Enter your footer part", "")
	}

	commit := fmt.Sprintf("%s: %s", typeName, msg)
	if scopes != "" {
		commit = fmt.Sprintf("%s(%s): %s", typeName, scopes, msg)
	}

	if body != "" {
		commit = fmt.Sprintf("%s\n\n%s", commit, body)
	}
	if footer != "" {
		commit = fmt.Sprintf("%s\n\n%s", commit, footer)
	}

	if a.validate(commit) {
		r, err := git.PlainOpen(".")
		if err != nil {
			log.Fatalf("Cannot repository open %v", err)
		}
		wt, err := r.Worktree()
		if err != nil {
			log.Fatalf("Cannot get worktree %v", err)
		}
		hash, err := wt.Commit(commit, &git.CommitOptions{
			Author: &object.Signature{
				Name: a.Config.User.Name,
				Email: a.Config.User.Email,
			},
		})
		if err != nil {
			log.Fatalf("Cannot commited %v", err)
		}
		fmt.Println(hash)
	} else {
		fmt.Println("Commit is not valid")
	}
}

func (a *App) validate(commit string) bool {
	if a.Config.Commit.ValidatePattern == "" {
		a.Config.Commit.ValidatePattern = "^(feat|fix|breaking|chore|ci|docs|build|pref|refactor|revert|style|test|improvement)(\\([a-zA-Z0-9_\\-,]+\\))?: ([^\n]+(\n[\\s\\S]+)$|.*$)"
	}
	validator := regexp.MustCompile(a.Config.Commit.ValidatePattern)
	return validator.MatchString(commit)
}

func (a *App) CheckIfCommitValid(hash string) bool {
	r, err := git.PlainOpen(".")
	if err != nil {
		log.Fatalf("Cannot repository open %v", err)
	}
	obj, err := r.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		log.Fatalf("Cannto find commit %s. %v", hash, err)
	}
	return a.validate(obj.Message)
}

func (a *App) CheckIfMsgValid(msg string) bool {
	return a.validate(msg)
}

func (a *App) InstallHook() {
	hook := "MSG=`cat .git/COMMIT_EDITMSG`\ngit convenient validate -m \"$MSG\""

	f, err := os.OpenFile(".git/hooks/commit-msg", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	if _, err = f.WriteString(hook); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var conf Config
	g := GitConfig{}
	g.ParseFile(&conf)

	app := &App{
		Config: &conf,
	}

	commit := &cobra.Command{
		Use:   "commit",
		Short: "Create git commit in convenient style",
		Run: func(cmd *cobra.Command, args []string) {
		    app.Commit()
		},
	}

	var hash string
	var msg string
	validate := &cobra.Command{
		Use:   "validate",
		Short: "Check commit message if it's valid",
		Run: func(cmd *cobra.Command, args []string) {
			var cValid = true
			var mValid = true
			if hash != "" {
				cValid = app.CheckIfCommitValid(hash)
			}

			if msg != "" {
				mValid = app.CheckIfMsgValid(msg)
			}

			if (cValid && mValid) {
				fmt.Println("It's valid")
			} else {
				fmt.Println("Commit is not valid")
				os.Exit(1)
			}
		},
	}
	validate.Flags().StringVarP(&hash, "commit", "c", "", "hash commit which need validate")
	validate.Flags().StringVarP(&msg, "message", "m", "", "message which need validate")

	installHook := &cobra.Command{
		Use:   "install-hook",
		Short: "Install commit validation hook to your project",
		Run: func(cmd *cobra.Command, args []string) {
		  app.InstallHook()
		},
	}

	var rootCmd = &cobra.Command{
		Use: "git-convenient",
		Version: fmt.Sprintf("%s (%s)", version, revision),
	}
	rootCmd.AddCommand(commit, validate, installHook)
	rootCmd.Execute()

}
