package tui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"bubbletea-ssh-manager/internal/config"
	str "bubbletea-ssh-manager/internal/stringutil"

	tea "github.com/charmbracelet/bubbletea"
)

// handleHostFormSubmit processes the submitted host form data.
//
// It builds the alias, closes the form, and returns a command to save the host.
// Validation errors are caught earlier in handleHostFormKeyMsg.
func (m model) handleHostFormSubmit(msg formSubmittedMsg) (model, tea.Cmd) {
	protocol := msg.protocol
	if protocol == "" {
		protocol = config.ProtocolSSH
	}

	alias, _ := str.BuildAliasFromGroupNickname(msg.group, msg.nickname)
	msg.spec.Alias = alias
	msg.spec = msg.spec.Normalized()
	msg.opts = msg.opts.Normalized()

	oldAlias := m.ms.hostFormOldAlias
	if oldAlias == "" {
		oldAlias = msg.oldAlias
	}

	// close the form before doing IO
	m, _ = m.closeHostForm("", statusInfo)

	return m, m.saveHostCmd(msg.mode, protocol, oldAlias, msg.spec, msg.opts)
}

// saveHostCmd returns a command that performs the host save operation.
func (m model) saveHostCmd(mode formMode, protocol config.Protocol, oldAlias string, spec config.Spec, opts config.SSHOptions) tea.Cmd {
	return func() tea.Msg {
		result := formSaveResultMsg{protocol: protocol, spec: spec}

		switch mode {
		case modeAdd:
			root, err := config.GetConfigPathForProtocol(protocol)
			if err != nil {
				result.err = err
				return result
			}
			result.err = config.AddHostToRootConfig(protocol, spec, opts)
			if result.err == nil {
				result.configPath = root
			}

		case modeEdit:
			if oldAlias == "" {
				result.err = errors.New("missing old alias")
				return result
			}
			configPath, err := config.GetConfigPathForAlias(protocol, oldAlias)
			if err != nil {
				result.err = err
				return result
			}
			if strings.TrimSpace(configPath) == "" {
				result.err = os.ErrNotExist
				return result
			}
			result.configPath = configPath
			result.err = config.UpdateHostInConfig(protocol, oldAlias, spec, opts)

		default:
			result.err = errors.New("unknown form mode")
		}

		return result
	}
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
		alias := msg.spec.Alias
		hostName := msg.spec.HostName
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
