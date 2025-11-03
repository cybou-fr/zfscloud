package gui

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// FileItem represents a file in the backup
type FileItem struct {
	Path         string
	Size         int64
	ModifiedTime time.Time
	IsDirectory  bool
	Hash         []byte
}

// FilesView displays backed up files and restore functionality
type FilesView struct {
	app           *App
	refreshButton widget.Clickable
	restoreButton widget.Clickable

	// File list
	fileList     widget.List
	files        []FileItem
	loading      bool
	errorMsg     string
	fileItems    []widget.Clickable // Clickable for each file
	selectedFile int                // Index of selected file (-1 if none)

	// Restore dialog
	dialogVisible     bool
	restorePath       widget.Editor
	restoreConfirmBtn widget.Clickable
	restoreCancelBtn  widget.Clickable
	restoring         bool
	restoreProgress   string
	restoreErrorMsg   string
}

// NewFilesView creates a new files view
func NewFilesView(app *App) *FilesView {
	restorePathEditor := widget.Editor{
		SingleLine: true,
		Submit:     false,
	}
	restorePathEditor.SetText("C:\\Restored\\")

	return &FilesView{
		app: app,
		fileList: widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		selectedFile: -1, // No selection initially
		restorePath:  restorePathEditor,
	}
}

// Layout renders the files view
func (f *FilesView) Layout(gtx layout.Context) layout.Dimensions {
	// Show restore dialog overlay if visible
	if f.dialogVisible {
		return f.layoutRestoreDialog(gtx)
	}

	return layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// Title and actions
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return f.layoutHeader(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

			// File browser/list
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return f.layoutFileList(gtx)
			}),
		)
	})
}

// layoutHeader renders the header with title and action buttons
func (f *FilesView) layoutHeader(gtx layout.Context) layout.Dimensions {
	// Handle button clicks
	if f.refreshButton.Clicked(gtx) {
		f.refreshFileList()
	}
	if f.restoreButton.Clicked(gtx) {
		f.showRestoreDialog()
	}

	theme := f.app.GetTheme()

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Spacing:   layout.SpaceBetween,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			title := theme.H5("Backup Files")
			return title.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Horizontal,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(theme.Theme, &f.refreshButton, "Refresh")
					btn.Background = theme.Info
					btn.Color = theme.OnPrimary
					return btn.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(theme.Theme, &f.restoreButton, "Restore...")
					btn.Background = theme.Primary
					btn.Color = theme.OnPrimary
					return btn.Layout(gtx)
				}),
			)
		}),
	)
}

// layoutFileList renders the list of backed up files
func (f *FilesView) layoutFileList(gtx layout.Context) layout.Dimensions {
	theme := f.app.GetTheme()

	// Show loading indicator
	if f.loading {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return theme.Body1("Loading files...").Layout(gtx)
		})
	}

	// Show error message
	if f.errorMsg != "" {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Vertical,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					text := theme.Body1("Error:")
					text.Color = theme.Error
					return text.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					text := theme.Body2(f.errorMsg)
					return text.Layout(gtx)
				}),
			)
		})
	}

	// Show empty state
	if len(f.files) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Vertical,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					text := theme.H6("No files to display")
					return text.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					text := theme.Body1("Start a backup to see files here")
					return text.Layout(gtx)
				}),
			)
		})
	}

	// Render scrollable file list
	return material.List(theme.Theme, &f.fileList).Layout(gtx, len(f.files), func(gtx layout.Context, index int) layout.Dimensions {
		file := f.files[index]

		// Handle file click
		if f.fileItems[index].Clicked(gtx) {
			f.selectedFile = index
			log.Printf("Selected file: %s", file.Path)
		}

		// Determine if this file is selected
		isSelected := f.selectedFile == index

		return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return material.Clickable(gtx, &f.fileItems[index], func(gtx layout.Context) layout.Dimensions {
				return f.layoutFileItem(gtx, file, isSelected)
			})
		})
	})
}

// layoutFileItem renders a single file item
func (f *FilesView) layoutFileItem(gtx layout.Context, file FileItem, isSelected bool) layout.Dimensions {
	theme := f.app.GetTheme()

	// Format file size
	sizeStr := formatBytes(file.Size)

	// Format modified time
	timeStr := file.ModifiedTime.Format("2006-01-02 15:04:05")

	return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// File path
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := file.Path
				if file.IsDirectory {
					text = file.Path + "/"
				}
				label := theme.Body1(text)
				if isSelected {
					label.Font.Weight = 700 // Bold weight
					label.Color = theme.Primary
				}
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
			// Size and modified time
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Horizontal,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := theme.Caption(fmt.Sprintf("Size: %s", sizeStr))
						label.Color = theme.TextSecondary
						return label.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := theme.Caption(fmt.Sprintf("Modified: %s", timeStr))
						label.Color = theme.TextSecondary
						return label.Layout(gtx)
					}),
				)
			}),
		)
	})
}

