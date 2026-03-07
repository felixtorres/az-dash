package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/felixtorres/az-dash/internal/azdo"
	"github.com/felixtorres/az-dash/internal/config"
	"github.com/felixtorres/az-dash/internal/tui/keys"
	"github.com/felixtorres/az-dash/internal/tui/theme"
	"github.com/felixtorres/az-dash/internal/tui/views"
	"github.com/felixtorres/az-dash/internal/utils"
)

type ViewType int

const (
	ViewPRs ViewType = iota
	ViewWorkItems
	ViewPipelines
)

// Messages — all state changes flow through these.
type bootstrapDoneMsg struct{ err error }

type prsFetchedMsg struct {
	section int
	data    []azdo.PullRequest
	err     error
}

type workItemsFetchedMsg struct {
	section int
	data    []azdo.WorkItem
	err     error
}

type buildsFetchedMsg struct {
	section int
	data    []azdo.Build
	err     error
}

type actionResultMsg struct {
	msg string
	err error
}

type refreshTickMsg struct{}

type Model struct {
	cfg        *config.Config
	client     *azdo.Client
	theme      *theme.Theme
	spinner    spinner.Model
	activeView ViewType
	prView     *views.PRView
	wiView     *views.WorkItemView
	pipeView   *views.PipelineView
	width      int
	height     int
	ready      bool
	bootErr    error
	statusMsg  string
	showHelp   bool
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
		m.scheduleRefresh(),
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
		cmds = append(cmds, m.fetchAllSections()...)

	case prsFetchedMsg:
		if msg.err != nil {
			m.prView.SetSectionError(msg.section, msg.err)
		} else {
			m.prView.SetSectionData(msg.section, msg.data)
		}

	case workItemsFetchedMsg:
		if msg.err != nil {
			m.wiView.SetSectionError(msg.section, msg.err)
		} else {
			m.wiView.SetSectionData(msg.section, msg.data)
		}

	case buildsFetchedMsg:
		if msg.err != nil {
			m.pipeView.SetSectionError(msg.section, msg.err)
		} else {
			m.pipeView.SetSectionData(msg.section, msg.data)
		}

	case actionResultMsg:
		if msg.err != nil {
			m.statusMsg = m.theme.Error.Render(fmt.Sprintf("Error: %v", msg.err))
		} else {
			m.statusMsg = msg.msg
		}
		// Re-fetch after actions to reflect changes
		if msg.err == nil && msg.msg != "" && msg.msg != "Opened in browser" && msg.msg != "Diff closed" {
			cmds = append(cmds, m.fetchCurrentSection())
		}

	case refreshTickMsg:
		if m.ready {
			cmds = append(cmds, m.fetchAllSections()...)
		}
		cmds = append(cmds, m.scheduleRefresh())

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

		cmd := m.handleKeypress(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleKeypress(msg tea.KeyMsg) tea.Cmd {
	// Help toggle
	if m.showHelp {
		if key.Matches(msg, keys.Global.Quit) || key.Matches(msg, keys.Global.Help) || msg.String() == "esc" {
			m.showHelp = false
		}
		return nil
	}

	// Global keys
	switch {
	case key.Matches(msg, keys.Global.Help):
		m.showHelp = true
		return nil
	case key.Matches(msg, keys.Global.Quit):
		return tea.Quit
	case key.Matches(msg, keys.Global.ViewPRs):
		m.activeView = ViewPRs
		return nil
	case key.Matches(msg, keys.Global.ViewWorkItems):
		m.activeView = ViewWorkItems
		return nil
	case key.Matches(msg, keys.Global.ViewPipelines):
		m.activeView = ViewPipelines
		return nil
	case key.Matches(msg, keys.Global.Refresh):
		return m.fetchCurrentSection()
	case key.Matches(msg, keys.Global.RefreshAll):
		return tea.Batch(m.fetchAllSections()...)
	case key.Matches(msg, keys.Global.NextSection):
		m.activeViewPtr().NextSection()
		return nil
	case key.Matches(msg, keys.Global.PrevSection):
		m.activeViewPtr().PrevSection()
		return nil
	case key.Matches(msg, keys.Global.Up):
		m.activeViewPtr().CursorUp()
		return nil
	case key.Matches(msg, keys.Global.Down):
		m.activeViewPtr().CursorDown()
		return nil
	case key.Matches(msg, keys.Global.FirstLine):
		m.activeViewPtr().CursorFirst()
		return nil
	case key.Matches(msg, keys.Global.LastLine):
		m.activeViewPtr().CursorLast()
		return nil
	case key.Matches(msg, keys.Global.PageDown):
		m.activeViewPtr().PageDown()
		return nil
	case key.Matches(msg, keys.Global.PageUp):
		m.activeViewPtr().PageUp()
		return nil
	case key.Matches(msg, keys.Global.TogglePreview):
		m.activeViewPtr().TogglePreview()
		return nil
	case key.Matches(msg, keys.Global.OpenBrowser):
		return m.openInBrowser()
	case key.Matches(msg, keys.Global.CopyURL):
		return m.copyURL()
	case key.Matches(msg, keys.Global.CopyID):
		return m.copyID()
	}

	// View-specific keys
	switch m.activeView {
	case ViewPRs:
		return m.handlePRKey(msg)
	case ViewWorkItems:
		return m.handleWorkItemKey(msg)
	case ViewPipelines:
		return m.handlePipelineKey(msg)
	}

	return nil
}

func (m *Model) handlePRKey(msg tea.KeyMsg) tea.Cmd {
	pr := m.prView.SelectedPR()
	if pr == nil {
		return nil
	}

	switch {
	case key.Matches(msg, keys.PR.Approve):
		return m.approvePR(pr)
	case key.Matches(msg, keys.PR.Complete):
		return m.completePR(pr)
	case key.Matches(msg, keys.PR.Abandon):
		return m.abandonPR(pr)
	case key.Matches(msg, keys.PR.Reopen):
		return m.reopenPR(pr)
	case key.Matches(msg, keys.PR.Diff):
		return m.viewDiff(pr)
	}

	return nil
}

func (m *Model) handleWorkItemKey(msg tea.KeyMsg) tea.Cmd {
	wi := m.wiView.SelectedWorkItem()
	if wi == nil {
		return nil
	}

	switch {
	case key.Matches(msg, keys.WorkItem.Assign):
		return m.assignWorkItem(wi)
	case key.Matches(msg, keys.WorkItem.Unassign):
		return m.unassignWorkItem(wi)
	case key.Matches(msg, keys.WorkItem.ChangeState):
		return m.cycleWorkItemState(wi)
	}

	return nil
}

func (m *Model) handlePipelineKey(msg tea.KeyMsg) tea.Cmd {
	build := m.pipeView.SelectedBuild()
	if build == nil {
		return nil
	}

	switch {
	case key.Matches(msg, keys.Pipeline.Cancel):
		return m.cancelBuild(build)
	case key.Matches(msg, keys.Pipeline.Logs):
		return m.viewBuildLogs(build)
	}

	return nil
}

// --- Actions ---

func (m *Model) openInBrowser() tea.Cmd {
	url := m.selectedURL()
	if url == "" {
		return nil
	}
	return func() tea.Msg {
		err := utils.OpenBrowser(url)
		if err != nil {
			return actionResultMsg{err: fmt.Errorf("open browser: %w", err)}
		}
		return actionResultMsg{msg: "Opened in browser"}
	}
}

func (m *Model) copyURL() tea.Cmd {
	url := m.selectedURL()
	if url == "" {
		return nil
	}
	return func() tea.Msg {
		if err := utils.CopyToClipboard(url); err != nil {
			return actionResultMsg{err: fmt.Errorf("copy: %w", err)}
		}
		return actionResultMsg{msg: "URL copied"}
	}
}

func (m *Model) copyID() tea.Cmd {
	var id string
	switch m.activeView {
	case ViewPRs:
		if pr := m.prView.SelectedPR(); pr != nil {
			id = fmt.Sprintf("%d", pr.PullRequestID)
		}
	case ViewWorkItems:
		if wi := m.wiView.SelectedWorkItem(); wi != nil {
			id = fmt.Sprintf("%d", wi.ID)
		}
	case ViewPipelines:
		if b := m.pipeView.SelectedBuild(); b != nil {
			id = fmt.Sprintf("%d", b.ID)
		}
	}
	if id == "" {
		return nil
	}
	return func() tea.Msg {
		if err := utils.CopyToClipboard(id); err != nil {
			return actionResultMsg{err: fmt.Errorf("copy: %w", err)}
		}
		return actionResultMsg{msg: fmt.Sprintf("Copied #%s", id)}
	}
}

func (m *Model) approvePR(pr *azdo.PullRequest) tea.Cmd {
	client := m.client
	userID := client.UserID()
	repoID := pr.Repository.ID
	prID := pr.PullRequestID

	return func() tea.Msg {
		err := client.VotePullRequest("", "", repoID, prID, userID, 10)
		if err != nil {
			return actionResultMsg{err: fmt.Errorf("approve PR #%d: %w", prID, err)}
		}
		return actionResultMsg{msg: fmt.Sprintf("Approved PR #%d", prID)}
	}
}

func (m *Model) completePR(pr *azdo.PullRequest) tea.Cmd {
	client := m.client
	repoID := pr.Repository.ID
	prID := pr.PullRequestID

	return func() tea.Msg {
		err := client.UpdatePullRequestStatus("", "", repoID, prID, "completed")
		if err != nil {
			return actionResultMsg{err: fmt.Errorf("complete PR #%d: %w", prID, err)}
		}
		return actionResultMsg{msg: fmt.Sprintf("Completed PR #%d", prID)}
	}
}

func (m *Model) abandonPR(pr *azdo.PullRequest) tea.Cmd {
	client := m.client
	repoID := pr.Repository.ID
	prID := pr.PullRequestID

	return func() tea.Msg {
		err := client.UpdatePullRequestStatus("", "", repoID, prID, "abandoned")
		if err != nil {
			return actionResultMsg{err: fmt.Errorf("abandon PR #%d: %w", prID, err)}
		}
		return actionResultMsg{msg: fmt.Sprintf("Abandoned PR #%d", prID)}
	}
}

func (m *Model) reopenPR(pr *azdo.PullRequest) tea.Cmd {
	client := m.client
	repoID := pr.Repository.ID
	prID := pr.PullRequestID

	return func() tea.Msg {
		err := client.UpdatePullRequestStatus("", "", repoID, prID, "active")
		if err != nil {
			return actionResultMsg{err: fmt.Errorf("reactivate PR #%d: %w", prID, err)}
		}
		return actionResultMsg{msg: fmt.Sprintf("Reactivated PR #%d", prID)}
	}
}

func (m *Model) viewDiff(pr *azdo.PullRequest) tea.Cmd {
	client := m.client
	repoID := pr.Repository.ID
	prID := pr.PullRequestID

	return tea.ExecProcess(
		utils.DiffCommand(client, repoID, prID),
		func(err error) tea.Msg {
			if err != nil {
				return actionResultMsg{err: fmt.Errorf("diff: %w", err)}
			}
			return actionResultMsg{msg: "Diff closed"}
		},
	)
}

// --- Work Item Actions ---

func (m *Model) assignWorkItem(wi *azdo.WorkItem) tea.Cmd {
	client := m.client
	wiID := wi.ID
	userName := client.UserDisplayName()

	return func() tea.Msg {
		err := client.UpdateWorkItemField("", "", wiID, "System.AssignedTo", userName)
		if err != nil {
			return actionResultMsg{err: fmt.Errorf("assign WI #%d: %w", wiID, err)}
		}
		return actionResultMsg{msg: fmt.Sprintf("Assigned WI #%d to you", wiID)}
	}
}

func (m *Model) unassignWorkItem(wi *azdo.WorkItem) tea.Cmd {
	client := m.client
	wiID := wi.ID

	return func() tea.Msg {
		err := client.UpdateWorkItemField("", "", wiID, "System.AssignedTo", "")
		if err != nil {
			return actionResultMsg{err: fmt.Errorf("unassign WI #%d: %w", wiID, err)}
		}
		return actionResultMsg{msg: fmt.Sprintf("Unassigned WI #%d", wiID)}
	}
}

func (m *Model) cycleWorkItemState(wi *azdo.WorkItem) tea.Cmd {
	client := m.client
	wiID := wi.ID
	currentState := wi.StringField("System.State")

	// Cycle: New → Active → Resolved → Closed → New
	nextState := map[string]string{
		"New":      "Active",
		"Active":   "Resolved",
		"Resolved": "Closed",
		"Closed":   "New",
		"To Do":    "Doing",
		"Doing":    "Done",
		"Done":     "To Do",
	}

	next, ok := nextState[currentState]
	if !ok {
		next = "Active"
	}

	return func() tea.Msg {
		err := client.UpdateWorkItemField("", "", wiID, "System.State", next)
		if err != nil {
			return actionResultMsg{err: fmt.Errorf("update WI #%d state: %w", wiID, err)}
		}
		return actionResultMsg{msg: fmt.Sprintf("WI #%d: %s → %s", wiID, currentState, next)}
	}
}

// --- Pipeline Actions ---

func (m *Model) cancelBuild(build *azdo.Build) tea.Cmd {
	client := m.client
	buildID := build.ID

	return func() tea.Msg {
		err := client.CancelBuild("", "", buildID)
		if err != nil {
			return actionResultMsg{err: fmt.Errorf("cancel build #%d: %w", buildID, err)}
		}
		return actionResultMsg{msg: fmt.Sprintf("Cancelled build #%d", buildID)}
	}
}

func (m *Model) viewBuildLogs(build *azdo.Build) tea.Cmd {
	client := m.client
	buildID := build.ID

	return tea.ExecProcess(
		utils.BuildLogCommand(client, buildID),
		func(err error) tea.Msg {
			if err != nil {
				return actionResultMsg{err: fmt.Errorf("logs: %w", err)}
			}
			return actionResultMsg{msg: "Logs closed"}
		},
	)
}

func (m *Model) selectedURL() string {
	switch m.activeView {
	case ViewPRs:
		if pr := m.prView.SelectedPR(); pr != nil {
			return pr.WebURL()
		}
	case ViewWorkItems:
		if wi := m.wiView.SelectedWorkItem(); wi != nil {
			return wi.WebURL()
		}
	case ViewPipelines:
		if b := m.pipeView.SelectedBuild(); b != nil {
			return b.WebURL()
		}
	}
	return ""
}

// --- View rendering ---

func (m Model) View() string {
	if m.bootErr != nil {
		return m.renderError(m.bootErr)
	}
	if !m.ready {
		return m.renderLoading("Connecting to Azure DevOps...")
	}

	if m.showHelp {
		return m.renderHelp()
	}

	var b strings.Builder
	b.WriteString(m.renderTabs())
	b.WriteString("\n")

	switch m.activeView {
	case ViewPRs:
		b.WriteString(m.prView.View())
	case ViewWorkItems:
		b.WriteString(m.wiView.View())
	case ViewPipelines:
		b.WriteString(m.pipeView.View())
	}

	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	return b.String()
}

func (m *Model) renderTabs() string {
	tabs := []struct {
		name string
		key  string
		view ViewType
	}{
		{"Pull Requests", "1", ViewPRs},
		{"Work Items", "2", ViewWorkItems},
		{"Pipelines", "3", ViewPipelines},
	}

	var rendered []string
	for _, t := range tabs {
		label := fmt.Sprintf("[%s] %s", t.key, t.name)
		if t.view == m.activeView {
			rendered = append(rendered, m.theme.ActiveTab.Render(label))
		} else {
			rendered = append(rendered, m.theme.InactiveTab.Render(label))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (m *Model) renderStatusBar() string {
	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0e0e0")).
		Background(lipgloss.Color("#1c2128"))

	left := barStyle.Render(fmt.Sprintf(" %s/%s ", m.cfg.Organization, m.cfg.Project))

	middle := ""
	if m.statusMsg != "" {
		middle = barStyle.Render("  " + m.statusMsg + " ")
	}

	// Contextual hints
	var hints string
	switch m.activeView {
	case ViewPRs:
		hints = "v approve  m merge  d diff  "
	case ViewWorkItems:
		hints = "a assign  S state  "
	case ViewPipelines:
		hints = "x cancel  l logs  "
	}
	right := barStyle.Render(hints + "? help  q quit ")

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(middle) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return left + barStyle.Width(gap).Render(strings.Repeat(" ", gap)) + middle + right
}

func (m *Model) renderHelp() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff"))
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#58a6ff"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	entry := func(k, d string) string {
		return fmt.Sprintf("  %s %s", keyStyle.Width(14).Render(k), descStyle.Render(d))
	}

	var lines []string
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("  az-dash — Keyboard Shortcuts"))
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("  Navigation"))
	lines = append(lines, entry("j / k", "Move down / up"))
	lines = append(lines, entry("g / G", "First / last item"))
	lines = append(lines, entry("Ctrl+d / u", "Page down / up"))
	lines = append(lines, entry("Tab / S-Tab", "Next / prev section"))
	lines = append(lines, entry("1 / 2 / 3", "PRs / Work Items / Pipelines"))
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("  General"))
	lines = append(lines, entry("p", "Toggle preview pane"))
	lines = append(lines, entry("r / R", "Refresh section / all"))
	lines = append(lines, entry("o", "Open in browser"))
	lines = append(lines, entry("y", "Copy URL"))
	lines = append(lines, entry("#", "Copy ID"))
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("  Pull Requests"))
	lines = append(lines, entry("v", "Approve"))
	lines = append(lines, entry("m", "Complete (merge)"))
	lines = append(lines, entry("x / X", "Abandon / reactivate"))
	lines = append(lines, entry("d", "View diff"))
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("  Work Items"))
	lines = append(lines, entry("a / A", "Assign to me / unassign"))
	lines = append(lines, entry("S", "Cycle state"))
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("  Pipelines"))
	lines = append(lines, entry("x", "Cancel build"))
	lines = append(lines, entry("l", "View logs"))
	lines = append(lines, "")
	lines = append(lines, descStyle.Render("  Press ? or Esc to close"))

	return strings.Join(lines, "\n")
}

