package main

import (
	"fmt"
	"regexp"
	"strings"
	"os/exec"

	// "gopkg.in/src-d/go-git.v4"
	"gopkg.in/yaml.v2"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
)

type OptionalPart struct {
	Skip bool
}

type Pattern struct {
	Match string
	Replace string
}

type Scope struct {
	OptionalPart
	Pattern Pattern
}

type Commit struct {
	Scope Scope
	Body OptionalPart
	Footer OptionalPart
}

type Config struct {
	Commit Commit
}

type App struct {
	Config Config
}

type CommitType struct {
	Value       string
	Description string
}

type Scopes []string

func (a *App) SelectCommitType() CommitType {
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
		Selected: `{{ "\u2714" | green }} {{ .Value | cyan | bold }}`,
	}

	promptType := promptui.Select{
		Label:     "Commit Type",
		Items:     types,
		Templates: templatesType,
		Size:      len(types),
	}

	i, _, err := promptType.Run()
	if err != nil {
		log.Panic("Incorrect commit type\nDetails: %v\n", err)
	}
	return types[i]
}

func (a *App) DetectScope() Scopes {
    
}

func main() {
	


	// diff, err := exec.Command("git", "diff", "--staged", "--name-status").Output()
	// if err != nil {
	// 	fmt.Printf("Git diff %v\n", err)
	// 	return
	// }
	// var scopes []string
	// re := regexp.MustCompile("^.*\\.([\\S]{2,3})$") // TODO: move it to config

	// files := strings.Split(strings.Trim(string(diff), "\n"), "\n")
	// for _, f := range files {
	// 	f = strings.Trim(f, " ")
	// 	if re.MatchString(f) {
	// 		scopes = append(scopes, re.ReplaceAllString(f, "$1")) // TODO: move it to config
	// 	}
	// }
	// templatesMessage := &promptui.PromptTemplates{
	// 	Prompt:  "{{ . }}: ",
	// 	Valid:   "{{ . }}: ",
	// 	Invalid: "{{ . }}: ",
	// 	Success: `{{ "\u2714" | green | bold }} `,
	// }
	// promptMessage := promptui.Prompt{
	// 	Label:     "Commit Message",
	// 	Templates: templatesMessage,
	// }
	// msg, _ := promptMessage.Run()

	// // TODO: IS THIS SCOPE CORRECT ?
	// // TODO: PROMPT FOR BODY AND FOOTER=
	// // TODO: JIRA CONNECTOR
	// if len(scopes) != 0 {
	// 	fmt.Printf("%s(%s): %s", types[i].Value, strings.Join(scopes, ","), msg)
	// } else {
	// 	fmt.Printf("%s: %s", types[i].Value, msg)
	// }
}
