package pkg

import (
	"fmt"
	"reflect"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const SPACE = " "

// Common interface for all nodes
type NodeInterface interface {
	Name() string
	Print() string
	Update()
}

func NewTree(childProvider interface{}) *CLITree {
	var rootInterface NodeInterface
	tree := newTree(rootInterface, childProvider, "")
	return &CLITree{
		tree:    *tree,
		cursors: []int{0},
	}
}

// Node represents a node in the tree
type Node struct {
	ID       string
	Value    NodeInterface
	Children []*Node
	Parent   *Node
}

func (n *Node) Update() {
	n.Value.Update()
}

func newNode(value NodeInterface) *Node {
	return &Node{
		Value:    value,
		Children: []*Node{},
	}
}

func (n *Node) branch() *Node {
	if n.Parent.isRoot() {
		return n
	}
	return n.Parent.branch()
}

func (n *Node) find(id string) *Node {
	if n.ID == id {
		return n
	}
	for i := range n.Children {
		if node := n.Children[i].find(id); node != nil {
			return node
		}
	}
	return nil
}

func (n *Node) addChild(child *Node, index int) {
	child.Parent = n
	child.ID = n.ID + fmt.Sprint(index)
	n.Children = append(n.Children, child)
}

func (n *Node) next() *Node {
	for i := range n.Parent.Children {
		if n.Parent.Children[i].ID == n.ID {
			if i == len(n.Parent.Children)-1 {
				return nil
			}
			return n.Parent.Children[i+1]
		}
	}
	return nil
}

func (n *Node) previous() *Node {
	if n.isRoot() {
		return nil
	}
	for i := range n.Parent.Children {
		if n.Parent.Children[i].ID == n.ID {
			if i == 0 {
				return nil
			}
			return n.Parent.Children[i-1]
		}
	}
	return nil
}

func (n *Node) isRoot() bool {
	return n.Parent == nil
}

func (n *Node) isLeaf() bool {
	return len(n.Children) == 0
}

// NewTree constructs a tree from a root and a struct with a Children method
func newTree(parent NodeInterface, childProvider interface{}, ID string) *Node {
	parentNode := newNode(parent)
	parentNode.ID = ID

	if childProvider != nil {
		// Check if childProvider is a single child or an array of children
		if reflect.TypeOf(childProvider).Kind() == reflect.Slice {
			childrenSlice := reflect.ValueOf(childProvider)
			for i := 0; i < childrenSlice.Len(); i++ {
				childValue := childrenSlice.Index(i).Interface()
				id := parentNode.ID + fmt.Sprint(i)
				childNode := newTree(childValue.(NodeInterface), childValue, id)
				parentNode.addChild(childNode, i)
			}
		} else {
			childMethod := reflect.ValueOf(childProvider).MethodByName("Children")
			if childMethod.IsValid() && childMethod.Type().NumOut() == 1 {
				children := childMethod.Call(nil)[0]
				if children.Kind() == reflect.Slice {
					for i := 0; i < children.Len(); i++ {
						childValue := children.Index(i).Interface()
						id := parentNode.ID + fmt.Sprint(i)
						childNode := newTree(childValue.(NodeInterface), childValue, id)
						parentNode.addChild(childNode, i)
					}
				}
			}
		}
	}
	return parentNode
}

type CLITree struct {
	tree    Node
	cursors cursos
}

type cursos []int

func (m *CLITree) previous() {
	lastCursor := len(m.cursors) - 1
	m.cursors[lastCursor] = m.cursors[lastCursor] - 1
}

func (m *CLITree) next() {
	lastCursor := len(m.cursors) - 1
	m.cursors[lastCursor] = m.cursors[lastCursor] + 1
}

func (m *CLITree) nextLevel() {
	m.cursors = append(m.cursors, 0)
}

func (m *CLITree) previousLevel() {
	m.cursors = m.cursors[:len(m.cursors)-1]
}

func (m CLITree) cursor() string {
	// convert cusors to a string

	ret := ""
	for i := range m.cursors {
		ret += fmt.Sprintf("%d", m.cursors[i])
	}

	return ret
}

func (m CLITree) Init() tea.Cmd {
	return nil
}

func (m CLITree) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		curr := m.tree.find(m.cursor())
		switch msg.String() {
		case "ctrl+c", "q", "enter":
			return m, tea.Quit
		case "up", "k":
			previous := curr.previous()
			if previous == nil {
				return m, nil
			}
			m.previous()
			return m, nil
		case "down", "j":
			previous := curr.next()
			if previous == nil {
				return m, nil
			}
			m.next()
			return m, nil
		case "l", "right":
			if curr.isLeaf() {
				curr.Update()
				return m, nil
			}
			m.nextLevel()
			// m.cursor = curr.Children[0].ID
			return m, nil
		case "delete", "left", "h", "backspace":
			if curr.Parent.isRoot() {
				return m, nil
			}
			m.previousLevel()
			// m.cursor = curr.Parent.ID
			return m, nil
		}
	}
	return m, nil
}

