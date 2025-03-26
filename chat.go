package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n\n"

type chat struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	userStyle   lipgloss.Style
	senderStyle lipgloss.Style
	err         error
}

// Init

func initialChat() chat {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to the chat area, have a good stay`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return chat{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		userStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
	}
}

func (c chat) Init() tea.Cmd {
	return textarea.Blink
}

// Update

func (c chat) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

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
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(c.textarea.Value())
			return c, tea.Quit
		case tea.KeyEnter:
			c.messages = append(c.messages, c.userStyle.Render("You: ")+(c.textarea.Value()))
			request := request{
				Model: "llama3.2",
				Messages: []messages{
					{
						Role:    "user",
						Content: strings.Join(c.messages, ","),
					},
				},
				Stream: false,
			}

			resp, err := sendingMessage(request)
			if err != nil {
				panic(err)
			} else if len(resp.Message.Content) > 0 {

				c.messages = append(c.messages, c.senderStyle.Render("Llama: ")+(resp.Message.Content))
			} else {
				c.messages = append(c.messages, c.senderStyle.Render("No response received"))
			}

			c.viewport.SetContent(lipgloss.NewStyle().Width(c.viewport.Width).Render(strings.Join(c.messages, "\n")))
			c.textarea.Reset()
			c.viewport.GotoBottom()
		}

	// handling errors
	case errMsg:
		c.err = msg
		return c, nil
	}

	return c, tea.Batch(tiCmd, vpCmd)
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
