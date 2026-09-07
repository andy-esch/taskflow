# Touring-bike Thread demo fixture

This self-contained planning space exists for the focused Threads topology recording. It tells one
workshop story: release a steel touring bike only after parallel chassis, drivetrain, wheel, and
loadout work converge in a loaded shakedown ride.

The graph is deliberately small enough to read in one terminal pane while exercising real Thread
semantics:

| Topology | Work state |
| --- | --- |
| Wave 1: frame inspection, rear-hub rebuild, drivetrain renewal | completed foundations |
| Wave 2: wheel truing, bottom-bracket service, rack and bags | eligible, in flight, and blocked |
| Wave 3: loaded shakedown ride | fan-in blocked by unfinished Wave 2 work |
| External gate: front-rack crown mount | outstanding parts dependency, not Thread-owned work |

`assets/vhs/threads.tape` records only against this fixture. Keep its dates and prose deterministic,
its dependencies mechanically credible, and the space lint-clean.

`shared-safety-review.compose.yml` is a release-dogfood input for a throwaway copy of this space. It
creates a second Thread sharing three members, retains the rear hub as nonmember graph context, and
adds one plausible frame-inspection → wheel-truing edge; never apply it to the committed fixture.
