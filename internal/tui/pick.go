package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/villawebcl/supashift/internal/model"
)

type item struct {
	name  string
	desc  string
	tags  string
	rank  int
	alias string
}

func (i item) FilterValue() string { return i.name + " " + i.tags + " " + i.alias }
func (i item) Title() string       { return i.name }
func (i item) Description() string { return i.desc }

type modelPick struct {
	list   list.Model
	choice string
	quit   bool
}

func (m modelPick) Init() tea.Cmd { return nil }

func (m modelPick) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if it, ok := m.list.SelectedItem().(item); ok {
				m.choice = it.name
				m.quit = true
				return m, tea.Quit
			}
		case "ctrl+c", "q":
			m.quit = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m modelPick) View() string {
	if m.quit && m.choice != "" {
		return ""
	}
	return m.list.View()
}

func PickProfile(cfg *model.Config) (string, error) {
	items := make([]list.Item, 0, len(cfg.Profiles))
	recentRank := map[string]int{}
	for i, p := range cfg.Recents {
		recentRank[p] = i
	}
	for name, p := range cfg.Profiles {
		desc := p.AccountLabel
		if p.Notes != "" {
			desc = fmt.Sprintf("%s | %s", p.AccountLabel, p.Notes)
		}
		rank := 9999
		if p.Favorite {
			rank = -100
		}
		if r, ok := recentRank[name]; ok {
			rank = r
		}
		items = append(items, item{name: name, desc: desc, tags: strings.Join(p.Tags, ","), alias: strings.Join(p.Aliases, ","), rank: rank})
	}
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("10"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("10"))
	l := list.New(items, delegate, 80, 20)
	l.Title = "supashift pick: perfil"
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(true)

	m := modelPick{list: l}
	p := tea.NewProgram(m)
	fin, err := p.Run()
	if err != nil {
		return "", err
	}
	res := fin.(modelPick)
	if res.choice == "" {
		return "", fmt.Errorf("selección cancelada")
	}
	return res.choice, nil
}
