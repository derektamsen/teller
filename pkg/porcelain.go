package pkg

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/thoas/go-funk"

	"github.com/fatih/color"
	"github.com/jftuga/ellipsis"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/utils"
)

type Porcelain struct {
	Out io.Writer
}

func (p *Porcelain) StartWizard() (*core.WizardAnswers, error) {
	wd, _ := os.Getwd()
	workingfolder := utils.LastSegment(wd)

	providers := BuiltinProviders{}
	providerNames := providers.ProviderHumanToMachine()
	// the questions to ask
	var qs = []*survey.Question{
		{
			Name: "project",
			Prompt: &survey.Input{
				Message: "Project name?",
				Default: workingfolder,
			},
			Validate: survey.Required,
		},
		{
			Name: "providers",
			Prompt: &survey.MultiSelect{
				Message:  "Select your secret providers",
				PageSize: 10,
				Options:  funk.Keys(providerNames).([]string),
			},
		},
		{
			Name: "confirm",
			Prompt: &survey.Confirm{
				Message: "Would you like extra confirmation before accessing secrets?",
			},
		},
	}

	answers := core.WizardAnswers{}
	// perform the questions
	err := survey.Ask(qs, &answers)
	if err != nil {
		return nil, err
	}

	if len(answers.Providers) == 0 {
		return nil, fmt.Errorf("you need to select at least one provider")
	}

	answers.ProviderKeys = map[string]bool{}
	for _, plabel := range answers.Providers {
		answers.ProviderKeys[providerNames[plabel]] = true
	}

	return &answers, nil
}

func (p *Porcelain) DidCreateNewFile(fname string) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Fprintf(p.Out, "Created file: %v\n", green(fname))
}

func (p *Porcelain) AskForConfirmation(msg string) bool {
	yesno := false
	prompt := &survey.Confirm{
		Message: msg,
	}
	_ = survey.AskOne(prompt, &yesno)
	return yesno
}

func (p *Porcelain) VSpace(size int) {
	fmt.Fprint(p.Out, strings.Repeat("\n", size))
}

func (p *Porcelain) PrintContext(projectName, loadedFrom string) {
	green := color.New(color.FgGreen).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	fmt.Fprintf(p.Out, "-*- %s: loaded variables for %s using %s -*-\n", white("teller"), green(projectName), green(loadedFrom))
}

func (p *Porcelain) PrintEntries(entries []core.EnvEntry) {
	var buf bytes.Buffer
	yellow := color.New(color.FgYellow).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	for _, v := range entries {
		ep := ellipsis.Shorten(v.ResolvedPath, 30)
		if v.Value == "" {
			fmt.Fprintf(&buf, "[%s %s %s] %s\n", yellow(v.Provider), gray(ep), red("missing"), green(v.Key))
		} else {
			fmt.Fprintf(&buf, "[%s %s] %s %s %s*****\n", yellow(v.Provider), gray(ep), green(v.Key), gray("="), v.Value[:int(math.Min(float64(len(v.Value)), 2))])
		}
	}

	out := buf.String()

	fmt.Fprint(p.Out, out)
}
