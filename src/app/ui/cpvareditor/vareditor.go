package cpvareditor

import (
	"log"

	"sdmm/app/ui/cpwsarea/wsmap/pmap/editor"

	"sdmm/app/config"
	"sdmm/app/ui/shortcut"
	"sdmm/dmapi/dmenv"
	"sdmm/dmapi/dmmap"
	"sdmm/dmapi/dmmap/dmmdata/dmmprefab"
	"sdmm/dmapi/dmmap/dmminstance"
	"sdmm/dmapi/dmvars"
	"sdmm/util/slice"
)

type App interface {
	DoSelectPrefab(prefab *dmmprefab.Prefab)
	CurrentEditor() *editor.Editor
	LoadedEnvironment() *dmenv.Dme

	ConfigRegister(config.Config)
	ConfigFind(name string) config.Config
}

type editMode int

func (e editMode) String() string {
	switch e {
	case emInstance:
		return "editModeInstance"
	case emPrefab:
		return "editModePrefab"
	}
	return ""
}

const (
	emInstance editMode = iota
	emPrefab
)

type VarEditor struct {
	app App

	shortcuts shortcut.Shortcuts

	instance *dmminstance.Instance
	prefab   *dmmprefab.Prefab

	variablesNames []string

	variablesPaths        []string
	variablesNamesByPaths map[string][]string

	sessionEditMode editMode
	sessionPrefabId uint64

	filterVarName  string
	filterTypeName string
}

func (v *VarEditor) Init(app App) {
	v.app = app
	v.addShortcuts()
	v.loadConfig()
}

func (v *VarEditor) Free() {
	v.resetSession()
	log.Println("[cpvareditor] vareditor free")
}

// Sync does the check if we edit an instance which is exists.
// If the instance doesn't exist, then the editor will switch its mode to the prefab editing.
func (v *VarEditor) Sync() {
	if v.prefab == nil {
		return
	}
	e := v.app.CurrentEditor()
	if e == nil || (v.instance != nil && !e.Dmm().IsInstanceExist(v.instance.Id())) {
		v.instance = nil
		v.sessionEditMode = emPrefab
	}
	log.Println("[cpvareditor] synced, editMode:", v.sessionEditMode)
}

func (v *VarEditor) EditInstance(instance *dmminstance.Instance) {
	v.resetSession()
	v.sessionEditMode = emInstance
	v.instance = instance
	v.setup(instance.Prefab())
}

func (v *VarEditor) EditPrefab(prefab *dmmprefab.Prefab) {
	v.resetSession()
	v.sessionEditMode = emPrefab
	v.setup(prefab)
}

func (v *VarEditor) EditedInstance() (*dmminstance.Instance, bool) {
	if v.instance != nil && v.sessionEditMode == emInstance {
		return v.instance, true
	}
	return nil, false
}

func (v *VarEditor) setup(prefab *dmmprefab.Prefab) {
	v.prefab = prefab
	v.variablesNames = collectVariablesNames(prefab.Vars())
	v.variablesPaths = collectVariablesPaths(v.app.LoadedEnvironment().Objects[v.prefab.Path()])
	v.variablesNamesByPaths = collectVariablesNamesByPaths(v.app.LoadedEnvironment(), v.variablesPaths)

	// Clear pinned variables from the common list, since they are showed separately.
	for _, pinnedVarName := range v.config().PinnedVarNames {
		v.variablesNames = slice.StrRemove(v.variablesNames, pinnedVarName)
	}

	log.Printf("[cpvareditor] setup finished: [%s], variablesNames: [%d], variablesPaths: [%d], pinnedVarNames: [%d]",
		prefab.Path(), len(v.variablesNames), len(v.variablesPaths), len(v.config().PinnedVarNames))
}

func (v *VarEditor) setInstanceVariable(varName, varValue string) {
	if len(varValue) == 0 {
		varValue = dmvars.NullValue
	}

	origPrefab := v.instance.Prefab()

	var newVars *dmvars.Variables
	if v.initialVarValue(varName) == varValue {
		newVars = dmvars.Delete(origPrefab.Vars(), varName)
	} else {
		newVars = dmvars.Set(origPrefab.Vars(), varName, varValue)
	}

	newPrefab, isNew := dmmap.PrefabStorage.GetV(origPrefab.Path(), newVars)

	// Newly created prefabs are sort of temporal objects, which need to exist only during the edit session.
	// So if we modified a variable of the instance and that creates a new prefab, the previous one will be deleted.
	if isNew {
		if origPrefab.Id() == v.sessionPrefabId {
			dmmap.PrefabStorage.Delete(origPrefab)
		}
		v.sessionPrefabId = newPrefab.Id()
	}

	v.instance.SetPrefab(newPrefab)
	v.app.CurrentEditor().CommitChanges("Edit Variable")
	v.app.DoSelectPrefab(newPrefab)

	v.prefab = newPrefab

	log.Printf("[cpvareditor] instance [%d] variable set; name: [%s], value: [%s]", v.instance.Id(), varName, varValue)
}

func (v *VarEditor) setPrefabVariable(varName, varValue string) {
	if len(varValue) == 0 {
		varValue = dmvars.NullValue
	}

	var newVars *dmvars.Variables
	if v.initialVarValue(varName) == varValue {
		newVars = dmvars.Delete(v.prefab.Vars(), varName)
	} else {
		newVars = dmvars.Set(v.prefab.Vars(), varName, varValue)
	}

	newPrefab := dmmap.PrefabStorage.Get(v.prefab.Path(), newVars)

	v.app.CurrentEditor().ReplacePrefab(v.prefab, newPrefab)
	v.app.CurrentEditor().CommitChanges("Replace Prefab")

	v.app.DoSelectPrefab(newPrefab)

	v.prefab = newPrefab

	log.Printf("[cpvareditor] prefab [%d] variable set; name: [%s], value: [%s]", v.prefab.Id(), varName, varValue)
}

func (v *VarEditor) resetSession() {
	v.variablesNames = nil
	v.variablesPaths = nil
	v.variablesNamesByPaths = nil
	v.sessionPrefabId = dmmprefab.IdNone
	v.sessionEditMode = emPrefab
	v.instance = nil
	v.prefab = nil
	log.Println("[cpvareditor] session reset")
}

func (v *VarEditor) initialVarValue(varName string) string {
	return v.app.LoadedEnvironment().Objects[v.prefab.Path()].Vars.ValueV(varName, dmvars.NullValue)
}

func (v *VarEditor) isCurrentVarInitial(varName string) bool {
	return v.currentVars().ValueV(varName, dmvars.NullValue) == v.initialVarValue(varName)
}

func (v *VarEditor) doToggleShowModified() {
	cfg := v.config()
	cfg.ShowModified = !cfg.ShowModified
	log.Println("[cpvareditor] toggle 'showModified':", cfg.ShowModified)
}

func (v *VarEditor) doToggleShowByType() {
	cfg := v.config()
	cfg.ShowByType = !cfg.ShowByType
	log.Println("[cpvareditor] toggle 'showByType':", cfg.ShowByType)
}

func (v *VarEditor) doToggleShowPins() {
	cfg := v.config()
	cfg.ShowPins = !cfg.ShowPins
	log.Println("[cpvareditor] toggle 'showPins':", cfg.ShowPins)
}
