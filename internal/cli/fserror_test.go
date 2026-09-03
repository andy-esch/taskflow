package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// readOnlyPlanningRepo builds a planning tree whose tasks/ directory cannot be
// written, so a create reaches a real permission error rather than a simulated one.
func readOnlyPlanningRepo(t *testing.T) string {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("running as root: mode bits do not deny writes")
	}
	root := t.TempDir()
	tasks := filepath.Join(root, "tasks")
	if err := os.MkdirAll(tasks, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "epics"), 0o755); err != nil {
		t.Fatal(err)
	}
	epic := "---\nstatus: active\npriority: high\ndescription: e\n---\n# E\n"
	if err := os.WriteFile(filepath.Join(root, "epics", "01-e.md"), []byte(epic), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(tasks, 0o500); err != nil {
		t.Fatal(err)
	}
	// Restore write permission so t.TempDir's cleanup can remove the tree.
	t.Cleanup(func() { _ = os.Chmod(tasks, 0o700) })
	return root
}

// The finding's own scenario: a routed mutation failed while creating its file with
// "operation not permitted", and the correct next action was to grant access — but
// the envelope said only {"code":"error","message":"..."}, so an agent had to read
// English to tell permission-denied from missing-directory from disk-full.
func TestJSONError_CarriesFilesystemRecoveryDetails(t *testing.T) {
	root := readOnlyPlanningRepo(t)

	_, err := runRootStreams(t, "-C", root, "--json", "task", "new", "Probe", "--epic", "01-e", "--tags", "t")
	if err == nil {
		t.Fatal("writing into a read-only directory should fail")
	}
	// Execute returns the error; main renders it through WriteError. Render it the
	// same way, so the assertion is about the envelope an agent actually receives.
	var buf bytes.Buffer
	WriteError(&buf, err, true)
	envelope := buf.String()

	var env struct {
		Error struct {
			Code       string `json:"code"`
			Message    string `json:"message"`
			Filesystem *struct {
				Class     string `json:"class"`
				Operation string `json:"operation"`
				Path      string `json:"path"`
				Retryable bool   `json:"retryable"`
			} `json:"filesystem"`
		} `json:"error"`
	}
	if jsonErr := json.Unmarshal([]byte(envelope), &env); jsonErr != nil {
		t.Fatalf("invalid error envelope: %v\n%s", jsonErr, envelope)
	}
	if env.Error.Filesystem == nil {
		t.Fatalf("an OS failure should carry filesystem details:\n%s", envelope)
	}
	fsDetail := env.Error.Filesystem
	if fsDetail.Class != "permission" {
		t.Errorf("class = %q, want permission", fsDetail.Class)
	}
	if fsDetail.Retryable {
		t.Error("a permission failure needs access granted, not a retry")
	}
	if fsDetail.Operation == "" || fsDetail.Path == "" {
		t.Errorf("operation and path should be named: %+v", fsDetail)
	}
	if !strings.Contains(fsDetail.Path, "tasks") {
		t.Errorf("path should name where the write was headed: %q", fsDetail.Path)
	}
	// The human message stays, and the exit-code policy is unchanged.
	if env.Error.Message == "" {
		t.Error("the prose message must be preserved alongside the details")
	}
}

// The block is for filesystem failures only: a domain error already classifies into a
// stable code, and attaching an empty filesystem object to it would be noise an agent
// has to branch on.
func TestJSONError_DomainFailureHasNoFilesystemBlock(t *testing.T) {
	root := setupRepo(t)

	_, err := runRootStreams(t, "-C", root, "--json", "task", "show", "no-such-task")
	if err == nil {
		t.Fatal("an unknown task should fail")
	}
	var buf bytes.Buffer
	WriteError(&buf, err, true)
	if !strings.Contains(buf.String(), "not-found") {
		t.Fatalf("setup: expected a classified domain envelope, got:\n%s", buf.String())
	}
	if strings.Contains(buf.String(), "filesystem") {
		t.Errorf("a domain failure should carry no filesystem block:\n%s", buf.String())
	}
	if got := ExitCode(err); got != 10 {
		t.Errorf("exit = %d, want 10 (not found) — the code policy is unchanged", got)
	}
}

// Human output is untouched: this finding is about the machine envelope.
func TestJSONError_HumanOutputUnchanged(t *testing.T) {
	root := readOnlyPlanningRepo(t)

	res, err := runRootStreams(t, "-C", root, "task", "new", "Probe", "--epic", "01-e", "--tags", "t")
	if err == nil {
		t.Fatal("writing into a read-only directory should fail")
	}
	if strings.Contains(res.Merged, "class") || strings.Contains(res.Merged, "retryable") {
		t.Errorf("the human path should not gain envelope fields:\n%s", res.Merged)
	}
}

func TestFilesystemDetails_ClassifiesTheRecoveryRelevantCases(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		class     string
		retryable bool
	}{
		{"permission", &fs.PathError{Op: "open", Path: "/p", Err: fs.ErrPermission}, "permission", false},
		{"not found", &fs.PathError{Op: "stat", Path: "/p", Err: fs.ErrNotExist}, "not-found", false},
		{"read only", &fs.PathError{Op: "open", Path: "/p", Err: syscall.EROFS}, "read-only", false},
		{"no space", &fs.PathError{Op: "write", Path: "/p", Err: syscall.ENOSPC}, "no-space", false},
		{"transient", &fs.PathError{Op: "open", Path: "/p", Err: syscall.EINTR}, "io", true},
		{"other io", &fs.PathError{Op: "open", Path: "/p", Err: syscall.EIO}, "io", false},
	}
	for _, tc := range cases {
		got := filesystemDetails(tc.err)
		if got == nil {
			t.Errorf("%s: expected details", tc.name)
			continue
		}
		if got.Class != tc.class || got.Retryable != tc.retryable {
			t.Errorf("%s: class=%q retryable=%v, want %q/%v", tc.name, got.Class, got.Retryable, tc.class, tc.retryable)
		}
	}
	// A rename carries two paths; the destination is the one a caller acts on.
	link := filesystemDetails(&os.LinkError{Op: "rename", Old: "/from", New: "/to", Err: fs.ErrPermission})
	if link == nil || link.Path != "/to" {
		t.Errorf("a link error should report its destination: %+v", link)
	}
	// Non-filesystem errors opt out entirely.
	for _, err := range []error{domain.ErrValidation, errors.New("plain"), nil} {
		if got := filesystemDetails(err); got != nil {
			t.Errorf("%v should carry no filesystem details, got %+v", err, got)
		}
	}
}
