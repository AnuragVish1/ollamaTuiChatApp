package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n\n"

// Add a new message type for API response
type apiResponseMsg struct {
	content string
	err     error
}

// Implement a command to fetch the API response asynchronously
func fetchAPIResponse(message []string) tea.Cmd {
	return func() tea.Msg {
		request := request{
			Model: "llama3.2",
			Messages: []messages{
				{
					Role:    "user",
					Content: strings.Join(message, ","),
				},
			},
			Stream: false,
		}

		resp, err := sendingMessage(request)
		if err != nil {
			return apiResponseMsg{content: "", err: err}
		} else if len(resp.Message.Content) > 0 {
			return apiResponseMsg{content: resp.Message.Content, err: nil}
		}
		return apiResponseMsg{content: "No response received", err: nil}
	}
}

// Add a tick message for the spinner
type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Message for updating the loading indicator
type updateLoadingMsg struct{}

type chat struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	userStyle   lipgloss.Style
	senderStyle lipgloss.Style
	err         error
	// Add spinner for loading animation
	spinner      spinner.Model
	isLoading    bool
	loadingLabel string
	// Track the loading message index so we can update it
	loadingMsgIndex int
}

// Init

func initialChat() chat {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "> "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(1)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to the chat area, have a good stay`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))

	return chat{
		textarea:        ta,
		messages:        []string{},
		viewport:        vp,
		userStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		senderStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:             nil,
		spinner:         s,
		isLoading:       false,
		loadingLabel:    "Llama is thinking",
		loadingMsgIndex: -1,
	}
}

func (c chat) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, c.spinner.Tick)
}

// Update

func (c chat) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		spCmd tea.Cmd
	)

	// Always update spinner to keep the animation working
	c.spinner, spCmd = c.spinner.Update(msg)
	c.textarea, tiCmd = c.textarea.Update(msg)
	c.viewport, vpCmd = c.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.viewport.Width = msg.Width
		c.textarea.SetWidth(msg.Width)
		c.viewport.Height = msg.Height - c.textarea.Height() - lipgloss.Height(gap)

		if len(c.messages) > 0 {
			// Wrap content before setting it.
			c.viewport.SetContent(lipgloss.NewStyle().Width(c.viewport.Width).Render(strings.Join(c.messages, "\n")))
		}
		c.viewport.GotoBottom()

	case tickMsg:
		// Continue ticking if we're still loading
		if c.isLoading {
			// Update the loading message with current spinner state
			return c, tea.Batch(
				tick(),
				func() tea.Msg { return updateLoadingMsg{} },
			)
		}

	case updateLoadingMsg:
		if c.isLoading && c.loadingMsgIndex >= 0 && c.loadingMsgIndex < len(c.messages) {
			// Update the loading message with the current spinner state
			c.messages[c.loadingMsgIndex] = c.senderStyle.Render("Llama: ") + c.spinner.View() + " " + c.loadingLabel
			c.viewport.SetContent(lipgloss.NewStyle().Width(c.viewport.Width).Render(strings.Join(c.messages, "\n")))
			c.viewport.GotoBottom()
		}

	case apiResponseMsg:
		// Handle API response
		c.isLoading = false
		if msg.err != nil {
			c.err = msg.err
			// Replace loading message with error message
			if c.loadingMsgIndex >= 0 && c.loadingMsgIndex < len(c.messages) {
				c.messages[c.loadingMsgIndex] = c.senderStyle.Render("Error: ") + msg.err.Error()
			} else {
				c.messages = append(c.messages, c.senderStyle.Render("Error: ")+msg.err.Error())
			}
		} else {
			// Replace loading message with actual response
			if c.loadingMsgIndex >= 0 && c.loadingMsgIndex < len(c.messages) {
				c.messages[c.loadingMsgIndex] = c.senderStyle.Render("Llama: ") + msg.content
			} else {
				c.messages = append(c.messages, c.senderStyle.Render("Llama: ")+msg.content)
			}
		}
		// Reset loading message index
		c.loadingMsgIndex = -1
		c.viewport.SetContent(lipgloss.NewStyle().Width(c.viewport.Width).Render(strings.Join(c.messages, "\n")))
		c.viewport.GotoBottom()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(c.textarea.Value())
			return c, tea.Quit
		case tea.KeyEnter:
			if c.isLoading {
				// Don't allow new messages while loading
				return c, tea.Batch(tiCmd, vpCmd, spCmd)
			}

			userMsg := c.textarea.Value()
			if strings.TrimSpace(userMsg) == "" {
				return c, tea.Batch(tiCmd, vpCmd)
			}

			// Add user message immediately
			c.messages = append(c.messages, c.userStyle.Render("You: ")+userMsg)

			// Add loading message with spinner animation
			c.messages = append(c.messages, c.senderStyle.Render("Llama: ")+c.spinner.View()+" "+c.loadingLabel)
			c.loadingMsgIndex = len(c.messages) - 1

			// Update viewport
			c.viewport.SetContent(lipgloss.NewStyle().Width(c.viewport.Width).Render(strings.Join(c.messages, "\n")))
			c.textarea.Reset()
			c.viewport.GotoBottom()

			// Set loading state and start spinner
			c.isLoading = true

			// Start the spinner and fetch API response
			return c, tea.Batch(
				tiCmd,
				vpCmd,
				c.spinner.Tick,
				tick(),
				fetchAPIResponse(c.messages[:len(c.messages)-1]), // Don't include the loading message in the API call
			)
		}

	// handling errors
	case errMsg:
		c.err = msg
		return c, nil
	}

	cmds := []tea.Cmd{tiCmd, vpCmd}
	if spCmd != nil {
		cmds = append(cmds, spCmd)
	}

	return c, tea.Batch(cmds...)
}

// View
func (c chat) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		c.viewport.View(),
		gap,
		c.textarea.View(),
	)
}
