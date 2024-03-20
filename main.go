package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
    cloud "github.com/terramate-io/terramate/cloud" // Import the terramate/cloud package)
)

type CredentialData struct {
	IDToken string `json:"id_token"`
}

func LoadCredentials(filePath string) (string, error) {
	var creds CredentialData
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", err
	}
	return creds.IDToken, nil
}

type SimpleTokenCredential struct {
	token string
}

func (c *SimpleTokenCredential) Token() (string, error) {
	return c.token, nil
}



type errMsg error
type userMsg struct {
	User cloud.User
}

type model struct {
	spinner  spinner.Model
	quitting bool
	err      error
	user     cloud.User
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{spinner: s}
}

func fetchUser(token string) tea.Msg {
	client := cloud.Client{
		BaseURL:    cloud.BaseURL,
		Credential: &SimpleTokenCredential{token: token},
		HTTPClient: &http.Client{},
		Logger:     &zerolog.Logger{},
	}

	user, err := client.Users(context.Background())
	if err != nil {
		return errMsg(fmt.Errorf("failed to fetch user: %w", err))
	}

	return userMsg{User: user}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(func() tea.Msg {
		token, err := LoadCredentials("/var/home/rene/.terramate.d/credentials.tmrc.json")
		if err != nil {
			return errMsg(fmt.Errorf("error loading credentials: %w", err))
		}

		return fetchUser(token)
	}, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}
	case errMsg:
		m.err = msg
		return m, nil
	case userMsg:
		m.user = cloud.User(msg.User) // Ensure your User type aligns with what is returned
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\n%s", m.err, quitKeys.Help().Desc)
	}
	if m.user.Email != "" {
		userInfo := fmt.Sprintf("Email: %s\nDisplay Name: %s\nJob Title: %s\nUUID: %s", m.user.Email, m.user.DisplayName, m.user.JobTitle, m.user.UUID)
		return fmt.Sprintf("\nUser Information:\n%s\n\n%s", userInfo, quitKeys.Help().Desc)
	}
	str := fmt.Sprintf("\n\n   %s Loading user information... %s\n\n", m.spinner.View(), quitKeys.Help().Desc)
	if m.quitting {
		return str + "\n"
	}
	return str
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
