package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/felixtorres/az-dash/internal/azdo"
	"github.com/felixtorres/az-dash/internal/config"
	"github.com/felixtorres/az-dash/internal/tui/keys"
	"github.com/felixtorres/az-dash/internal/tui/theme"
	"github.com/felixtorres/az-dash/internal/tui/views"
)

type ViewType int

const (
	ViewPRs ViewType = iota
	ViewWorkItems
	ViewPipelines
)

type bootstrapDoneMsg struct{ err error }
type fetchDoneMsg struct {
	view ViewType
	section int
	err  error
}

type Model struct {
	cfg       *config.Config
	client    *azdo.Client
	theme     *theme.Theme
	spinner   spinner.Model
	activeView ViewType
	prView     *views.PRView
	wiView     *views.WorkItemView
	pipeView   *views.PipelineView
	width      int
	height     int
	ready      bool
	bootErr    error
	statusMsg  string
}

func Start(cfg *config.Config) error {
	var auth azdo.AuthProvider
	switch cfg.Auth.Method {
	case "pat":
		auth = azdo.NewPatAuth(cfg.Auth.PAT)
	default:
		auth = azdo.NewAzCliAuth()
	}

	client := azdo.NewClient(cfg.BaseURL, cfg.Organization, cfg.Project, auth)
	th := theme.New(cfg.Theme)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = th.Spinner

	var activeView ViewType
	switch cfg.Defaults.View {
	case "workitems":
		activeView = ViewWorkItems
	case "pipelines":
		activeView = ViewPipelines
	default:
		activeView = ViewPRs
	}

	m := Model{
		cfg:        cfg,
		client:     client,
		theme:      th,
		spinner:    s,
		activeView: activeView,
		prView:     views.NewPRView(cfg.PRSections, th),
		wiView:     views.NewWorkItemView(cfg.WorkItemSections, th),
		pipeView:   views.NewPipelineView(cfg.PipelineSections, th),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.bootstrap,
	)
}

