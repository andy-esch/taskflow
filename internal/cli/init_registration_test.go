package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/userconfig"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestInitAutoRegistration_ScaffoldAndPointerUseCheckoutWithPlanningIdentity(t *testing.T) {
	t.Setenv(userconfig.DirEnv, t.TempDir())
	parent := t.TempDir()
	planning := filepath.Join(parent, "planning")

	out, errOut, err := runIn(t, parent, "init", "--path", planning,
		"--taskflow-root", "planning", "--json")
	if err != nil {
		t.Fatalf("scaffold init: %v\nstdout:\n%s\nstderr:\n%s", err, out, errOut)
	}
	var scaffold wire.InitEnvelope
	if err := json.Unmarshal([]byte(out), &scaffold); err != nil {
		t.Fatalf("decode scaffold receipt: %v\n%s", err, out)
	}
	planningConfig, err := config.Discover(planning)
	if err != nil {
		t.Fatal(err)
	}
	if scaffold.Registration == nil || !scaffold.Registration.Changed ||
		scaffold.Registration.DryRun || scaffold.Registration.Path != planningConfig.Dir ||
		scaffold.Registration.VerifyID == "" {
		t.Fatalf("scaffold registration = %+v", scaffold.Registration)
	}
	if strings.Contains(scaffold.Registration.Path, filepath.Join("planning", "planning")) {
		t.Fatalf("registered planning root instead of checkout: %+v", scaffold.Registration)
	}

	impl := filepath.Join(parent, "implementation")
	out, errOut, err = runIn(t, parent, "init", "--path", impl,
		"--planning-repo", "../planning", "--json")
	if err != nil {
		t.Fatalf("pointer init: %v\nstdout:\n%s\nstderr:\n%s", err, out, errOut)
	}
	var pointer wire.InitEnvelope
	if err := json.Unmarshal([]byte(out), &pointer); err != nil {
		t.Fatalf("decode pointer receipt: %v\n%s", err, out)
	}
	implConfig, err := config.Discover(impl)
	if err != nil {
		t.Fatal(err)
	}
	if pointer.Registration == nil || pointer.Registration.Path != implConfig.Dir ||
		pointer.Registration.VerifyID != scaffold.Registration.VerifyID {
		t.Fatalf("pointer registration = %+v, scaffold = %+v", pointer.Registration, scaffold.Registration)
	}

	spaces, err := userconfig.Spaces()
	if err != nil || len(spaces) != 2 {
		t.Fatalf("spaces = %+v, err=%v", spaces, err)
	}
	if spaces[0].Path != planningConfig.Dir || spaces[1].Path != implConfig.Dir ||
		spaces[0].VerifyID != spaces[1].VerifyID {
		t.Fatalf("registered entry points = %+v", spaces)
	}

	out, errOut, err = runIn(t, impl, "init", "--path", impl, "--json")
	if err != nil || errOut != "" {
		t.Fatalf("bare pointer re-init: %v\nstdout:\n%s\nstderr:\n%s", err, out, errOut)
	}
	var again wire.InitEnvelope
	if err := json.Unmarshal([]byte(out), &again); err != nil {
		t.Fatal(err)
	}
	if !again.AlreadyInitialized || again.Registration != nil {
		t.Fatalf("bare re-init receipt = %+v", again)
	}
	if spaces, err = userconfig.Spaces(); err != nil || len(spaces) != 2 {
		t.Fatalf("bare re-init changed spaces = %+v, err=%v", spaces, err)
	}
}

func TestInitAutoRegistration_DryRunIsHonestAndWriteFree(t *testing.T) {
	home := t.TempDir()
	t.Setenv(userconfig.DirEnv, home)
	repo := filepath.Join(t.TempDir(), "preview")

	out, errOut, err := runIn(t, filepath.Dir(repo), "init", "--path", repo,
		"--dry-run", "--json")
	if err != nil || errOut != "" {
		t.Fatalf("dry-run init: %v\nstdout:\n%s\nstderr:\n%s", err, out, errOut)
	}
	var envelope wire.InitEnvelope
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("decode dry-run receipt: %v\n%s", err, out)
	}
	if envelope.Registration == nil || !envelope.Registration.Changed ||
		!envelope.Registration.DryRun || envelope.Registration.VerifyID == "" {
		t.Fatalf("dry-run registration = %+v", envelope.Registration)
	}
	if _, err := os.Stat(filepath.Join(repo, config.ConfigFile)); !os.IsNotExist(err) {
		t.Fatalf("dry-run config stat = %v, want not-exist", err)
	}
	if _, err := os.Stat(filepath.Join(home, userconfig.SpacesFile)); !os.IsNotExist(err) {
		t.Fatalf("dry-run registry stat = %v, want not-exist", err)
	}
}

