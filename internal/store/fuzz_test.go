package store

import "testing"

// These fuzz targets feed random bytes into the hand-rolled byte parsers, which
// are the panic-prone surface (index math on untrusted input). They assert no
// panic; the seed corpus also runs under a normal `go test`.

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
	f.Fuzz(func(_ *testing.T, data []byte) {
		fm, body := splitFrontmatter(data)
		_, _ = fm, body
	})
}

func FuzzUpdateFrontmatter(f *testing.F) {
	seeds(f)
	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _ = updateFrontmatter(data, map[string]any{"status": "in-progress"})
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