func (m Model) bootstrap() tea.Msg {
	if err := m.client.Bootstrap(); err != nil {
		return bootstrapDoneMsg{err: err}
	}
	return bootstrapDoneMsg{}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.prView.SetSize(m.contentWidth(), m.contentHeight())
		m.wiView.SetSize(m.contentWidth(), m.contentHeight())
		m.pipeView.SetSize(m.contentWidth(), m.contentHeight())

	case bootstrapDoneMsg:
		if msg.err != nil {
			m.bootErr = msg.err
			return m, nil
		}
		m.ready = true
		cmds = append(cmds, m.fetchActiveView())

	case fetchDoneMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if m.bootErr != nil {
			if key.Matches(msg, keys.Global.Quit) {
				return m, tea.Quit
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, keys.Global.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Global.ViewPRs):
			m.activeView = ViewPRs
			cmds = append(cmds, m.fetchActiveView())
		case key.Matches(msg, keys.Global.ViewWorkItems):
			m.activeView = ViewWorkItems
			cmds = append(cmds, m.fetchActiveView())
		case key.Matches(msg, keys.Global.ViewPipelines):
			m.activeView = ViewPipelines
			cmds = append(cmds, m.fetchActiveView())
		case key.Matches(msg, keys.Global.Refresh):
			cmds = append(cmds, m.fetchActiveSection())
		case key.Matches(msg, keys.Global.RefreshAll):
			cmds = append(cmds, m.fetchActiveView())
		case key.Matches(msg, keys.Global.NextSection):
			m.activeViewPtr().NextSection()
		case key.Matches(msg, keys.Global.PrevSection):
			m.activeViewPtr().PrevSection()
		case key.Matches(msg, keys.Global.Up):
			m.activeViewPtr().CursorUp()
		case key.Matches(msg, keys.Global.Down):
			m.activeViewPtr().CursorDown()
		case key.Matches(msg, keys.Global.FirstLine):
			m.activeViewPtr().CursorFirst()
		case key.Matches(msg, keys.Global.LastLine):
			m.activeViewPtr().CursorLast()
		case key.Matches(msg, keys.Global.TogglePreview):
			m.activeViewPtr().TogglePreview()
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.bootErr != nil {
		return m.renderError(m.bootErr)
	}
	if !m.ready {
		return m.renderLoading("Connecting to Azure DevOps...")
	}

	var b strings.Builder

	// Tab bar
	b.WriteString(m.renderTabs())
	b.WriteString("\n")

	// Active view
	switch m.activeView {
	case ViewPRs:
		b.WriteString(m.prView.View())
	case ViewWorkItems:
		b.WriteString(m.wiView.View())
	case ViewPipelines:
		b.WriteString(m.pipeView.View())
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	return b.String()
}

func (m *Model) renderTabs() string {
	tabs := []struct {
		name string
		view ViewType
	}{
		{"Pull Requests", ViewPRs},
		{"Work Items", ViewWorkItems},
		{"Pipelines", ViewPipelines},
	}

	var rendered []string
	for _, t := range tabs {
		if t.view == m.activeView {
			rendered = append(rendered, m.theme.ActiveTab.Render(t.name))
		} else {
			rendered = append(rendered, m.theme.InactiveTab.Render(t.name))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (m *Model) renderStatusBar() string {
	left := m.theme.StatusBar.Render(fmt.Sprintf(" %s/%s", m.cfg.Organization, m.cfg.Project))
	right := m.theme.StatusBar.Render("? help  q quit ")

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return left + strings.Repeat(" ", gap) + right
}

func (m *Model) renderLoading(msg string) string {
	return fmt.Sprintf("\n  %s %s\n", m.spinner.View(), msg)
}

func (m *Model) renderError(err error) string {
	return fmt.Sprintf("\n  %s\n\n  Press q to quit.\n",
		m.theme.Error.Render(fmt.Sprintf("Error: %v", err)))
}

func (m *Model) contentWidth() int {
	return m.width
}

func (m *Model) contentHeight() int {
	return m.height - 3 // tabs + status bar + spacing
}

func (m *Model) activeViewPtr() views.View {
	switch m.activeView {
	case ViewWorkItems:
		return m.wiView
	case ViewPipelines:
		return m.pipeView
	default:
		return m.prView
	}
}

func (m *Model) fetchActiveView() tea.Cmd {
	switch m.activeView {
	case ViewPRs:
		return m.fetchPRs()
	case ViewWorkItems:
		return m.fetchWorkItems()
	case ViewPipelines:
		return m.fetchBuilds()
	}
	return nil
}

func (m *Model) fetchActiveSection() tea.Cmd {
	return m.fetchActiveView()
}

func (m *Model) fetchPRs() tea.Cmd {
	return func() tea.Msg {
		for i, section := range m.cfg.PRSections {
			criteria := azdo.PRSearchCriteria{
				Status:     section.Filters.Status,
				CreatorID:  section.Filters.CreatorID,
				ReviewerID: section.Filters.ReviewerID,
				Repository: section.Filters.Repository,
				TargetRef:  section.Filters.TargetBranch,
				SourceRef:  section.Filters.SourceRefName,
			}
			org := section.Organization
			project := section.Project

			prs, err := m.client.ListPullRequests(org, project, criteria, section.Limit)
			if err != nil {
				m.prView.SetSectionError(i, err)
			} else {
				m.prView.SetSectionData(i, prs)
			}
		}
		return fetchDoneMsg{view: ViewPRs}
	}
}

func (m *Model) fetchWorkItems() tea.Cmd {
	return func() tea.Msg {
		for i, section := range m.cfg.WorkItemSections {
			org := section.Organization
			project := section.Project

			items, err := m.client.QueryWorkItems(org, project, section.WIQL, section.Limit)
			if err != nil {
				m.wiView.SetSectionError(i, err)
			} else {
				m.wiView.SetSectionData(i, items)
			}
		}
		return fetchDoneMsg{view: ViewWorkItems}
	}
}

func (m *Model) fetchBuilds() tea.Cmd {
	return func() tea.Msg {
		for i, section := range m.cfg.PipelineSections {
			criteria := azdo.BuildSearchCriteria{
				DefinitionID: section.Filters.DefinitionID,
				RequestedFor: section.Filters.RequestedFor,
				StatusFilter: section.Filters.StatusFilter,
				ResultFilter: section.Filters.ResultFilter,
				BranchName:   section.Filters.BranchName,
			}
			org := section.Organization
			project := section.Project

			builds, err := m.client.ListBuilds(org, project, criteria, section.Limit)
			if err != nil {
				m.pipeView.SetSectionError(i, err)
			} else {
				m.pipeView.SetSectionData(i, builds)
			}
		}
		return fetchDoneMsg{view: ViewPipelines}
	}
}