func TestInitAutoRegistration_HumanOutputCarriesOneReceiptLine(t *testing.T) {
	t.Setenv(userconfig.DirEnv, t.TempDir())
	repo := filepath.Join(t.TempDir(), "human-receipt")
	out, errOut, err := runIn(t, filepath.Dir(repo), "init", "--path", repo, "--color=never")
	if err != nil || errOut != "" {
		t.Fatalf("human init: %v\nstdout:\n%s\nstderr:\n%s", err, out, errOut)
	}
	if strings.Count(out, "registered space human-receipt") != 1 {
		t.Fatalf("registration receipt should appear once:\n%s", out)
	}
}

func TestInitAutoRegistration_OptOutsAndExistingInitDoNotWriteRegistry(t *testing.T) {
	home := t.TempDir()
	t.Setenv(userconfig.DirEnv, home)
	parent := t.TempDir()

	flagged := filepath.Join(parent, "flagged")
	if out, errOut, err := runIn(t, parent, "init", "--path", flagged, "--no-register"); err != nil {
		t.Fatalf("--no-register: %v\n%s%s", err, out, errOut)
	}
	// A bare re-run is a topology read. It must not retroactively register a checkout
	// that was intentionally bootstrapped without registration.
	if out, errOut, err := runIn(t, flagged, "init", "--path", flagged); err != nil {
		t.Fatalf("bare existing init: %v\n%s%s", err, out, errOut)
	}
	repair := filepath.Join(parent, "repair")
	if _, err := config.Init(repair, "", false); err != nil {
		t.Fatal(err)
	}
	if out, errOut, err := runIn(t, repair, "init", "--path", repair,
		"--taskflow-root", "."); err != nil {
		t.Fatalf("explicit scaffold repair: %v\n%s%s", err, out, errOut)
	}

	t.Setenv("TSKFLW_NO_REGISTER", "1")
	environment := filepath.Join(parent, "environment")
	if out, errOut, err := runIn(t, parent, "init", "--path", environment); err != nil {
		t.Fatalf("TSKFLW_NO_REGISTER: %v\n%s%s", err, out, errOut)
	}
	spaces, err := userconfig.Spaces()
	if err != nil || len(spaces) != 0 {
		t.Fatalf("opt-outs left spaces = %+v, err=%v", spaces, err)
	}
}

func TestInitAutoRegistration_FailureWarnsWithoutBreakingTopologyOrJSON(t *testing.T) {
	t.Setenv(userconfig.DirEnv, t.TempDir())
	parent := t.TempDir()
	other := filepath.Join(parent, "other")
	if _, err := config.Init(other, "", false); err != nil {
		t.Fatal(err)
	}
	otherConfig, err := config.Discover(other)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := userconfig.AddSpace(userconfig.Space{
		ID: "collision", Path: other, VerifyID: otherConfig.ID,
	}, false); err != nil {
		t.Fatal(err)
	}

	repo := filepath.Join(parent, "collision")
	out, errOut, err := runIn(t, parent, "init", "--path", repo, "--json")
	if err != nil {
		t.Fatalf("registration failure made init fail: %v\n%s%s", err, out, errOut)
	}
	var envelope wire.InitEnvelope
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("warning corrupted JSON stdout: %v\nstdout:\n%s\nstderr:\n%s", err, out, errOut)
	}
	if envelope.Registration != nil {
		t.Fatalf("failed registration produced a receipt: %+v", envelope.Registration)
	}
	if !strings.Contains(errOut, "space registration skipped") ||
		!strings.Contains(errOut, "tskflwctl space add") || !strings.Contains(errOut, "--id my-space") {
		t.Fatalf("stderr lacks warning/remedy:\n%s", errOut)
	}
	if _, err := config.Discover(repo); err != nil {
		t.Fatalf("topology was not initialized: %v", err)
	}
}

func TestInitAutoRegistration_StalePhysicalPathIdentityWarnsInsteadOfClaimingSuccess(t *testing.T) {
	t.Setenv(userconfig.DirEnv, t.TempDir())
	parent := t.TempDir()
	repo := filepath.Join(parent, "checkout")
	if _, _, err := userconfig.AddSpace(userconfig.Space{
		ID: "existing-label", Path: repo, VerifyID: "old-id",
	}, false); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := runIn(t, parent, "init", "--path", repo, "--json")
	if err != nil {
		t.Fatalf("stale-path init: %v\nstdout:\n%s\nstderr:\n%s", err, out, errOut)
	}
	var envelope wire.InitEnvelope
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Registration != nil {
		t.Fatalf("stale identity produced a success receipt: %+v", envelope.Registration)
	}
	if !strings.Contains(errOut, "verify_id") || !strings.Contains(errOut, "space forget") {
		t.Fatalf("stale identity warning lacks repair:\n%s", errOut)
	}
	spaces, err := userconfig.Spaces()
	if err != nil || len(spaces) != 1 {
		t.Fatalf("dedup spaces = %+v, err=%v", spaces, err)
	}
}
