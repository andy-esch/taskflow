package store

// Path resolvers: slug/id → absolute file path WITHOUT reading or parsing the file.
// The `<entity> path` commands use these so they still work on a file whose
// frontmatter won't parse — the case where you most need the path (to open and
// repair it). GetTask/GetEpic/GetAudit would fail at the parse step first.

// ResolveTaskPath returns a task's file path from its slug/id, parse-free.
func (s *FS) ResolveTaskPath(slug string) (string, error) { return s.resolve(slug) }

// ResolveEpicPath returns an epic's file path from its id, parse-free.
func (s *FS) ResolveEpicPath(id string) (string, error) { return s.resolveEpicPath(id) }

// ResolveAuditPath returns an audit's file path from its slug/id, parse-free.
func (s *FS) ResolveAuditPath(slug string) (string, error) { return s.resolveAudit(slug) }
