package src

import (
	"fmt"
	"math"
	"strings"

	"starminal/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type starsLoadedMsg struct {
	stars    []utils.VisibleStar
	location utils.Location
	err      error
}

type screenshotMsg struct {
	path string
	err  error
}

type SkyModel struct {
	catalog    utils.StarCatalog
	stars      []utils.VisibleStar
	location   utils.Location
	width      int
	height     int
	pincode    string
	input      string
	loaded     bool
	err        string
	screenshot string
}

func NewSkyModel(catalog utils.StarCatalog) SkyModel {
	return SkyModel{
		catalog: catalog,
		width:   120,
		height:  40,
	}
}

func (m SkyModel) Init() tea.Cmd {
	return nil
}

func (m SkyModel) loadStars() tea.Msg {
	loc, err := utils.GeolocatePin(m.pincode)
	if err != nil {
		return starsLoadedMsg{err: err}
	}
	stars := utils.GetVisibleStars(m.catalog, loc.Lat, loc.Lon)
	return starsLoadedMsg{stars: stars, location: loc}
}

func (m SkyModel) takeScreenshot() tea.Msg {
	path, err := utils.SaveScreenshot(m.stars, m.width, m.height, m.pincode)
	if err != nil {
		return screenshotMsg{err: err}
	}
	return screenshotMsg{path: path}
}

func (m SkyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.loaded {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				if m.input != "" {
					m.pincode = m.input
					m.err = ""
					return m, m.loadStars
				}
			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.input += msg.String()
				}
			}
		} else {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.loaded = false
				m.input = ""
				m.stars = nil
				m.screenshot = ""
			case "s":
				m.screenshot = ""
				return m, m.takeScreenshot
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height - 2
	case screenshotMsg:
		if msg.err != nil {
			m.screenshot = "Error: " + msg.err.Error()
		} else {
			m.screenshot = "Saved: " + msg.path
		}
	case starsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.stars = msg.stars
			m.location = msg.location
			m.loaded = true
		}
	}
	return m, nil
}

func (m SkyModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	if !m.loaded {
		return m.viewInput()
	}
	return m.viewSky()
}

func (m SkyModel) viewInput() string {
	var sb strings.Builder

	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	title := lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Bold(true)
	cursor := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	padTop := m.height / 3
	for range padTop {
		sb.WriteRune('\n')
	}

	line := title.Render("  Starminal")
	sb.WriteString(line + "\n\n")

	sb.WriteString(dim.Render("  Enter pincode: "))
	sb.WriteString(cursor.Render(m.input))
	sb.WriteString(cursor.Render("█"))
	sb.WriteString("\n\n")

	if m.err != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		sb.WriteString(errStyle.Render("  Error: " + m.err))
		sb.WriteString("\n")
	}

	sb.WriteString(dim.Render("  Press Enter to view sky"))

	return sb.String()
}

func (m SkyModel) viewSky() string {
	type cell struct {
		char  rune
		style lipgloss.Style
	}

	skyW := m.width - 2
	skyH := m.height

	grid := make([][]cell, skyH)
	for y := range grid {
		grid[y] = make([]cell, skyW)
		for x := range grid[y] {
			grid[y][x] = cell{char: ' '}
		}
	}

	centerX := float64(skyW) / 2.0
	centerY := float64(skyH) / 2.0
	radius := math.Min(centerX, centerY)

	for _, s := range m.stars {
		altRad := s.Alt * math.Pi / 180.0
		azRad := s.Az * math.Pi / 180.0

		r := math.Cos(altRad) / (1.0 + math.Sin(altRad))
		px := r * math.Sin(azRad)
		py := -r * math.Cos(azRad)

		x := int(centerX + px*radius)
		y := int(centerY + py*radius)

		if x < 0 || x >= skyW || y < 0 || y >= skyH {
			continue
		}

		ch := starChar(s.Mag)
		style := starStyle(s.Ci)

		existing := grid[y][x]
		if existing.char == ' ' || ch == '*' || (ch == '·' && existing.char == '.') {
			grid[y][x] = cell{char: ch, style: style}
		}
	}

	info := m.buildInfoPanel()
	infoLines := strings.Split(info, "\n")
	infoWidth := 0
	for _, l := range infoLines {
		w := lipgloss.Width(l)
		if w > infoWidth {
			infoWidth = w
		}
	}

	infoStartX := skyW - infoWidth - 1
	if infoStartX < 0 {
		infoStartX = 0
	}

	var sb strings.Builder
	border := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	sb.WriteString(border.Render("┌" + strings.Repeat("─", skyW) + "┐"))
	sb.WriteRune('\n')

	for rowIdx, row := range grid {
		sb.WriteString(border.Render("│"))

		var infoLine string
		if rowIdx < len(infoLines) {
			infoLine = infoLines[rowIdx]
		}

		if infoLine != "" {
			for x := 0; x < infoStartX; x++ {
				c := row[x]
				if c.char == ' ' {
					sb.WriteRune(' ')
				} else {
					sb.WriteString(c.style.Render(string(c.char)))
				}
			}
			sb.WriteString(infoLine)
			remaining := skyW - infoStartX - lipgloss.Width(infoLine)
			if remaining > 0 {
				sb.WriteString(strings.Repeat(" ", remaining))
			}
		} else {
			for _, c := range row {
				if c.char == ' ' {
					sb.WriteRune(' ')
				} else {
					sb.WriteString(c.style.Render(string(c.char)))
				}
			}
		}

		sb.WriteString(border.Render("│"))
		sb.WriteRune('\n')
	}

	sb.WriteString(border.Render("└" + strings.Repeat("─", skyW) + "┘"))
	sb.WriteRune('\n')

	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	status := " [s] screenshot  [esc] new location  [q] quit"
	if m.screenshot != "" {
		status += "  | " + m.screenshot
	}
	sb.WriteString(statusStyle.Render(status))

	return sb.String()
}

func (m SkyModel) buildInfoPanel() string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	bright := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("75"))

	var lines []string
	lines = append(lines, label.Render("  Location "))
	lines = append(lines, dim.Render("  Country: ")+bright.Render(m.location.Country))
	lines = append(lines, dim.Render("  Pin:     ")+bright.Render(m.pincode))
	lines = append(lines, dim.Render("  Lat:     ")+bright.Render(fmt.Sprintf("%.4f°", m.location.Lat)))
	lines = append(lines, dim.Render("  Lon:     ")+bright.Render(fmt.Sprintf("%.4f°", m.location.Lon)))
	lines = append(lines, "")
	lines = append(lines, label.Render("  Sky "))
	lines = append(lines, dim.Render("  Stars: ")+bright.Render(fmt.Sprintf("%d", len(m.stars))))

	return strings.Join(lines, "\n")
}

func starChar(mag float64) rune {
	switch {
	case mag < 3.0:
		return '*'
	case mag < 5.0:
		return '·'
	default:
		return '.'
	}
}

func starStyle(ci float64) lipgloss.Style {
	var color string
	switch {
	case math.IsNaN(ci) || ci == 0:
		color = "255"
	case ci < -0.1:
		color = "75"
	case ci < 0.3:
		color = "153"
	case ci < 0.6:
		color = "255"
	case ci < 1.0:
		color = "229"
	case ci < 1.5:
		color = "215"
	default:
		color = "196"
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color))
}
