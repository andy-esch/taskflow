package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/wire"
)

func ConfigJSON(w io.Writer, snapshot core.ConfigurationSnapshot) error {
	return wire.EncodeJSON(w, wire.ToConfigEnvelope(snapshot))
}

func ConfigHuman(w io.Writer, st Style, snapshot core.ConfigurationSnapshot) {
	repo := snapshot.Repository
	user := snapshot.User
	effective := snapshot.Effective
	field := func(label, value string) {
		if value == "" {
			value = st.Dim("—")
		}
		fmt.Fprintf(w, "  %s %s\n", st.Dim(fmt.Sprintf("%-18s", label)), value)
	}
	boolValue := func(value *bool) string {
		if value == nil {
			return st.Dim("inherit")
		}
		return fmt.Sprintf("%t", *value)
	}

	fmt.Fprintln(w, st.Bold("Repository"))
	field("config", repo.Path)
	field("mode", string(repo.Mode))
	field("planning root", repo.PlanningRoot)
	field("planning id", repo.ID)
	if repo.Mode == core.ConfigModeScaffold {
		field("taskflow_root", repo.TaskflowRoot)
	}
	if repo.Mode == core.ConfigModePointer {
		field("planning_repo", repo.PlanningRepo)
		field("planning_repo_id", repo.PlanningRepoID)
	}
	field("tracked repos", formatList(repo.TrackedRepos))
	field("theme", emptyAsInherit(st, repo.ThemeName))
	field("pager enabled", boolValue(repo.PagerEnabled))
	field("pager command", emptyAsInherit(st, repo.PagerCommand))
	if len(repo.PendingMigration) > 0 {
		values := make([]string, len(repo.PendingMigration))
		for i, migration := range repo.PendingMigration {
			values[i] = string(migration)
		}
		field("migration", st.Warn(strings.Join(values, ", ")))
		fmt.Fprintf(w, "  %s\n", st.Dim("→ tskflwctl config migrate"))
	}
	if repo.MigrationWarning != "" {
		field("migration", st.Warn(repo.MigrationWarning))
	}

	fmt.Fprintf(w, "\n%s\n", st.Bold("User"))
	field("config", user.Path)
	field("file", map[bool]string{true: "present", false: st.Dim("not created")}[user.Exists])
	field("theme", emptyAsInherit(st, user.ThemeName))
	field("pager enabled", boolValue(user.PagerEnabled))
	field("pager command", emptyAsInherit(st, user.PagerCommand))
	field("space registry", user.RegistryPath)
	if user.Warning != "" {
		field("warning", st.Warn(user.Warning))
	}

	fmt.Fprintf(w, "\n%s\n", st.Bold("Effective"))
	field("theme", effectiveString(effective.Theme))
	field("pager enabled", fmt.Sprintf("%t %s", effective.PagerEnabled.Value, st.Dim("("+string(effective.PagerEnabled.Source)+")")))
	field("pager command", effectiveString(effective.PagerCommand))
}

func effectiveString(value core.EffectiveString) string {
	return fmt.Sprintf("%s (%s)", value.Value, value.Source)
}

func emptyAsInherit(st Style, value string) string {
	if strings.TrimSpace(value) == "" {
		return st.Dim("inherit")
	}
	return value
}

func formatList(values []string) string {
	if len(values) == 0 {
		return "—"
	}
	return strings.Join(values, ", ")
}

func ConfigMigrationJSON(w io.Writer, migration core.ConfigurationMigration, workspace wire.WorkspaceJSON) error {
	return wire.EncodeJSON(w, wire.ToConfigMigrationEnvelope(migration, workspace))
}

func ConfigMigrationHuman(w io.Writer, st Style, migration core.ConfigurationMigration) {
	if !migration.Changed() {
		fmt.Fprintf(w, "%s configuration is current: %s\n", st.Dim("·"), migration.ConfigPath)
		return
	}
	verb := "updated"
	if migration.DryRun {
		verb = "would update"
	}
	fmt.Fprintf(w, "%s %s %s\n", st.Green("✔"), verb, st.Bold(migration.ConfigPath))
	for _, step := range migration.Steps {
		fmt.Fprintf(w, "  %s %s = %q  %s\n", st.Dim("+"), step.Key, step.Value, st.Dim("("+string(step.Kind)+")"))
	}
}
