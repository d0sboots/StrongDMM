package menu

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"sdmm/app/ui/shortcut"
)

func (m *Menu) addShortcuts() {
	m.shortcuts.Add(shortcut.Shortcut{
		Name:        "menu#DoOpenMap",
		FirstKey:    glfw.KeyLeftControl,
		FirstKeyAlt: glfw.KeyRightControl,
		SecondKey:   glfw.KeyO,
		Action:      m.app.DoOpenMap,
		IsEnabled:   m.app.HasLoadedEnvironment,
	})

	m.shortcuts.Add(shortcut.Shortcut{
		Name:        "menu#DoSave",
		FirstKey:    glfw.KeyLeftControl,
		FirstKeyAlt: glfw.KeyRightControl,
		SecondKey:   glfw.KeyS,
		Action:      m.app.DoSave,
		IsEnabled:   m.app.HasActiveMap,
	})

	m.shortcuts.Add(shortcut.Shortcut{
		Name:        "menu#DoUndo",
		FirstKey:    glfw.KeyLeftControl,
		FirstKeyAlt: glfw.KeyRightControl,
		SecondKey:   glfw.KeyZ,
		Action:      m.app.DoUndo,
		IsEnabled:   m.app.CommandStorage().HasUndo,
	})

	m.shortcuts.Add(shortcut.Shortcut{
		Name:         "menu#DoRedo",
		FirstKey:     glfw.KeyLeftControl,
		FirstKeyAlt:  glfw.KeyRightControl,
		SecondKey:    glfw.KeyLeftShift,
		SecondKeyAlt: glfw.KeyRightShift,
		ThirdKey:     glfw.KeyZ,
		Action:       m.app.DoRedo,
		IsEnabled:    m.app.CommandStorage().HasRedo,
	})
	m.shortcuts.Add(shortcut.Shortcut{
		Name:        "menu#DoRedo",
		FirstKey:    glfw.KeyLeftControl,
		FirstKeyAlt: glfw.KeyRightControl,
		SecondKey:   glfw.KeyY,
		Action:      m.app.DoRedo,
		IsEnabled:   m.app.CommandStorage().HasRedo,
	})

	m.shortcuts.Add(shortcut.Shortcut{
		Name:        "menu#DoCopy",
		FirstKey:    glfw.KeyLeftControl,
		FirstKeyAlt: glfw.KeyRightControl,
		SecondKey:   glfw.KeyC,
		Action:      m.app.DoCopy,
	})
	m.shortcuts.Add(shortcut.Shortcut{
		Name:        "menu#DoPaste",
		FirstKey:    glfw.KeyLeftControl,
		FirstKeyAlt: glfw.KeyRightControl,
		SecondKey:   glfw.KeyV,
		Action:      m.app.DoPaste,
		IsEnabled:   m.app.Clipboard().HasData,
	})
	m.shortcuts.Add(shortcut.Shortcut{
		Name:        "menu#DoCut",
		FirstKey:    glfw.KeyLeftControl,
		FirstKeyAlt: glfw.KeyRightControl,
		SecondKey:   glfw.KeyX,
		Action:      m.app.DoCut,
	})
	m.shortcuts.Add(shortcut.Shortcut{
		Name:     "menu#DoDelete",
		FirstKey: glfw.KeyDelete,
		Action:   m.app.DoDelete,
	})
	m.shortcuts.Add(shortcut.Shortcut{
		Name:        "menu#DoSearch",
		FirstKey:    glfw.KeyLeftControl,
		FirstKeyAlt: glfw.KeyRightControl,
		SecondKey:   glfw.KeyF,
		Action:      m.app.DoSearch,
		IsEnabled:   m.app.HasActiveMap,
	})

	m.shortcuts.Add(shortcut.Shortcut{
		Name:     "menu#DoResetLayout",
		FirstKey: glfw.KeyF5,
		Action:   m.app.DoResetLayout,
	})

	m.shortcuts.SetVisible(true)
}