func (m *Model) renderLoading(msg string) string {
	return fmt.Sprintf("\n  %s %s\n", m.spinner.View(), msg)
}

func (m *Model) renderError(err error) string {
	return fmt.Sprintf("\n  %s\n\n  Press q to quit.\n",
		m.theme.Error.Render(fmt.Sprintf("Error: %v", err)))
}

func (m *Model) contentWidth() int  { return m.width }
func (m *Model) contentHeight() int { return m.height - 3 }

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

// --- Data fetching (returns messages, never mutates state) ---

func (m *Model) fetchAllSections() []tea.Cmd {
	var cmds []tea.Cmd

	for i, section := range m.cfg.PRSections {
		cmds = append(cmds, m.fetchPRSection(i, section))
	}
	for i, section := range m.cfg.WorkItemSections {
		cmds = append(cmds, m.fetchWISection(i, section))
	}
	for i, section := range m.cfg.PipelineSections {
		cmds = append(cmds, m.fetchBuildSection(i, section))
	}

	return cmds
}

func (m *Model) fetchCurrentSection() tea.Cmd {
	switch m.activeView {
	case ViewPRs:
		idx := m.prView.ActiveSectionIndex()
		if idx < len(m.cfg.PRSections) {
			return m.fetchPRSection(idx, m.cfg.PRSections[idx])
		}
	case ViewWorkItems:
		idx := m.wiView.ActiveSectionIndex()
		if idx < len(m.cfg.WorkItemSections) {
			return m.fetchWISection(idx, m.cfg.WorkItemSections[idx])
		}
	case ViewPipelines:
		idx := m.pipeView.ActiveSectionIndex()
		if idx < len(m.cfg.PipelineSections) {
			return m.fetchBuildSection(idx, m.cfg.PipelineSections[idx])
		}
	}
	return nil
}