// refreshFileList fetches the latest file list from server
func (f *FilesView) refreshFileList() {
	log.Println("Refreshing file list...")
	f.loading = true
	f.errorMsg = ""
	f.app.window.Invalidate()

	// Call ListFiles RPC in background
	go func() {
		if f.app.client == nil {
			f.errorMsg = "Client not initialized"
			f.loading = false
			f.app.window.Invalidate()
			return
		}

		// Call ListFiles using the client wrapper
		fileInfos, err := f.app.client.ListFiles("/", true)
		if err != nil {
			f.errorMsg = fmt.Sprintf("Failed to list files: %v", err)
			log.Printf("ListFiles error: %v", err)
			f.loading = false
			f.app.window.Invalidate()
			return
		}

		// Convert to FileItem
		files := make([]FileItem, len(fileInfos))
		for i, file := range fileInfos {
			modTime := time.Unix(0, 0)
			if file.ModifiedTime != nil {
				modTime = file.ModifiedTime.AsTime()
			}
			files[i] = FileItem{
				Path:         file.Path,
				Size:         file.Size,
				ModifiedTime: modTime,
				IsDirectory:  file.IsDirectory,
				Hash:         file.Hash,
			}
		}

		f.files = files
		f.fileItems = make([]widget.Clickable, len(files))
		f.loading = false
		log.Printf("Loaded %d files", len(files))
		f.app.window.Invalidate()
	}()
}

// showRestoreDialog shows the file restore dialog
func (f *FilesView) showRestoreDialog() {
	if f.selectedFile < 0 || f.selectedFile >= len(f.files) {
		f.restoreErrorMsg = "Please select a file to restore"
		return
	}
	f.dialogVisible = true
	f.restoreErrorMsg = ""
	f.app.window.Invalidate()
}

// layoutRestoreDialog renders the restore dialog overlay
func (f *FilesView) layoutRestoreDialog(gtx layout.Context) layout.Dimensions {
	// Handle button clicks
	if f.restoreCancelBtn.Clicked(gtx) {
		f.dialogVisible = false
		f.restoring = false
		f.restoreProgress = ""
		f.restoreErrorMsg = ""
	}

	if f.restoreConfirmBtn.Clicked(gtx) && !f.restoring {
		f.performRestore()
	}

	theme := f.app.GetTheme()

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Dialog size
		gtx.Constraints.Max.X = gtx.Dp(unit.Dp(500))
		gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(400))

		return layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				// Title
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := theme.H6("Restore File")
					return title.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

				// Selected file info
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if f.selectedFile >= 0 && f.selectedFile < len(f.files) {
						file := f.files[f.selectedFile]
						text := theme.Body2(fmt.Sprintf("File: %s\nSize: %s",
							file.Path, formatBytes(file.Size)))
						return text.Layout(gtx)
					}
					return layout.Dimensions{}
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

				// Restore path input
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							label := theme.Body1("Restore to:")
							return label.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							editor := material.Editor(theme.Theme, &f.restorePath, "C:\\Restored\\")
							return editor.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

				// Progress or error message
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if f.restoring {
						text := theme.Body2(f.restoreProgress)
						text.Color = theme.Info
						return text.Layout(gtx)
					}
					if f.restoreErrorMsg != "" {
						text := theme.Body2(f.restoreErrorMsg)
						if strings.Contains(f.restoreErrorMsg, "✅") {
							text.Color = theme.Success
						} else {
							text.Color = theme.Error
						}
						return text.Layout(gtx)
					}
					return layout.Dimensions{}
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

				// Buttons
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:    layout.Horizontal,
						Spacing: layout.SpaceEnd,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(theme.Theme, &f.restoreCancelBtn, "Cancel")
							return btn.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(theme.Theme, &f.restoreConfirmBtn, "Restore")
							btn.Background = theme.Success
							btn.Color = theme.OnPrimary
							if f.restoring {
								// Disable button during restore
								gtx = gtx.Disabled()
							}
							return btn.Layout(gtx)
						}),
					)
				}),
			)
		})
	})
}

// performRestore executes the file restore operation
func (f *FilesView) performRestore() {
	if f.selectedFile < 0 || f.selectedFile >= len(f.files) {
		f.restoreErrorMsg = "Invalid file selection"
		return
	}

	file := f.files[f.selectedFile]
	destPath := f.restorePath.Text()

	if destPath == "" {
		f.restoreErrorMsg = "Please specify a restore destination"
		return
	}

	f.restoring = true
	f.restoreProgress = "Starting restore..."
	f.restoreErrorMsg = ""
	f.app.window.Invalidate()

	go func() {
		log.Printf("Restoring file: %s to %s", file.Path, destPath)

		if f.app.client == nil {
			f.restoring = false
			f.restoreErrorMsg = "Client not initialized"
			f.app.window.Invalidate()
			return
		}

		// Update progress
		f.restoreProgress = "Downloading file..."
		f.app.window.Invalidate()

		// Download the file using the client wrapper
		err := f.app.client.DownloadFile(file.Path, destPath)

		f.restoring = false
		f.restoreProgress = ""

		if err != nil {
			f.restoreErrorMsg = fmt.Sprintf("Failed to restore file: %v", err)
			log.Printf("Restore error: %v", err)
		} else {
			f.restoreErrorMsg = fmt.Sprintf("✅ File successfully restored to:\n%s", destPath)
			log.Printf("File restored successfully to %s", destPath)
		}

		f.app.window.Invalidate()
	}()
}
