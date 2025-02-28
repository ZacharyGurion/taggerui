package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"math"

	tea "github.com/charmbracelet/bubbletea"
	
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"github.com/h2non/filetype"

	//"github.com/dhowden/tag"
	//"github.com/KristoforMaynard/music-tag"

	taglib "go.senan.xyz/taglib"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type tag struct {
	enabled		bool
	name		string
	taglibName	string
	width		float64
}

var songFiles []string

var show = []tag{
	{name: "File", width: 0.20, enabled: true, taglibName: "none"},
	{name: "Title", width: 0.30, enabled: true, taglibName: taglib.AlbumArtist},
	{name: "Album", width: 0.20, enabled: true, taglibName: taglib.Album},
	{name: "DiscNumber", width: 0.20, enabled: false, taglibName: taglib.DiscNumber},
}

type model struct {
	table			table.Model
	height			int
	width			int
	tags			[][]string
}

func isAudioFile(filePath string) (bool, error) {
	buf := make([]byte, 261)

	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	_, err = file.Read(buf)
	if err != nil {
		return false, err
	}

	kind, _ := filetype.Match(buf)
    if kind.MIME.Type == "audio"{
        return true, nil
    }

	return false, nil
}


func scanFiles(dir string) ([]string) {
	root := os.DirFS(dir)
	dirFiles, err := fs.Glob(root, "*")

	if err != nil {
		log.Fatal(err)
	}

	var fileNames	[]string

	for _, v := range dirFiles {
		isAudio, err := isAudioFile(v)
		if err != nil {
			fmt.Println("Error:", err)
		}
		if isAudio {
			fileNames = append(fileNames, v)
		}
	}
	return fileNames
}

func initialModel() model {
	m := model{
	}
	return m
}
func setTable(m model) model{
	cols := []table.Column{}
	total := 0.0
	for _, t := range show{
		if t.enabled{
			total += t.width
		}
	}
	for _, t := range show{
		if t.enabled{
			cols = append(cols, table.Column{Title: t.name, Width: int(math.Round(t.width*float64(m.width-10)/total))})
		}
	}

	var rows []table.Row
	for _, f := range(songFiles) {
		rs := []string{}
		tags, _ := taglib.ReadTags(f)

		for _, t := range show {
			if (t.enabled && t.name == "File") {
				rs = append(rs, f)
			} else if (t.enabled && t.taglibName != "none") {
				if (len(tags[t.taglibName]) == 1) {
					rs = append(rs, tags[t.taglibName][0])
				}
			}
		}
		rows = append(rows, table.Row(rs))
	}

	t := table.New(
			table.WithColumns(cols),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(m.height-3),
		)
	
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	m.table = t

	return m
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd

    switch msg := msg.(type) {
	
	case tea.WindowSizeMsg:
		tea.ClearScreen()
		m.width = msg.Width
		m.height = msg.Height
		fmt.Print("\033[H\033[2J")

		m = setTable(m)
		
    case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			fmt.Print("\033[H\033[2J")
			return m, tea.Quit
		case "enter":
			tagsOut, _ := taglib.ReadTags(songFiles[m.table.Cursor()])
			return m, tea.Batch(
				tea.Printf("Edit song: %s", tagsOut[taglib.Title]),
			)
		}
    }

	m.table, cmd = m.table.Update(msg)

    return m, cmd
}

func (m model) View() string {
	if (m.width < 20 || m.height < 8) {
		tea.ClearScreen()
		return "Window too small"
	}
	s := ""
	//s += "Songs in directory: \n"

    // The footer
	s += baseStyle.Render(m.table.View())
    //s += "\nPress q to quit.\n"
    // Send the UI for rendering
    return s
}

func main() {
	
	//argsWithProg := os.Args
	//argsWithoutProg := os.Args[1:]

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(pwd)
	//fmt.Printf("pwd = %T\n", pwd) 

	songFiles = scanFiles(pwd)

    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Alas, there's been an error: %v", err)
        os.Exit(1)
    }
}
