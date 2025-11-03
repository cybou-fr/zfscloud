package gui

import (
	"fmt"
	"log"
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

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Spacing:   layout.SpaceBetween,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			title := material.H5(f.app.theme, "Backup Files")
			return title.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Horizontal,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(f.app.theme, &f.refreshButton, "Refresh")
					return btn.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(f.app.theme, &f.restoreButton, "Restore...")
					btn.Background = f.app.theme.Palette.ContrastBg
					return btn.Layout(gtx)
				}),
			)
		}),
	)
}

// layoutFileList renders the list of backed up files
func (f *FilesView) layoutFileList(gtx layout.Context) layout.Dimensions {
	// Show loading indicator
	if f.loading {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return material.Body1(f.app.theme, "Loading files...").Layout(gtx)
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
					text := material.Body1(f.app.theme, "Error:")
					text.Color = f.app.theme.Palette.Fg
					return text.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					text := material.Body2(f.app.theme, f.errorMsg)
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
					text := material.H6(f.app.theme, "No files to display")
					return text.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					text := material.Body1(f.app.theme, "Start a backup to see files here")
					return text.Layout(gtx)
				}),
			)
		})
	}

	// Render scrollable file list
	return material.List(f.app.theme, &f.fileList).Layout(gtx, len(f.files), func(gtx layout.Context, index int) layout.Dimensions {
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
				label := material.Body1(f.app.theme, text)
				if isSelected {
					label.Font.Weight = 700 // Bold weight
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
						label := material.Caption(f.app.theme, fmt.Sprintf("Size: %s", sizeStr))
						return label.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Caption(f.app.theme, fmt.Sprintf("Modified: %s", timeStr))
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
					title := material.H6(f.app.theme, "Restore File")
					return title.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

				// Selected file info
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if f.selectedFile >= 0 && f.selectedFile < len(f.files) {
						file := f.files[f.selectedFile]
						text := material.Body2(f.app.theme, fmt.Sprintf("File: %s\nSize: %s",
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
							label := material.Body1(f.app.theme, "Restore to:")
							return label.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							editor := material.Editor(f.app.theme, &f.restorePath, "C:\\Restored\\")
							return editor.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

				// Progress or error message
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if f.restoring {
						text := material.Body2(f.app.theme, f.restoreProgress)
						return text.Layout(gtx)
					}
					if f.restoreErrorMsg != "" {
						text := material.Body2(f.app.theme, f.restoreErrorMsg)
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
							btn := material.Button(f.app.theme, &f.restoreCancelBtn, "Cancel")
							return btn.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(f.app.theme, &f.restoreConfirmBtn, "Restore")
							btn.Background = f.app.theme.Palette.ContrastBg
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

		// TODO: Implement actual restore via DownloadFile RPC
		// For now, show placeholder message
		time.Sleep(2 * time.Second)

		f.restoring = false
		f.restoreProgress = ""
		f.restoreErrorMsg = "Restore functionality coming soon! (DownloadFile RPC not yet implemented)"
		f.app.window.Invalidate()
	}()
}
