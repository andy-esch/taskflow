package store

// The misfiled concept is gone under the Phase-B flat task layout: there is no
// status directory to disagree with the frontmatter, so tasks are never
// misfiled and lifecycle verbs never relocate files. The tests that lived here
// (misfiled detection, folder-status fallback, misfiled relocation on
// FixFrontmatter/Move, and dir-fallback status healing) no longer describe
// possible behavior and have been removed.
