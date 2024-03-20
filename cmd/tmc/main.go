	package main

	import (
		"fmt"

		"os"

		config "github.com/RocketRene/tmc/internal/config"   // Import the internal/config package
		service "github.com/RocketRene/tmc/internal/service" // Import the internal/service package
		"github.com/charmbracelet/bubbles/key"
		"github.com/charmbracelet/bubbles/spinner"
		tea "github.com/charmbracelet/bubbletea"
		"github.com/charmbracelet/lipgloss"
		cloud "github.com/terramate-io/terramate/cloud" // Import the terramate/cloud package)
	)

	var (
		docStyle = lipgloss.NewStyle().Margin(1, 2) // Overall document style
		topBar   = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 1).
				MarginBottom(1)
		mainArea = lipgloss.NewStyle().
				MarginTop(1)
		// Define more styles for other areas like a status bar if needed
	)

	type model struct {
		spinner  spinner.Model
		quitting bool
		err      error
		user     cloud.User
		width    int
		height   int
	}

	var quitKeys = key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("", "press q to quit"),
	)

	func initialModel() model {
		s := spinner.New(spinner.WithSpinner(spinner.Dot))
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

		return model{
			spinner: s,
			width:   80, // Default values, will be updated on start
			height:  24,
		}
	}

	func (m model) Init() tea.Cmd {
		return tea.Batch(func() tea.Msg {
			token, err := config.LoadCredentials("/var/home/rene/.terramate.d/credentials.tmrc.json")
			if err != nil {
				return service.ErrMsg(fmt.Errorf("error loading credentials: %w", err))
			}

			user, err := service.FetchUser(token)
			if err != nil {
				return service.ErrMsg(fmt.Errorf("failed to fetch user: %w", err))
			}
			return service.UserMsg{User: user}
		}, m.spinner.Tick)
	}

	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
		switch msg := msg.(type) {

		case tea.WindowSizeMsg:
			// Handle terminal resize events
			m.width, m.height = msg.Width, msg.Height
			return m, nil

		case tea.KeyMsg:
			// Handle key press events
			switch msg.String() {
			case "q", "esc", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			}

		case service.ErrMsg:
			// Handle error messages
			m.err = msg
			return m, nil

		case service.UserMsg:
			// Correctly update the model with user data
			m.user = msg.User
			return m, nil
		

		default:
			// Handle spinner update and other messages
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

		return m, nil
	}

	func (m model) View() string {
		var view string

		// Define dimensions for layout components based on the terminal size.
		// You might want to dynamically adjust sizes or styles based on `m.width` and `m.height`.
		topBar := lipgloss.NewStyle().
			Foreground(lipgloss.Color("228")).
			Background(lipgloss.Color("63")).
			Padding(0, 1).
			MarginBottom(1).
			Width(m.width). // Make the top bar span the entire width
			// Your styling
			Render(fmt.Sprintf("User: %s | Display Name: %s | Job Title: %s", m.user.Email, m.user.DisplayName, m.user.JobTitle))
		
		// Placeholder for the main content area
		mainContent := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height - 3). // Adjust the height based on the top bar and potential status bar
			Render("Main content area")

		// Combine the components into the full view
		view = lipgloss.JoinVertical(lipgloss.Top, topBar, mainContent)

		// In case of an error, prepend the error message to the view
		if m.err != nil {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")).
				Bold(true).
				Render(fmt.Sprintf("\nError: %v\n", m.err))
			view = lipgloss.JoinVertical(lipgloss.Top, errorStyle, view)
		}

		// If quitting, add a quitting message or handle as needed
		if m.quitting {
			quittingStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("11")).
				Render("\nQuitting...\n")
			view = lipgloss.JoinVertical(lipgloss.Top, view, quittingStyle)
		}

		// Add spinner view if still loading data
		if m.user.Email == "" && !m.quitting {
			spinnerView := m.spinner.View()
			loadingStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Render(fmt.Sprintf("\n\n%s Loading user information... Press 'q' to quit.\n\n", spinnerView))
			view = lipgloss.JoinVertical(lipgloss.Top, view, loadingStyle)
		}

		return view
	}

	func main() {
		p := tea.NewProgram(initialModel())
		if _, err := p.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
