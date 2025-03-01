package pmap

import (
	"log"

	"sdmm/app/ui/cpwsarea/wsmap/tools"
	"sdmm/imguiext/icon"
	"sdmm/imguiext/style"
	w "sdmm/imguiext/widget"
)

type toolDesc struct {
	btnIcon string
	tooltip w.Layout
}

const tSeparator = "toolsSeparator"

var (
	toolsOrder = []string{
		tools.TNAdd,
		tools.TNFill,
		tools.TNGrab,
		tSeparator,
		tools.TNPick,
		tools.TNDelete,
		tools.TNReplace,
	}

	toolsDesc = map[string]toolDesc{
		tools.TNAdd: {
			btnIcon: icon.Add,
			tooltip: w.Layout{
				w.AlignTextToFramePadding(),
				w.Text(tools.TNAdd),
				w.SameLine(),
				w.TextFrame("1"),
				w.Separator(),
				w.Text("Place the selected object\nAlt: place the selected object with replace"),
			},
		},
		tools.TNFill: {
			btnIcon: icon.BorderAll,
			tooltip: w.Layout{
				w.AlignTextToFramePadding(),
				w.Text(tools.TNFill),
				w.SameLine(),
				w.TextFrame("2"),
				w.Separator(),
				w.Text("Fill the area with the selected object\nAlt: fill the selected area with the selected object with replace"),
			},
		},
		tools.TNGrab: {
			btnIcon: icon.BorderStyle,
			tooltip: w.Layout{
				w.AlignTextToFramePadding(),
				w.Text(tools.TNGrab),
				w.SameLine(),
				w.TextFrame("3"),
				w.Separator(),
				w.Text("Select the area / Move the selection with visible objects inside"),
			},
		},
		tools.TNPick: {
			btnIcon: icon.EyeDropper,
			tooltip: w.Layout{
				w.AlignTextToFramePadding(),
				w.Text(tools.TNPick),
				w.SameLine(),
				w.TextFrame("Hold S"),
				w.Separator(),
				w.Text("Pick the hovered instance"),
			},
		},
		tools.TNDelete: {
			btnIcon: icon.Eraser,
			tooltip: w.Layout{
				w.AlignTextToFramePadding(),
				w.Text(tools.TNDelete),
				w.SameLine(),
				w.TextFrame("Hold D"),
				w.Separator(),
				w.Text("Delete the hovered instance\nAlt: delete the whole tile"),
			},
		},
		tools.TNReplace: {
			btnIcon: icon.Repeat,
			tooltip: w.Layout{
				w.AlignTextToFramePadding(),
				w.Text(tools.TNReplace),
				w.SameLine(),
				w.TextFrame("Hold R"),
				w.Separator(),
				w.Text("Replace the hovered instance with the selected object"),
			},
		},
	}
)

func (p *PaneMap) showToolsPanel() {
	w.Layout{
		p.panelToolsLayoutTools(),
		w.SameLine(),
		w.Layout{
			w.AlignRight,
			p.panelToolsLayoutSettings(),
		},
	}.Build()
}

func (p *PaneMap) panelToolsLayoutTools() (layout w.Layout) {
	for idx, toolName := range toolsOrder {
		var toolName = toolName // Closure (hello, js)

		if idx > 0 || idx < len(toolsOrder)-1 {
			layout = append(layout, w.SameLine())
		}

		if toolName == tSeparator {
			layout = append(layout, w.TextDisabled("|"))
			continue
		}

		tool := tools.Tools()[toolName]
		desc := toolsDesc[toolName]

		btn := w.Button(desc.btnIcon, func() {
			tools.SetSelected(toolName)
		}).Round(true)

		if tools.Selected() == tool {
			if tool.AltBehaviour() {
				btn.Style(style.ButtonGold{}).TextColor(style.ColorBlack)
			} else {
				btn.Style(style.ButtonGreen{})
			}
		}

		layout = append(layout, btn, w.Tooltip(desc.tooltip))
	}
	return layout
}

func (p *PaneMap) panelToolsLayoutSettings() w.Layout {
	var bntStyle w.ButtonStyle
	if p.showSettings {
		bntStyle = style.ButtonGreen{}
	} else {
		bntStyle = style.ButtonDefault{}
	}
	return w.Layout{
		w.Button(icon.Cog, p.doToggleSettings).
			Tooltip("Settings").
			Style(bntStyle).
			Round(true),
	}
}

func (p *PaneMap) doToggleSettings() {
	log.Println("[pmap] toggle settings:")
	p.showSettings = !p.showSettings
}

func (p *PaneMap) doPreviousLevel() {
	if p.hasPreviousLevel() {
		p.activeLevel--
		log.Println("[pmap] active level switched to previous:", p.activeLevel)
	}
}

func (p *PaneMap) doNextLevel() {
	if p.hasNextLevel() {
		p.activeLevel++
		log.Println("[pmap] active level switched to next:", p.activeLevel)
	}
}

func (p *PaneMap) hasPreviousLevel() bool {
	return p.activeLevel > 1
}

func (p *PaneMap) hasNextLevel() bool {
	return p.activeLevel < p.dmm.MaxZ
}
