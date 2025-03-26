package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	program := tea.NewProgram(
		newGreetingPage("This is the first cli with go"),
	)
	model, err := program.Run()

	if err != nil {
		panic(err)
	}
	model.Init()

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
	chatProgram := tea.NewProgram(initialChat())
	chats, err := chatProgram.Run()
	if err != nil {
		print("Error hai yaha")
		panic(err)
	}
	chats.Init()

}
