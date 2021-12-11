package cpvareditor

import (
	"fmt"
	"log"
	"strings"

	"github.com/SpaiR/imgui-go"
	"sdmm/dmapi/dmvars"
	"sdmm/imguiext/icon"
	"sdmm/imguiext/style"
	w "sdmm/imguiext/widget"
)

func (v *VarEditor) Process() {
	v.shortcuts.SetVisible(imgui.IsWindowFocusedV(imgui.FocusedFlagsRootAndChildWindows))

	if len(v.variablesNames) == 0 {
		imgui.TextDisabled("No Instance/Prefab Selected")
		return
	}

	v.showEditModeToggle()
	v.showControls()
	imgui.Separator()
	imgui.BeginChild("variables")
	v.showVariables()
	imgui.EndChild()
}

func (v *VarEditor) showEditModeToggle() {
	if imgui.BeginTableV("edit_mode_toggle", 2, imgui.TableFlagsNoPadInnerX, imgui.Vec2{}, 0) {
		imgui.PushStyleVarVec2(imgui.StyleVarButtonTextAlign, imgui.Vec2{X: .5, Y: .5})

		imgui.TableNextColumn()

		if v.instance == nil {
			imgui.BeginDisabled()
		}

		v.showInstanceModeButton()

		if v.instance == nil {
			imgui.EndDisabled()
		}

		imgui.TableNextColumn()

		v.showPrefabModeButton()

		imgui.PopStyleVar()
		imgui.EndTable()
	}
}

func (v *VarEditor) showInstanceModeButton() {
	var buttonStyle w.ButtonStyle
	if v.sessionEditMode == emInstance {
		buttonStyle = style.ButtonGreen{}
	} else {
		buttonStyle = style.ButtonDefault{}
	}

	w.Button("Instance", func() {
		v.sessionEditMode = emInstance
		log.Println("[cpvareditor] set instance mode")
	}).Style(buttonStyle).Size(imgui.Vec2{X: -1}).Build()
}

func (v *VarEditor) showPrefabModeButton() {
	var buttonStyle w.ButtonStyle
	if v.sessionEditMode == emPrefab {
		buttonStyle = style.ButtonGreen{}
	} else {
		buttonStyle = style.ButtonDefault{}
	}

	w.Button("Prefab", func() {
		v.sessionEditMode = emPrefab
		log.Println("[cpvareditor] set prefab mode")
	}).Style(buttonStyle).Size(imgui.Vec2{X: -1}).Build()
}

func (v *VarEditor) showControls() {
	imgui.Button(icon.FaFilter)
	if imgui.BeginPopupContextItemV("var_editor_filter", imgui.PopupFlagsMouseButtonLeft) {
		if imgui.MenuItemV("Show modified only", "Ctrl+1", v.showModified, true) {
			v.doToggleShowModified()
		}
		if imgui.MenuItemV("Show with types", "Ctrl+2", v.showByType, true) {
			v.doToggleShowByType()
		}
		imgui.EndPopup()
	}

	imgui.SameLine()

	w.InputTextWithHint("##filter_var_name", v.filterVarNameHint(), &v.filterVarName).
		ButtonClear().
		Width(-1).
		Build()

	if v.showByType {
		w.InputTextWithHint("##filter_type_name", "Filter Type", &v.filterTypeName).
			ButtonClear().
			Width(-1).
			Build()
	}
}

func (v *VarEditor) filterVarNameHint() string {
	if v.showByType {
		return "Filter Name"
	}
	return "Filter"
}

const (
	varsTableFlags = imgui.TableFlagsResizable | imgui.TableFlagsBordersInner
	varsInputFlags = imgui.InputTextFlagsAutoSelectAll | imgui.InputTextFlagsEnterReturnsTrue | imgui.InputTextFlagsCtrlEnterForNewLine
)

func (v *VarEditor) showVariables() {
	if v.showByType {
		v.showVariablesByType()
	} else {
		v.showAllVariables()
	}
}

func (v *VarEditor) showVariablesByType() {
	for _, path := range v.variablesPaths {
		if v.isFilteredPath(path) {
			continue
		}

		variablesNames := v.variablesNamesByPaths[path]

		imgui.TextColored(style.ColorGold, path)
		imgui.SameLine()
		imgui.TextDisabled(fmt.Sprintf("(%d)", len(variablesNames)))

		if imgui.BeginTableV("variables", 2, varsTableFlags, imgui.Vec2{}, 0) {
			v.showVariablesNames(variablesNames)
			imgui.EndTable()
		}
	}
}

func (v *VarEditor) showAllVariables() {
	if imgui.BeginTableV("variables", 2, varsTableFlags, imgui.Vec2{}, 0) {
		v.showVariablesNames(v.variablesNames)
		imgui.EndTable()
	}
}

func (v *VarEditor) showVariablesNames(variablesNames []string) {
	for _, varName := range variablesNames {
		v.showVariable(varName)
	}
}

func (v *VarEditor) showVariable(varName string) {
	if v.isFilteredVariable(varName) {
		return
	}

	imgui.TableNextColumn()
	v.showVarName(varName)
	imgui.TableNextColumn()
	v.showVarInput(varName)
}

func (v *VarEditor) showVarName(varName string) {
	if !v.isCurrentVarInitial(varName) {
		imgui.TextColored(style.ColorGreen3, varName)
	} else {
		imgui.Text(varName)
	}
}

func (v *VarEditor) showVarInput(varName string) {
	varValue := v.currentVars().ValueV(varName, dmvars.NullValue)
	initialValue := v.initialVarValue(varName)
	isModified := initialValue != varValue

	var resetBtn *w.ButtonWidget
	if isModified {
		resetBtn = w.Button(icon.FaUndo+"##"+varName, func() {
			v.setCurrentVariable(varName, initialValue)
		}).Tooltip(initialValue).Style(style.ButtonFrame{})
	}

	w.InputText(fmt.Sprint("##", v.prefab.Id(), varName), &varValue).
		Button(resetBtn).
		Width(-1).
		Flags(varsInputFlags).
		OnDeactivatedAfterEdit(func() {
			v.setCurrentVariable(varName, varValue)
		}).
		Build()
}

func (v *VarEditor) setCurrentVariable(varName, varValue string) {
	if v.sessionEditMode == emInstance {
		v.setInstanceVariable(varName, varValue)
	} else {
		v.setPrefabVariable(varName, varValue)
	}
}

func (v *VarEditor) currentVars() *dmvars.Variables {
	if v.sessionEditMode == emInstance {
		return v.instance.Prefab().Vars()
	}
	return v.prefab.Vars()
}

func (v *VarEditor) isFilteredVariable(varName string) bool {
	// Show modified only
	if v.showModified && v.isCurrentVarInitial(varName) {
		return true
	}
	// Show filtered by name only
	if len(v.filterVarName) > 0 && !strings.Contains(varName, v.filterVarName) {
		return true
	}
	return false
}

func (v *VarEditor) isFilteredPath(path string) bool {
	return v.showByType && len(v.filterTypeName) > 0 && !strings.Contains(path, v.filterTypeName)
}
