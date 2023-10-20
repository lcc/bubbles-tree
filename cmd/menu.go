package cmd

import (
	"bubbles-tree/pkg"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

const (
	IndigoBlue = "#89CFF0"
	Red        = "#FF0000"
	White      = "#FFFFFF"
)

var parentStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(IndigoBlue))
var secretsStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(Red))
var leafStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(White))

type Namespace struct {
	name     string
	children []*Secrets
}

func (n *Namespace) Name() string {
	return n.name
}

func (n *Namespace) Print() string {
	return fmt.Sprint(parentStyle.Render(n.name))
}

func (n *Namespace) Update() {}

func (n *Namespace) Children() []*Secrets {
	return n.children
}

type Secrets struct {
	name     string
	children []*Values
}

func (n *Secrets) Name() string {
	return n.name
}

func (n *Secrets) Print() string {
	return fmt.Sprint(secretsStyle.Render(n.name))
}

func (n *Secrets) Update() {}

func (n *Secrets) Children() []*Values {
	return n.children
}

type Values struct {
	name     string
	Selected bool
}

func (n *Values) Name() string {
	if n.Selected {
		return n.name + " (x)"
	}
	return n.name + " ( )"
}

func (n *Values) Print() string {
	return leafStyle.Render(n.Name())
}

func (n *Values) Update() {
	n.Selected = !n.Selected
}

var namespaceCmd = &cobra.Command{
	Use: "namespace",
	Run: func(cmd *cobra.Command, args []string) {
		namespace := []*Namespace{
			{
				name: "namespace 1",
				children: []*Secrets{
					{
						name: "secrets",
						children: []*Values{
							{
								name: "value1",
							},
							{
								name: "value2",
							},
						},
					},
				},
			},
			{
				name: "namespace 2",
				children: []*Secrets{
					{
						name: "secrets",
						children: []*Values{
							{
								name: "valu3",
							},
							{
								name: "value4",
							},
						},
					},
				},
			},
		}

		tree := pkg.NewTree(namespace)
		p := tea.NewProgram(tree)
		if _, err := p.Run(); err != nil {
			panic(err)
		}

		for _, namespace := range namespace {
			for _, secret := range namespace.children {
				for _, value := range secret.children {
					if value.Selected {
						println(value.name)
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(namespaceCmd)
}
