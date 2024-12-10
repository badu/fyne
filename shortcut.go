package fyne

import "sync"

// ShortcutHandler is a default implementation of the shortcut handler
// for [CanvasObject].
type ShortcutHandler struct {
	entry map[string]func(Shortcut)
	mu    sync.RWMutex
}

// TypedShortcut handle the registered shortcut
func (sh *ShortcutHandler) TypedShortcut(shortcut Shortcut) {
	sh.mu.RLock()
	fn, has := sh.entry[shortcut.ShortcutName()]
	sh.mu.RUnlock()
	if !has {
		return
	}
	fn(shortcut)
}

// AddShortcut register a handler to be executed when the shortcut action is triggered
func (sh *ShortcutHandler) AddShortcut(shortcut Shortcut, handler func(shortcut Shortcut)) {
	if sh.entry == nil {
		sh.entry = make(map[string]func(Shortcut))
	}

	sh.mu.RLock()
	sh.entry[shortcut.ShortcutName()] = handler
	sh.mu.RUnlock()
}

// RemoveShortcut removes a registered shortcut
func (sh *ShortcutHandler) RemoveShortcut(shortcut Shortcut) {
	sh.mu.Lock()
	_, has := sh.entry[shortcut.ShortcutName()]
	if has {
		delete(sh.entry, shortcut.ShortcutName())
	}
	sh.mu.Unlock()
}

// Shortcut is the interface used to describe a shortcut action
type Shortcut interface {
	ShortcutName() string
}

// KeyboardShortcut describes a shortcut meant to be triggered by a keyboard action.
type KeyboardShortcut interface {
	ShortcutName() string
	Key() KeyName
	Mod() KeyModifier
}

// ShortcutPaste describes a shortcut paste action.
type ShortcutPaste struct {
	Clipboard Clipboard
}

// Key returns the [KeyName] for this shortcut.
//
// Implements: [KeyboardShortcut]
func (se *ShortcutPaste) Key() KeyName {
	return KeyV
}

// Mod returns the [KeyModifier] for this shortcut.
//
// Implements: [KeyboardShortcut]
func (se *ShortcutPaste) Mod() KeyModifier {
	return KeyModifierShortcutDefault
}

// ShortcutName returns the shortcut name
func (se *ShortcutPaste) ShortcutName() string {
	return "Paste"
}

// ShortcutCopy describes a shortcut copy action.
type ShortcutCopy struct {
	Clipboard Clipboard
}

// Key returns the [KeyName] for this shortcut.
//
// Implements: [KeyboardShortcut]
func (se *ShortcutCopy) Key() KeyName {
	return KeyC
}

// Mod returns the [KeyModifier] for this shortcut.
//
// Implements: [KeyboardShortcut]
func (se *ShortcutCopy) Mod() KeyModifier {
	return KeyModifierShortcutDefault
}

// ShortcutName returns the shortcut name
func (se *ShortcutCopy) ShortcutName() string {
	return "Copy"
}

// ShortcutCut describes a shortcut cut action.
type ShortcutCut struct {
	Clipboard Clipboard
}

// Key returns the [KeyName] for this shortcut.
//
// Implements: [KeyboardShortcut]
func (se *ShortcutCut) Key() KeyName {
	return KeyX
}

// Mod returns the [KeyModifier] for this shortcut.
//
// Implements: [KeyboardShortcut]
func (se *ShortcutCut) Mod() KeyModifier {
	return KeyModifierShortcutDefault
}

// ShortcutName returns the shortcut name
func (se *ShortcutCut) ShortcutName() string {
	return "Cut"
}

// ShortcutSelectAll describes a shortcut selectAll action.
type ShortcutSelectAll struct{}

// Key returns the [KeyName] for this shortcut.
//
// Implements: [KeyboardShortcut]
func (se *ShortcutSelectAll) Key() KeyName {
	return KeyA
}

// Mod returns the [KeyModifier] for this shortcut.
//
// Implements: [KeyboardShortcut]
func (se *ShortcutSelectAll) Mod() KeyModifier {
	return KeyModifierShortcutDefault
}

// ShortcutName returns the shortcut name
func (se *ShortcutSelectAll) ShortcutName() string {
	return "SelectAll"
}

// ShortcutUndo describes a shortcut undo action.
//
// Since: 2.5
type ShortcutUndo struct{}

// Key returns the [KeyName] for this shortcut.
//
// Implements: [KeyboardShortcut]
func (se *ShortcutUndo) Key() KeyName {
	return KeyZ
}

// Mod returns the [KeyModifier] for this shortcut.
//
// Implements: [KeyboardShortcut]
func (se *ShortcutUndo) Mod() KeyModifier {
	return KeyModifierShortcutDefault
}

// ShortcutName returns the shortcut name
func (se *ShortcutUndo) ShortcutName() string {
	return "Undo"
}

// ShortcutRedo describes a shortcut redo action.
//
// Since: 2.5
type ShortcutRedo struct{}

// Key returns the [KeyName] for this shortcut.
//
// Implements: [KeyboardShortcut]
func (se *ShortcutRedo) Key() KeyName {
	return KeyY
}

// Mod returns the [KeyModifier] for this shortcut.
//
// Implements: [KeyboardShortcut]
func (se *ShortcutRedo) Mod() KeyModifier {
	return KeyModifierShortcutDefault
}

// ShortcutName returns the shortcut name
func (se *ShortcutRedo) ShortcutName() string {
	return "Redo"
}
