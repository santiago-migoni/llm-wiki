package vault

// InitRequest contains reviewed generated context, not instructions to execute.
type InitRequest struct{ Manual, Index, LogEntry string }

// PlanInit changes only explicit generated sections and appends audit history.
// _CLAUDE.md and AGENTS.md are never renamed or replaced.
func (v *Vault) PlanInit(req InitRequest) ([]Change, error) {
	changes := []Change{}
	for _, section := range []struct{ path, body string }{{".llm-wiki/manual.md", req.Manual}, {"index.md", req.Index}} {
		part, err := v.PlanGeneratedSection(section.path, section.body)
		if err != nil {
			return nil, err
		}
		changes = append(changes, part...)
	}
	if req.LogEntry != "" {
		part, err := v.PlanAuditEntry("log.md", req.LogEntry)
		if err != nil {
			return nil, err
		}
		changes = append(changes, part...)
	}
	return changes, nil
}
