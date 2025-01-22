package main

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed Icon.png
var logo []byte

var (
	window          fyne.Window
	fyneApp         fyne.App
	scripts_tab     *container.TabItem
	preferences_tab *container.TabItem
	tabs            *container.AppTabs
	preferences     fyne.Preferences
)

func main() {
	fyneApp = app.NewWithID("dev.ibanks.opendesk-server")
	window = fyneApp.NewWindow("OpenDesk Server")
	preferences = fyneApp.Preferences()

	fyneApp.Settings().SetTheme(theme.DarkTheme())

	window.Resize(fyne.Size{Width: 600, Height: 400})

	StartServer()

	buildGui()

	window.SetCloseIntercept(func() {
		window.Hide()
	})

	if desk, ok := fyneApp.(desktop.App); ok {
		menu := fyne.NewMenu("OpenDesk Server", fyne.NewMenuItem("Show", func() {
			window.Show()
		}))

		desk.SetSystemTrayMenu(menu)
		desk.SetSystemTrayIcon(fyne.NewStaticResource("logo", logo))
	}

	if preferences.Bool("minimized") {
		fyneApp.Run()
	} else {
		window.ShowAndRun()
	}
}

func buildGui() {
	scripts_tab = container.NewTabItem("Builtin Tasks", container.NewVBox())
	preferences_tab = container.NewTabItem("Settings", container.NewVBox())

	tabs = container.NewAppTabs(scripts_tab, preferences_tab)
	tabs.SetTabLocation(container.TabLocationLeading)

	tabs.OnSelected = func(tab *container.TabItem) {
		refreshGui(tabs.SelectedIndex())
	}

	menu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Refresh", buildGui),
			fyne.NewMenuItem("New Task", func() {
				id_entry := widget.NewEntry()
				title_entry := widget.NewEntry()
				command := widget.NewMultiLineEntry()

				items := []*widget.FormItem{
					widget.NewFormItem("ID", id_entry),
					widget.NewFormItem("Name", title_entry),
					widget.NewFormItem("Script", command),
				}

				form := dialog.NewForm("New Task", "Confirm", "Cancel", items,
					func(confirmed bool) {
						if confirmed {
							id, err := strconv.Atoi(id_entry.Text)
							if err != nil {
								fmt.Println("Failed to create script: ID is not a number")
								return
							}
							err = writeScript(id, title_entry.Text+".ts", command.Text)
							if err != nil {
								// TODO: Notify user?
								fmt.Println("Failed to create script", err.Error())
							}

							refreshGui(0)
						}
					}, window)
				form.Resize(fyne.NewSize(400, 300))
				form.Show()
			})))
	window.SetMainMenu(menu)

	refreshGui(0)

	window.SetContent(tabs)
}

func refreshGui(tabIndex int) {
	buildScriptsTab()
	buildPrefencesTab()

	tabs.SelectIndex(tabIndex)
}

func buildScriptsTab() {
	scripts := getScripts()
	sort.Slice(scripts, func(i, j int) bool {
		return scripts[i].ID < scripts[j].ID
	})

	listmap := binding.NewUntypedList()
	for _, script := range scripts {
		listmap.Append(script)
	}

	list := widget.NewListWithData(listmap,
		// Create item
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				layout.NewSpacer(),
				container.NewGridWithColumns(2,
					widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {}),
					widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {}),
				))
		},
		// Update item
		func(dataItem binding.DataItem, canvasObject fyne.CanvasObject) {
			untyped, _ := dataItem.(binding.Untyped).Get()
			script := untyped.(Script)
			objects := canvasObject.(*fyne.Container).Objects
			script_noext := strings.TrimSuffix(script.File, filepath.Ext(script.File))

			objects[0].(*widget.Label).SetText(script_noext)

			// edit button
			editBtn := objects[2].(*fyne.Container).Objects[0].(*widget.Button)
			editBtn.OnTapped = func() {
				task_data, err := readScript(script.File)
				if err != nil {
					// TODO: Notify user
					return
				}

				id_entry := widget.NewEntry()
				filename_entry := widget.NewEntry()
				command_entry := widget.NewMultiLineEntry()

				id_entry.SetText(strconv.Itoa(script.ID))
				filename_entry.SetText(script.File)
				command_entry.SetText(task_data)

				items := []*widget.FormItem{
					widget.NewFormItem("ID", id_entry),
					widget.NewFormItem("Name", filename_entry),
					widget.NewFormItem("Script", command_entry),
				}

				form := dialog.NewForm("Edit Task", "Confirm", "Cancel", items,
					func(confirmed bool) {
						if confirmed {
							id, err := strconv.Atoi(id_entry.Text)
							if err != nil {
								fmt.Println("Failed to update task: ID not a number")
								return
							}
							err = updateScript(script, id, command_entry.Text)
							if err != nil {
								// TODO: Notify user?
								fmt.Println("Failed to update custom task", err.Error())
							}
							refreshGui(0)
						}
					}, window)
				form.Resize(fyne.NewSize(400, 200))
				form.Show()
			}

			deleteBtn := objects[2].(*fyne.Container).Objects[1].(*widget.Button)
			deleteBtn.OnTapped = func() {
				// deleteCustomTask(task.ID)
				refreshGui(0)
			}
		})

	scroll := container.NewScroll(list)
	padded := layout.NewCustomPaddedLayout(0, 0, 16, 0)
	tab := container.New(padded, scroll)

	scripts_tab.Content = tab
}

func buildPrefencesTab() {
	minimized := preferences.Bool("minimized")
	port := preferences.StringWithFallback("port", "9212")

	minimizedBinding := binding.BindBool(&minimized)
	portBinding := binding.BindString(&port)

	minimizedCheck := widget.NewCheckWithData("", minimizedBinding)
	portInput := widget.NewEntryWithData(portBinding)

	form := widget.NewForm(widget.NewFormItem("Port", portInput),
		widget.NewFormItem("Start Minimized", minimizedCheck))
	form.OnSubmit = func() {
		preferences.SetBool("minimized", minimizedCheck.Checked)
		preferences.SetString("port", portInput.Text)

		StartServer()
	}

	container := container.NewVBox(form)

	preferences_tab.Content = container
}

func getFallbackContainer(tabIndex int, text string) *fyne.Container {
	label := widget.NewLabel(text)
	button := widget.NewButtonWithIcon("Reload", theme.ViewRefreshIcon(), func() {
		refreshGui(tabIndex)
	})

	vbox := container.NewVBox(label, button)
	center := container.NewCenter(vbox)

	return center
}
