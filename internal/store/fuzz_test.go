package store

import (
	"bytes"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

// These fuzz targets feed random bytes into the hand-rolled byte parsers, which
// are the panic-prone surface (index math on untrusted input). Beyond no-panic,
// the split/update targets assert the load-bearing invariants the byte parsers
// risk corrupting (body preserved verbatim, output re-parses) — see CLAUDE.md's
// surgical-frontmatter rule. The seed corpus also runs under a normal `go test`.

func seeds(f *testing.F) {
	for _, s := range []string{
		"---\nstatus: open\n---\nbody",
		"---\r\nstatus: open\r\n---\r\nbody\r\n",
		"---\n---\n",
		"---\nunterminated frontmatter",
		"no frontmatter at all",
		"---",
		"---\n",
		"---\ntags: a,b,c\ndescription: x: y\n---\n",
		"---\n<<<<<<< HEAD\nstatus: open\n=======\nstatus: done\n>>>>>>> b\n---\n",
		"",
	} {
		f.Add([]byte(s))
	}
}

func FuzzSplitFrontmatter(f *testing.F) {
	seeds(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		_, body := splitFrontmatter(data)
		// The body is the unmodified tail of the input — the split must never lose
		// or rewrite body bytes (a corruption class the offset math could introduce).
		if !bytes.HasSuffix(data, body) {
			t.Fatalf("split body is not a verbatim suffix of the input")
		}
	})
}

func FuzzUpdateFrontmatter(f *testing.F) {
	seeds(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		out, err := updateFrontmatter(data, map[string]any{"status": "in-progress"})
		if err != nil {
			return // the parser legitimately rejected this input — nothing to assert
		}
		// Only hold the surgical-preservation invariants for inputs whose frontmatter
		// parses cleanly into a typed map. Messy-but-node-parseable inputs (e.g.
		// duplicate keys) are the caller's job to reject (SetFields' parse-before-
		// commit), not updateFrontmatter's, so asserting on them would be a false alarm.
		fmIn, bodyIn, serr := splitFrontmatterStrict(data)
		var in map[string]any
		if serr != nil || yaml.Unmarshal(fmIn, &in) != nil {
			return
		}
		// 1. Body preserved verbatim (assembleFile writes it last, untouched).
		if !bytes.HasSuffix(out, bodyIn) {
			t.Fatalf("update did not preserve the body verbatim")
		}
		// 2. The output still splits and its frontmatter re-parses as valid YAML.
		fmOut, _, oerr := splitFrontmatterStrict(out)
		if oerr != nil {
			t.Fatalf("output no longer splits: %v", oerr)
		}
		var got map[string]any
		if err := yaml.Unmarshal(fmOut, &got); err != nil {
			t.Fatalf("output frontmatter is no longer valid YAML: %v", err)
		}
		// 3. The requested key now holds the new value.
		if got["status"] != "in-progress" {
			t.Fatalf("status not applied: got %v", got["status"])
		}
	})
}

func FuzzFixFrontmatterText(f *testing.F) {
	seeds(f)
	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _ = fixFrontmatterText(data)
	})
}

func FuzzFrontmatterError(f *testing.F) {
	seeds(f)
	f.Fuzz(func(_ *testing.T, data []byte) {
		_ = frontmatterError(data, errBadFrontmatter)
	})
}