func (m CLITree) View() string {
	s := fmt.Sprintf("Select items: %v \n\n", m.cursors)
	curr := m.tree.find(m.cursor())

	renderMax := 10
	startAt := nonNegative(m.cursors[len(m.cursors)-1] - renderMax)
	s += fmt.Sprint(startAt) + "\n"
	for i, choice := range m.tree.Children {
		cursor := ""
		if choice.ID == curr.branch().ID {
			ret := fmt.Sprintf("%s\n", m.displayChildren(*choice, m.cursor()))
			s += keepRangeAfterNNewlines(ret, startAt, renderMax) + "\n"
			// ignore everything before startAt \n
			continue
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice.Value.Name())

		if i == renderMax {
			break
		}

	}
	s += "\nPress q or ctrl-c to quit.\n"
	return s
}

func (m CLITree) displayChildren(n Node, cursor string) string {
	ret := " "
	if n.ID == cursor {
		ret = ">"
	}

	if n.isLeaf() {
		return ret + n.Value.Print() + "\n"
	}

	// startAt := nonNegative(m.cursors[len(m.cursors)-1] - renderMax)
	for i := range n.Children {
		child := n.Children[i]
		identation := recursiveIdentation(*child)
		name := m.displayChildren(*child, cursor)

		if i == 0 {
			identation := ident(*child)
			parentIdentation := parentIdentation(*child.Parent, identation)
			ret += n.Value.Print() + parentIdentation + name
			continue
		}
		ret += identation + name
	}

	return ret + "\n"
}

func ident(n Node) string {
	if n.Parent.isRoot() {
		return ""
	}

	identationNode := biggestAncestor(n)
	return sameSizeStrWhiteSpaces(identationNode.Value.Name())
}

func parentIdentation(n Node, identation string) string {
	if n.isRoot() {
		return ""
	}

	parentIdentation := ""
	for i := 0; i < len(identation)-len(n.Value.Name())-1; i++ {
		parentIdentation += SPACE
	}

	return parentIdentation
}

func recursiveIdentation(n Node) string {
	if n.Parent.isRoot() {
		return ""
	}

	identation := ""
	identationNode := biggestAncestor(n)
	identation += recursiveIdentation(*identationNode)
	return identation + sameSizeStrWhiteSpaces(identationNode.Value.Name())
}

func biggestAncestor(n Node) *Node {
	identNode := n.Parent
	directAncestors := getDirectAncestors(n)
	for i := range directAncestors {
		sibling := directAncestors[i]
		if len(sibling.Value.Name()) > len(identNode.Value.Name()) {
			identNode = sibling
		}
	}
	return identNode
}

func getDirectAncestors(n Node) []*Node {
	if n.Parent.isRoot() {
		return n.Children
	}
	return n.Parent.Parent.Children
}

func sameSizeStrWhiteSpaces(s string) string {
	ret := " "
	for range s {
		ret += SPACE
	}
	return ret
}

func nonNegative(val int) int {
	if val < 0 {
		return 0
	}
	return val
}

// keepRangeAfterNNewlines keeps a specific range of lines after ignoring the first n newlines.
func keepRangeAfterNNewlines(s string, n int, x int) string {
	// Split the string by newline characters
	lines := strings.Split(s, "\n")

	// If n is greater than the number of lines, return an empty string
	if n >= len(lines) {
		return ""
	}

	// Calculate the end index, but ensure it doesn't exceed the total number of lines
	end := n + x
	if end > len(lines) {
		end = len(lines)
	}

	// Join and return only the lines within the range after skipping the first n
	return strings.Join(lines[n:end], "\n")
}
