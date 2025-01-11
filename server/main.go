package main

import (
	_ "embed"
	"fmt"
	"path/filepath"
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
		fyne.NewMenu("File", fyne.NewMenuItem("New Task", func() {
			title_entry := widget.NewEntry()
			command := widget.NewMultiLineEntry()

			items := []*widget.FormItem{
				widget.NewFormItem("Title", title_entry),
				widget.NewFormItem("Command", command),
			}

			form := dialog.NewForm("Edit Task", "Confirm", "Cancel", items,
				func(confirmed bool) {
					if confirmed {
						// _, err := createTask(title_entry.Text, command.Text)
						// if err != nil {
						// 	// TODO: Notify user?
						// 	fmt.Println("Failed to update custom task", err.Error())
						// }

						refreshGui(1)
					}
				}, window)
			form.Resize(fyne.NewSize(400, 200))
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

	listmap := binding.NewStringList()
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
			script, _ := dataItem.(binding.String).Get()
			objects := canvasObject.(*fyne.Container).Objects
			script_noext := strings.TrimSuffix(script, filepath.Ext(script))

			objects[0].(*widget.Label).SetText(script_noext)

			// edit button
			editBtn := objects[2].(*fyne.Container).Objects[0].(*widget.Button)
			editBtn.OnTapped = func() {
				task_data, err := readScript(script)
				if err != nil {
					// TODO: Notify user
					return
				}

				filename_entry := widget.NewEntry()
				command_entry := widget.NewMultiLineEntry()

				filename_entry.SetText(script)
				command_entry.SetText(task_data)

				items := []*widget.FormItem{
					widget.NewFormItem("Title", filename_entry),
					widget.NewFormItem("Command", command_entry),
				}

				form := dialog.NewForm("Edit Task", "Confirm", "Cancel", items,
					func(confirmed bool) {
						if confirmed {
							err := writeScript(script, command_entry.Text)
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
