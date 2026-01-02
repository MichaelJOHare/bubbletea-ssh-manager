package tui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	str "bubbletea-ssh-manager/internal/stringutil"

	tea "github.com/charmbracelet/bubbletea"
)

// handleHostFormSubmit processes the submitted host form data.
//
// It validates the data, closes the form, and returns a command to save the host.
// If there are validation errors, it closes the form with an error status.
func (m model) handleHostFormSubmit(msg formResultMsg) (model, tea.Cmd) {
	if msg.kind != formResultSubmitted {
		return m, nil
	}
	protocol := str.NormalizeString(msg.protocol)

	alias, err := str.BuildAliasFromGroupNickname(msg.group, msg.nickname)
	if err != nil {
		m, _ = m.closeHostForm("", statusInfo)
		return m, m.setStatusError("❌ Invalid group/nickname: "+err.Error(), 0)
		// probably won't need this after making enter not submit on validation errors?
	}
	msg.spec.Alias = alias

	oldAlias := strings.TrimSpace(m.ms.hostFormOldAlias)
	if oldAlias == "" {
		oldAlias = strings.TrimSpace(msg.oldAlias)
	}

	// close the form before doing IO
	m, _ = m.closeHostForm("", statusInfo)

	saveCmd := func() tea.Msg {
		result := formSaveResultMsg{protocol: protocol, spec: msg.spec}
		switch msg.mode {
		case modeAdd:
			root, err := getProtocolConfigPath(protocol)
			if err != nil {
				result.err = err
				return result
			}
			err = AddHostToRootConfig(protocol, msg.spec, msg.opts)
			result.err = err
			if err == nil {
				result.configPath = root
			}
			return result
		case modeEdit:
			if oldAlias == "" {
				result.err = errors.New("missing old alias")
				return result
			}
			configPath, err := getConfigPathForAlias(protocol, oldAlias)
			if err != nil {
				result.err = err
				return result
			}
			if strings.TrimSpace(configPath) == "" {
				result.err = os.ErrNotExist
				return result
			}
			result.configPath = configPath
			result.err = UpdateHostInConfig(protocol, oldAlias, msg.spec, msg.opts)
			return result
		default:
			result.err = errors.New("unknown form mode")
			return result
		}
	}

	return m, tea.Cmd(func() tea.Msg { return saveCmd() })
}

// handleHostFormSaveResult processes the result of saving a host form.
//
// It updates the status based on whether the save was successful or if there
// were errors.
func (m model) handleHostFormSaveResult(msg formSaveResultMsg) (model, tea.Cmd) {
	if msg.err == nil {
		root, err := seedMenu()
		cmd := func() tea.Msg {
			return menuReloadedMsg{root: root, err: err}
		}
		alias := strings.TrimSpace(msg.spec.Alias)
		hostName := strings.TrimSpace(msg.spec.HostName)
		targetText := alias
		if hostName != "" {
			targetText = fmt.Sprintf("%s <%s>", alias, hostName)
		}
		status := fmt.Sprintf("✔️ Saved Host %s to %s", targetText, msg.configPath)
		return m, tea.Batch(m.setStatusSuccess(status, statusTTL), cmd)
	}
	if errors.Is(msg.err, os.ErrNotExist) {
		return m, m.setStatusError("❌ Host not found.", statusTTL)
	}
	return m, m.setStatusError("❌ Save failed: "+msg.err.Error(), 0)
}
