//go:build threadspike

package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/threadspike"
)

func addThreadSpikeCommand(root *cobra.Command, app *App) {
	root.AddCommand(newThreadSpikeCmd(app))
}

func newThreadSpikeCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thread",
		Short: "EXPERIMENTAL: exercise ADR-0006 Thread DAGs in a throwaway planning space",
		Long: "This command is compiled only with `-tags threadspike`. It is a manual spike\n" +
			"surface, not a production CLI or wire contract. Use it only with disposable data.",
	}
	cmd.AddCommand(
		newThreadSpikeComposeCmd(app),
		newThreadSpikeApplyCmd(app),
		newThreadSpikeListCmd(app),
		newThreadSpikeShowCmd(app),
		newThreadSpikePlanCmd(app),
	)
	return cmd
}

func threadSpikeRepository(app *App) *store.ThreadSpikeRepository {
	return store.NewThreadSpikeRepository(app.Cfg.Root, app.Cfg.ID)
}

func newThreadSpikeComposeCmd(app *App) *cobra.Command {
	var from, outPath string
	cmd := &cobra.Command{
		Use:         "compose",
		Short:       "Compile a Thread authoring manifest into a durable stable-ID apply plan",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "mutating"},
		RunE: func(_ *cobra.Command, _ []string) error {
			if strings.TrimSpace(from) == "" || strings.TrimSpace(outPath) == "" {
				return fmt.Errorf("%w: --from and --out are required", domain.ErrValidation)
			}
			content, err := readThreadSpikeInput(app.In, from)
			if err != nil {
				return err
			}
			var manifest threadspike.Manifest
			if err := yaml.Unmarshal(content, &manifest); err != nil {
				return fmt.Errorf("%w: parse Thread manifest: %v", domain.ErrValidation, err)
			}
			snapshot, err := threadSpikeRepository(app).Snapshot()
			if err != nil {
				return err
			}
			plan, err := threadspike.Compose(snapshot, manifest, id.New, time.Now())
			if err != nil {
				return err
			}
			encoded, err := yaml.Marshal(plan)
			if err != nil {
				return fmt.Errorf("encode Thread apply plan: %w", err)
			}
			if app.DryRun {
				_, err = app.Out.Write(encoded)
				return err
			}
			if err := writeThreadSpikePlan(outPath, encoded); err != nil {
				return err
			}
			if app.JSON {
				return writeThreadSpikeJSON(app.Out, map[string]any{
					"experimental": true, "thread_id": plan.Thread.ID,
					"thread_slug": plan.Thread.Slug, "plan": outPath,
				})
			}
			fmt.Fprintf(app.Out, "composed Thread %s (%s) -> %s\n", plan.Thread.ID, plan.Thread.Slug, outPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "authoring manifest path, or - for stdin")
	cmd.Flags().StringVar(&outPath, "out", "", "new durable apply-plan path (must not already exist)")
	return cmd
}

func newThreadSpikeApplyCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:         "apply <materialized-plan>",
		Short:       "Converge the throwaway planning space toward a materialized Thread plan",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"safety": "mutating"},
		RunE: func(_ *cobra.Command, args []string) error {
			content, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read Thread apply plan %s: %w", args[0], err)
			}
			var plan threadspike.MaterializedPlan
			if err := yaml.Unmarshal(content, &plan); err != nil {
				return fmt.Errorf("%w: parse Thread apply plan: %v", domain.ErrValidation, err)
			}
			receipt, err := threadSpikeRepository(app).Apply(plan, threadspike.ApplyOptions{DryRun: app.DryRun})
			if err != nil {
				return err
			}
			if app.JSON {
				return writeThreadSpikeJSON(app.Out, map[string]any{
					"experimental": true, "dry_run": app.DryRun, "receipt": receipt,
				})
			}
			for _, operation := range receipt.Entries {
				fmt.Fprintf(app.Out, "%s %s %s\n", operation.Action, operation.Kind, operation.ID)
			}
			return nil
		},
	}
}

func newThreadSpikeListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List experimental Threads and their derived rollups",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			snapshot, graph, err := loadThreadSpikeGraph(app)
			if err != nil {
				return err
			}
			ids := make([]string, 0, len(snapshot.Threads))
			for threadID := range snapshot.Threads {
				ids = append(ids, threadID)
			}
			sort.Strings(ids)
			views := make([]threadspike.ThreadView, 0, len(ids))
			for _, threadID := range ids {
				views = append(views, graph.ViewThread(snapshot.Threads[threadID]))
			}
			if app.JSON {
				return writeThreadSpikeJSON(app.Out, map[string]any{"experimental": true, "threads": views})
			}
			for _, view := range views {
				fmt.Fprintf(app.Out, "%s  %-24s  %-11s  %d/%d  frontier:%d%s\n",
					view.Thread.ID, view.Thread.Slug, view.Thread.Status, view.Done, view.Total,
					len(view.Frontier), threadSpikeHealth(view))
			}
			return nil
		},
	}
}

func newThreadSpikeShowCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "show <thread>",
		Short: "Show members, external gates, frontier, rollup, and graph health",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			snapshot, graph, err := loadThreadSpikeGraph(app)
			if err != nil {
				return err
			}
			thread, err := resolveThreadSpike(snapshot, args[0])
			if err != nil {
				return err
			}
			view := graph.ViewThread(thread)
			if app.JSON {
				return writeThreadSpikeJSON(app.Out, map[string]any{"experimental": true, "view": view})
			}
			fmt.Fprintf(app.Out, "Thread %s (%s)  %s  %d/%d%s\n", thread.ID, thread.Slug, thread.Status, view.Done, view.Total, threadSpikeHealth(view))
			fmt.Fprintln(app.Out, "members:")
			for _, member := range view.Members {
				fmt.Fprintf(app.Out, "  %s  %-24s  %-18s gate:%-7s eligible:%t inconsistent:%t\n",
					member.ID, member.Slug, member.Role, member.Gate, member.Eligible, member.Inconsistent)
			}
			if len(view.ExternalGates) > 0 {
				fmt.Fprintln(app.Out, "external gates:")
				for _, gate := range view.ExternalGates {
					fmt.Fprintf(app.Out, "  %s  %-24s  %-18s gate:%s\n", gate.ID, gate.Slug, gate.Role, gate.Gate)
				}
			}
			fmt.Fprintf(app.Out, "frontier: %s\n", strings.Join(view.Frontier, ", "))
			return nil
		},
	}
}

func newThreadSpikePlanCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "plan <thread>",
		Short: "Show deterministic explanatory topological waves for Thread members",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			snapshot, graph, err := loadThreadSpikeGraph(app)
			if err != nil {
				return err
			}
			thread, err := resolveThreadSpike(snapshot, args[0])
			if err != nil {
				return err
			}
			waves, err := graph.Plan(thread.Tasks)
			if err != nil {
				return err
			}
			view := graph.ViewThread(thread)
			if app.JSON {
				return writeThreadSpikeJSON(app.Out, map[string]any{
					"experimental": true, "thread_id": thread.ID, "waves": waves,
					"external_gates": view.ExternalGates,
				})
			}
			for i, wave := range waves {
				fmt.Fprintf(app.Out, "wave %d:\n", i+1)
				for _, taskID := range wave {
					task := snapshot.Tasks[taskID]
					fmt.Fprintf(app.Out, "  %s  %s\n", taskID, task.Record.Slug)
				}
			}
			if len(view.ExternalGates) > 0 {
				fmt.Fprintln(app.Out, "external gates:")
				for _, gate := range view.ExternalGates {
					fmt.Fprintf(app.Out, "  %s  %s\n", gate.ID, gate.Slug)
				}
			}
			return nil
		},
	}
}

func loadThreadSpikeGraph(app *App) (threadspike.Snapshot, *threadspike.Graph, error) {
	snapshot, err := threadSpikeRepository(app).Snapshot()
	if err != nil {
		return threadspike.Snapshot{}, nil, err
	}
	if len(snapshot.Problems) > 0 {
		return threadspike.Snapshot{}, nil, fmt.Errorf("%w: planning graph cannot be read soundly: %s: %s",
			domain.ErrValidation, snapshot.Problems[0].Path, snapshot.Problems[0].Message)
	}
	graph := threadspike.NewGraph(snapshot.Tasks)
	if err := graph.Validate(); err != nil {
		return threadspike.Snapshot{}, nil, err
	}
	return snapshot, graph, nil
}

func resolveThreadSpike(snapshot threadspike.Snapshot, query string) (threadspike.Thread, error) {
	query = strings.TrimSpace(query)
	if query == "" || strings.ContainsAny(query, `/\\`) || strings.Contains(query, "..") {
		return threadspike.Thread{}, fmt.Errorf("%w: Thread name %q must be a plain name", domain.ErrValidation, query)
	}
	ids := make([]string, 0, len(snapshot.Threads))
	for threadID := range snapshot.Threads {
		ids = append(ids, threadID)
	}
	sort.Strings(ids)
	q := strings.ToLower(query)
	tiers := []func(string) bool{
		func(value string) bool { return strings.EqualFold(value, q) },
		func(value string) bool { return strings.HasPrefix(strings.ToLower(value), q) },
		func(value string) bool { return strings.Contains(strings.ToLower(value), q) },
	}
	for _, match := range tiers {
		var hits []threadspike.Thread
		for _, threadID := range ids {
			thread := snapshot.Threads[threadID]
			if match(thread.ID) || match(thread.Slug) {
				hits = append(hits, thread)
			}
		}
		switch len(hits) {
		case 1:
			return hits[0], nil
		case 0:
			continue
		default:
			return threadspike.Thread{}, fmt.Errorf("thread %q matches %d Threads: %w", query, len(hits), domain.ErrAmbiguous)
		}
	}
	return threadspike.Thread{}, fmt.Errorf("thread %q: %w", query, domain.ErrNotFound)
}

func readThreadSpikeInput(in io.Reader, path string) ([]byte, error) {
	if path == "-" {
		content, err := io.ReadAll(in)
		if err != nil {
			return nil, fmt.Errorf("read Thread manifest from stdin: %w", err)
		}
		return content, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Thread manifest %s: %w", path, err)
	}
	return content, nil
}

func writeThreadSpikePlan(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create apply-plan directory %s: %w", dir, err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("apply plan %s already exists: %w", path, domain.ErrConflict)
		}
		return fmt.Errorf("create apply plan %s: %w", path, err)
	}
	remove := true
	defer func() {
		_ = file.Close()
		if remove {
			_ = os.Remove(path)
		}
	}()
	if _, err := file.Write(content); err != nil {
		return fmt.Errorf("write apply plan %s: %w", path, err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync apply plan %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close apply plan %s: %w", path, err)
	}
	remove = false
	return nil
}

func writeThreadSpikeJSON(out io.Writer, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func threadSpikeHealth(view threadspike.ThreadView) string {
	switch {
	case view.Broken:
		return "  broken"
	case view.Inconsistent:
		return "  inconsistent"
	default:
		return ""
	}
}
