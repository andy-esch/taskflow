# Stage a throwaway multi-space registry for atlas.tape.
#
# Every other demo tape records against ONE planning tree, so it can just `cd`
# into assets/demo-planning. The atlas has nothing to show with one space: it
# needs several registered entry points, at least two of which share a planning
# identity, plus a home registry to read them from. That is more setup than fits
# on a tape's one-line Hide block, so it lives here.
#
# Two isolation rules this script exists to guarantee:
#
#   1. It NEVER touches the recorder's real ~/.config/tskflwctl/spaces.toml.
#      TSKFLW_CONFIG_HOME redirects the whole home config into the staging dir,
#      so `space add` writes there and nowhere else.
#   2. It NEVER dirties the committed fixtures. Everything is a copy under
#      $STAGE, following the same throwaway-copy pattern as picker.tape.
#
# The directories are renamed on copy so the paths the atlas prints in dim text
# read as the space you are looking at ("/private/tmp/tskflw-atlas/bike-workshop")
# rather than as fixture plumbing ("…/assets/demo-planning"). The pointer's
# relative planning_repo is rewritten to match.
#
# SOURCE this from the repo root, with ./bin on PATH (which is what `just gifs`
# does) — it has to change the caller's cwd and environment, which a subprocess
# could not do. Because it is sourced, strict mode is deliberately confined to
# the staging subshell: a `set -e` left behind in the recording shell would kill
# the take the first time any later command returned non-zero.

STAGE="${STAGE:-/private/tmp/tskflw-atlas}"

(
  set -euo pipefail

  rm -rf "$STAGE"
  mkdir -p "$STAGE/home"
  export TSKFLW_CONFIG_HOME="$STAGE/home"

  # bike-workshop: the direct planning checkout.
  cp -R assets/demo-planning "$STAGE/bike-workshop"
  # bike-shop: an impl repo POINTING at that same tree — the second entry point
  # into one logical space, which is what `h`/`l` on an atlas card selects between.
  cp -R assets/demo-bike-shop "$STAGE/bike-shop"
  # kitchen: a second, unrelated planning identity, so the atlas has more than one card.
  cp -R assets/demo-kitchen "$STAGE/kitchen"

  # Re-point the pointer at the renamed sibling. planning_repo is stored relative,
  # and `config migrate`/`init` deliberately preserve that spelling, so this is a
  # one-key edit rather than a re-init.
  python3 - "$STAGE/bike-shop/.tskflwctl.toml" <<'PY'
import pathlib, re, sys
p = pathlib.Path(sys.argv[1])
p.write_text(re.sub(r'(planning_repo\s*=\s*")\.\./demo-planning(")', r'\1../bike-workshop\2', p.read_text()))
PY

  # Register all three. The --id is the machine-local label the atlas card shows,
  # so it is chosen for the demo rather than inherited from the directory name.
  tskflwctl space add "$STAGE/bike-workshop" --id bike-workshop >/dev/null
  tskflwctl space add "$STAGE/bike-shop" --id bike-shop >/dev/null
  tskflwctl space add "$STAGE/kitchen" --id kitchen >/dev/null
) || return 1

# The two things only a sourced script can hand back: the isolated registry, and
# a directory with NO planning repo above it. That second one is the case `ui`
# answers by landing on the atlas, so the demo opens there without spending a
# keystroke getting to it.
export TSKFLW_CONFIG_HOME="$STAGE/home"
cd "$STAGE"