func (m *Model) fetchPRSection(i int, section config.PRSection) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		criteria := azdo.PRSearchCriteria{
			Status:     section.Filters.Status,
			CreatorID:  section.Filters.CreatorID,
			ReviewerID: section.Filters.ReviewerID,
			Repository: section.Filters.Repository,
			TargetRef:  section.Filters.TargetBranch,
			SourceRef:  section.Filters.SourceRefName,
		}
		prs, err := client.ListPullRequests(section.Organization, section.Project, criteria, section.Limit)
		return prsFetchedMsg{section: i, data: prs, err: err}
	}
}

func (m *Model) fetchWISection(i int, section config.WorkItemSection) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		items, err := client.QueryWorkItems(section.Organization, section.Project, section.WIQL, section.Limit)
		return workItemsFetchedMsg{section: i, data: items, err: err}
	}
}

func (m *Model) scheduleRefresh() tea.Cmd {
	interval := time.Duration(m.cfg.Defaults.RefetchIntervalMinutes) * time.Minute
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return refreshTickMsg{}
	})
}

func (m *Model) fetchBuildSection(i int, section config.PipelineSection) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		criteria := azdo.BuildSearchCriteria{
			DefinitionID: section.Filters.DefinitionID,
			RequestedFor: section.Filters.RequestedFor,
			StatusFilter: section.Filters.StatusFilter,
			ResultFilter: section.Filters.ResultFilter,
			BranchName:   section.Filters.BranchName,
		}
		builds, err := client.ListBuilds(section.Organization, section.Project, criteria, section.Limit)
		return buildsFetchedMsg{section: i, data: builds, err: err}
	}
}
