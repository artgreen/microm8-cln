# Coverage Plan: drive valid testable code to ≥70%

**Goal.** Reach ≥70% statement coverage on the *testable* surface of the
codebase, exercising both happy and failure paths, so regressions in the
modernization migration (and going forward) are caught locally by
`tools/scripts/check.sh`.

This is a multi-phase effort. Today (after Phase 7) only 17 packages are
measured, totalling 8.8% of *their* statements. The codebase has ~213K
non-quarantined Go LOC and ~108 packages. Hitting 70% on the full
testable surface is realistically **10–16 weeks of focused work**. Each
phase below lands as one PR (or a small cluster) so progress is
visible and the coverage ratchet only ever goes up.

---

## 1. Scope: what counts as "valid testable code"

The denominator is **production code that can plausibly be exercised by
a deterministic Go test on a headless CI box**. Everything else is out
of scope for the 70% target — not because it doesn't matter, but
because pretending to test it via Go's `cover` tool produces dishonest
numbers.

**IN scope** (measured, must hit 70%):

- Pure logic packages: `utils`, `runestring`, `buffer`, `panic`,
  `encoding/*`, `filerecord`, `fmt`, `log`, `core/types`,
  `core/vduproto`, `core/vduconst`, `core/settings`, `core/exception`,
  `internal/*`.
- Format parsers/decoders: `disk`, `core/hardware/apple2/woz`,
  `core/hardware/apple2/woz2`, `core/hires`, `files`,
  `decoding/ogg`, `decoding/wav`, `freeze`.
- CPUs & emulation core: `core/hardware/cpu/mos6502`,
  `core/hardware/cpu/z80`, `core/hardware/cpu/z80` *handwritten only*
  (gen files excluded — see below), `core/memory`,
  `core/hardware/apple2` (cards), `core/hardware/common`,
  `core/hardware/spectrum`, `core/hardware/servicebus`,
  `core/hardware/control`, `core/hardware/apple2helpers`,
  `core/hardware/buzzer`, `core/instrument`.
- Language layer: `core/dialect`, `core/dialect/applesoft`,
  `core/dialect/appleinteger`, `core/dialect/logo`, `core/dialect/shell`,
  `core/dialect/plus`, `core/dialect/parcel`, `core/interpreter`,
  `core/editor`, `octalyzer/tokenizer`.
- Audio/music: `restalgia`, `restalgia/control`,
  `core/hardware/restalgia`, `microtracker`, `microtracker/tracker`,
  `microtracker/mock`, `core/buzzer`.
- Networking & state: `api`, `update`, `ducktape`, `ducktape/client`,
  `ducktape/server`, `fastserv`, `fastserv/client`, `postoffice`,
  `postoffice/core`, `postoffice/client`, `core/syncmanager`,
  `filerecord`.
- Glue and tooling: `core/types/glmath`, `accelimage`,
  `core/hardware/cpu/mos6502/asm`, `octalyzer/tokenizer`,
  `octalyzer/backend`, `octalyzer/errorreport`.

**OUT of scope** (excluded from the coverage denominator, with reason):

| Package / file pattern | Reason |
|---|---|
| `gl/` (~20K LOC) | Generated CGO bindings to OpenGL; no logic to test |
| `glumby/` (~2.7K LOC) | GLFW + OpenGL window/texture wrappers; require display |
| `octalyzer/video/` (~5.4K LOC) | Renderer; requires GL context |
| `octalyzer/assets/` (~3.9K LOC) | Generated embedded assets |
| `octalyzer/main.go`, `main_remint.go` | CLI + window wiring; tested indirectly via the AST policy test |
| `octalyzer/clientperipherals/` | Joystick/audio device drivers; OS-dependent |
| `*_gen.go` files (z80/opcodes_*_gen.go ~12.5K LOC; restalgia tables; etc.) | Generated code, by definition no hand-written logic |
| `restalgia/driver/` | Audio output device; can't test without an audio device |
| `octalyzer/ui/`, `octalyzer/ui/chat/`, `octalyzer/ui/forumtool/` | UI controllers; smoke-tested manually |
| `octalyzer/hamburger_*`, `octalyzer/launcher.go` | UI menu wiring |
| `_*/` (already quarantined) | Pre-existing exclusion |
| `*/cmd/*` small main packages | Wrappers around the libraries they import |
| `glumby/demo/` | Demo program |

**Mechanism.** `.ci/coverage-exclude.txt` (new in Phase 8) holds the
exclusion list as one path-or-glob per line with comments. The coverage
runner skips matching packages when computing the aggregate %.

After exclusions the testable surface is roughly **~145K LOC across
~75 packages** — that's the denominator. 70% of that is the goal.

---

## 2. Strategy

1. **Define and freeze the testable surface** (Phase 8). Make the
   denominator explicit, build the measurement infra around it, and
   record the baseline. Without this we'd be aiming at a moving target.
2. **Domain-grouped sweeps**, ordered by **regression-detection ROI**:
   touch code we recently modernized first (highest "did I break it?"
   risk), then move outward to the emulator core. Phase 9 onward.
3. **Coverage ratchet rises with every merge.** `.ci/coverage-thresholds.txt`
   already enforces per-package floors; each PR raises the floor for
   packages it touched. Coverage can never regress without an explicit
   threshold lowering, which would need to be argued for in the PR.

**Per-phase contract:** every phase PR must

- Drive its target packages to **≥70% (small packages: ≥90%)** with
  table-driven tests for happy paths AND deliberate failure paths.
- Add **fuzz tests** for any parser / decoder / wire-format code touched
  (the existing `base91` / `ffpak` / `mempak` fuzzers are the model).
- Add **goroutine-leak assertions** (`testutil.NoGoroutineLeaks`) on
  any test that spawns goroutines.
- Use **`t.Parallel()`** in outer and subtests where state isn't shared.
- Update `.ci/coverage-thresholds.txt` with the new floors so the
  ratchet locks in.
- **Find real bugs.** If a phase produces zero bug fixes, the tests are
  probably too superficial — re-examine.

---

## 3. Phase 8 — scope & measurement infrastructure (~1 day)

One PR. No tests yet — this defines what we're measuring.

**Deliverables:**

- `.ci/coverage-exclude.txt` — the exclusion list from §1, with reason
  per entry.
- `.ci/testable-packages.txt` — the inclusion list (the denominator),
  auto-generated as `all-non-quarantined ∖ excluded`. Regenerated by a
  small `tools/scripts/refresh-testable.sh`.
- `tools/scripts/test.sh cover` updated to:
  - Run `go test -cover ./...` against the testable set (not just the
    allowlist).
  - Report per-package coverage + aggregate-against-testable.
  - Surface the gap to the 70% target.
- Initial `.ci/coverage-thresholds.txt` entries for every package in
  the testable set, set at **current coverage** (most will be `0.0`).
- Replace `.ci/test-allowlist.txt` with `.ci/testable-packages.txt`
  OR — preferred — keep both, with the test allowlist being the
  must-pass set and testable-packages being the measurement set.
- `COVERAGE-PLAN.md` (this file) lands so progress is tracked.

**Acceptance:** `tools/scripts/test.sh cover` prints
"`Coverage X.X% of testable surface (target 70.0%)`" and the threshold
ratchet enforces per-package floors.

---

## 4. Execution phases, ordered by regression-detection ROI

Each phase below is **one PR (or a small cluster for the largest
packages)**, drives its packages to ≥70% (or ≥90% if small), and
ratchets the threshold. The rough effort column is "engineer-days with
focus" — half-time work easily 2× these.

### Phase 9 — pure helpers (already partly there) — ~3 days

Touched during Phases 1–7; high "did I break this?" risk; mostly
single-file packages where 70% is a few hundred lines of tests.

| Package | LOC | Now | Target | Techniques |
|---|---:|---:|---:|---|
| `utils` | 601 | 13.6% | 90% | Table-driven; fuzz `StrToFloat*`, `IntToStr`, `Pos`, `Copy`, `Delete`, the (X)Z/GZIP helpers; golden for `FloatToStrApple`/`StrToFloatStrApple*` |
| `runestring` | 315 | 96.3% | 96% | Already at target; just ratchet |
| `buffer` | 193 | 93.1% | 93% | Already; ratchet |
| `panic` | 79 | 100% | 100% | Already; ratchet |
| `fmt` | 90 | 11.8% | 90% | Capture stdout to test the Print/Println/Printf paths under Verbose on/off |
| `log` | 152 | 20.6% | 80% | Capture output, exercise formatStr edge cases (`fmt.Sprint(v...)` regression is the obvious one) |
| `filerecord` | 431 | 47.6% | 80% | Round-trip JSON/BSON, edge cases for `AddMeta` (we fixed a bug there), file-record validation paths |
| `encoding/base91` | 216 | 100% | 100% | Already |
| `encoding/ffpak` | 155 | 88% | 90% | Drive a few edge cases; ratchet |
| `encoding/mempak` | 680 | 72% | 85% | Add the missing branches; existing fuzz coverage extends |
| `encoding/octadec` | 60 | 0% | 80% | New round-trip + fuzz |
| `internal/lifecycle` | 121 | 100% | 100% | Already |
| `internal/testutil` | 119 | n/a | 80% | The helpers themselves deserve tests — leak-detector golden cases, eventually-helper races |

### Phase 10 — core types & settings (foundation) — ~3 days

Foundational. Bugs here propagate everywhere. We just touched
`Algorithm` and `LoopState` for the lock-by-value cluster, so the
"did I break the parser?" risk is high.

| Package | LOC | Now | Target | Techniques |
|---|---:|---:|---:|---|
| `core/types` | 10227 | 13.2% | 70% | Table-driven for Token, TokenList, TokenType.String; Algorithm operations under -race; FuncPtr5b serialisation; Turtle math; VideoColor sRGB ⇄ Lab round-trip (the new constants!); CodeRef arithmetic |
| `core/vduproto` | 1181 | 11.3% | 80% | Struct marshalling, message ID dispatch happy + malformed-payload paths |
| `core/vduconst` | 245 | 0% | 95% | Single-purpose constants — just exhaustively test the lookup tables |
| `core/settings` | 829 | 0% | 70% | Settings load/save, video-palette-zone math, speaker-redirect routing (we touched this in plus_vmspeaker) |
| `core/exception` | 5 | 0% | 100% | Trivial |
| `core/types/glmath` | 1027 | 0% | 85% | Pure math — vec/mat ops, golden for transforms |

### Phase 11 — files / network / api (recently modernized; highest "did the migration break me?" risk) — ~1 week

We rewrote error handling here in Phase 6 and lifecycle in Phase 7.
This is the most likely place for the migration to have introduced a
behaviour change we don't yet detect.

| Package | LOC | Now | Target | Techniques |
|---|---:|---:|---:|---|
| `files` | 5009 | 0% | 70% | In-memory file provider tests; round-trip via memory/package/zip providers; the duplicate-condition fix in `packagefileprovider` deserves a regression test |
| `api` | 2587 | 3% | 75% | **In-memory DuckTape server** fake — exercise Login/Register/ChangePassword/etc.; assert `errors.Is(err, ErrTimeout)` paths; the Phase-7 monitor lifecycle covers a corner; fuzz the message payload parsers |
| `update` | 371 | 10% | 85% | Already-isolated `Disabled` paths; add tests for the live path using `httptest.Server` (CheckVersion, GetChecksum, partial DownloadVersion to a tempdir) |
| `ducktape` | (200ish) | 0% | 80% | Wire-format round-trip (`ducktape.DuckTapeBundle.MarshalBinary` / parse); timeout vs framing-error distinction; fuzz the bundle parser — there's a documented `strings.Contains(err.Error(), "timeout")` consumer we should not break |
| `ducktape/client`, `ducktape/server` | 836 | 0% | 70% | Loopback test (server + client over `net.Pipe()`); confirm the Phase-5 `return`/`OK=false` shutdown fixes survive |
| `fastserv`, `fastserv/client` | 659 | 0% | 70% | Same loopback strategy |

### Phase 12 — interpreter (huge blast radius) — ~3 weeks (split into 3 PRs)

`core/interpreter` is 7058 LOC of language semantics. Every dialect
sits on top of it. Worth its own phase split.

**Phase 12a — VM lifecycle, code execution, scoping.** ~1 week.
Build a tiny test harness that constructs an Interpreter without the
full hardware stack (we already proved this is feasible with the
Phase-7 Producer tests). Cover NewInterpreter, Push/Pop scope, label
table, code-ref arithmetic, error propagation.

**Phase 12b — Token stream + expression evaluator.** ~1 week.
SplitOnTokenWithBrackets (we fixed this for SA6005), expression
parse, function-call dispatch, type coercion. Heavy fuzz on the
tokenizer.

**Phase 12c — Control flow + GOSUB/RETURN, FOR/NEXT, GOTO.** ~1 week.
This is where the Phase 4b lock-by-value cluster lived. Tests:
nested FOR, GOSUB chains, abort semantics, exception unwind.

| Package | LOC | Now | Target | Techniques |
|---|---:|---:|---:|---|
| `core/interpreter` | 7058 | 0% | 70% | See sub-phases above |
| `octalyzer/tokenizer` | 703 | 0% | 80% | Pure lexer — table-driven happy/fail + fuzz |
| `core/interfaces` | 961 | 0% | 60%¹ | Mostly interface declarations; coverage will be low but stable |

¹ For interface-heavy packages a 70% target is unfair — the bodies are
mostly elsewhere. The threshold should reflect the testable subset.

### Phase 13 — dialects (frequently-touched, complex semantics) — ~3 weeks (split per dialect)

We touched DIM, RENUMBER, the Algorithm method receivers, and the
break-in-switch dialect cases. Each dialect deserves its own PR.

**Phase 13a — `core/dialect` common machinery.** ~3 days.
Decolon, Renumber, command/function registration, scope handling,
watch-vars, the OneOfTokens helper (we touched the SA6005 sites
here). Build a tiny dialect harness in `internal/testutil/dialect/`.

**Phase 13b — Applesoft BASIC.** ~1 week.
Build a corpus of short, deterministic BASIC programs with expected
output. Drive each command (HOME, PRINT, INPUT, GOTO/GOSUB, FOR/NEXT,
IF/THEN, DIM, DEF FN, …) end-to-end. Capture VDU output via golden
files. Targeted regression tests for: DIM type-suffix handling
(the SA4011 break-in-switch fix), implicit assignment fall-throughs,
RENUMBER, the lock-by-value-fixed Algorithm.

**Phase 13c — Integer BASIC + Logo + shell + plus.** ~1.5 weeks.
Same harness; smaller corpora per dialect. Logo turtle commands have
turtle-coordinate goldens. Shell has env-var + command-history tests.
Plus has the SETCLASSICCPU/SETSPEAKER-redirect tests (we touched
`plus_vmspeaker.go`).

| Package | LOC | Now | Target |
|---|---:|---:|---:|
| `core/dialect` | 2556 | 0% | 70% |
| `core/dialect/applesoft` | 10513 | 0% | 70% |
| `core/dialect/appleinteger` | 3141 | 0% | 70% |
| `core/dialect/logo` | 13883 | 0% | 70% |
| `core/dialect/shell` | 1875 | 0% | 70% |
| `core/dialect/plus` | 16050 | 0% | 70% |
| `core/dialect/parcel` | 499 | 0% | 75% |

### Phase 14 — memory + CPUs — ~2 weeks (split)

**Phase 14a — `core/memory`.** ~3 days.
MemoryMap, MemoryBlock, banking, the BlockMapper. Synthetic memory
fixtures. Heavy on edge cases around state transitions ("off"/"on",
ROM vs RAM, expansion ROM swap-in).

**Phase 14b — `core/hardware/cpu/mos6502`.** ~1 week.
Port the Klaus-Dormann 6502 functional tests if license-compatible
(decimal mode quirks, every documented opcode). Add tests for the
specific opcode-decode paths the emulator implements. Tests for
clock/cycle accounting (the Phase 4b lock-by-value cluster lives near
here).

**Phase 14c — `core/hardware/cpu/z80`.** ~3 days.
Same approach. Many `_gen.go` files (excluded from coverage); test
the hand-written `z80.go` + `port.go` + `memory.go`.

| Package | LOC (hand-written) | Now | Target |
|---|---:|---:|---:|
| `core/memory` | 3732 | 0% | 75% |
| `core/hardware/cpu/mos6502` | 3397 | 0% | 75% |
| `core/hardware/cpu/mos6502/asm` | 831 | 0% | 80% |
| `core/hardware/cpu/z80` | ~600 hand-written | 0% | 75% |

### Phase 15 — disk formats — ~1 week

Pure file-format code; deterministic; perfect for golden + fuzz tests.

| Package | LOC | Now | Target | Techniques |
|---|---:|---:|---:|---|
| `disk` | 4729 | 0% | 80% | Golden round-trips for `.dsk` / `.po` / `.2mg` / `.nib`; AppleDOS catalog parse fuzz; ProDOS volume traversal; the EqualFold sites we fixed deserve regression tests |
| `core/hardware/apple2/woz` | 1087 | 0% | 80% | WOZ1 file-format round-trip; the `mMin==mMin` bug we fixed in `WriteBit` deserves a property test |
| `core/hardware/apple2/woz2` | 594 | 0% | 80% | WOZ2 round-trip; same property test |
| `core/hires` | 2639 | 0% | 70% | Hi-res pixel layout; golden bitmaps |

### Phase 16 — hardware cards & helpers — ~2 weeks

The lock-by-value fixes (`mos6551StatusRegister`, `FuncPtr5b`), the
Mockingboard chip-2 base-address fix, and the IOCard ServiceBus
machinery all live here.

| Package | LOC | Now | Target | Techniques |
|---|---:|---:|---:|---|
| `core/hardware/apple2` | 9592 | 0% | 70% | Per-card tests with synthetic memory maps; the Mockingboard fix gets a "chip 0 ≠ chip 1 base" assertion; DiskII drive-state state-machine tests; SmartPort packet round-trip |
| `core/hardware/common` | 5106 | 0% | 75% | mos6551 register flow under -race (lock-by-value regression test!); virtualmodem command dispatch (we removed an empty branch); AY-3-8910 register snapshots |
| `core/hardware/apple2helpers` | 4662 | 0% | 70% | Text mode helpers, colour-zone math, palette transforms |
| `core/hardware/spectrum` | 2737 | 0% | 65% | ZX Spectrum hardware; lower target because less in-scope going forward |
| `core/hardware/servicebus`, `core/hardware/control` | 794 | 0% | 80% | Pure machinery |
| `core/hardware/buzzer` | 75 | 0% | 90% | Trivial |

### Phase 17 — audio synth + music + state — ~1.5 weeks

| Package | LOC | Now | Target |
|---|---:|---:|---:|
| `restalgia` | 4522 | 0% | 70% |
| `restalgia/control` | 356 | 0% | 80% |
| `core/hardware/restalgia` | 210 | 0% | 80% |
| `microtracker`, `microtracker/tracker`, `microtracker/mock` | 4224 | 2.8% | 70% |
| `freeze` | 290 | 0% | 80% |
| `core/syncmanager` | 54 | 0% | 90% |
| `core/buzzer` | 75 | 0% | 90% |
| `core/instrument` | 119 | 0% | 80% |

### Phase 18 — long tail — ~1 week

Everything not yet at threshold. Includes the editor (3527 LOC),
debugger (2724 + 322), presentation (926), postoffice (207+284+72),
small encoders (`octadec`), decoding/ogg/wav helpers, `accelimage`.
Some of these may move OUT of the testable set in this phase if
inspection reveals they're rendering-coupled.

---

## 5. Tactical playbook (apply per package)

1. **Read the package's exports and major code paths.** Note error
   types, edge cases, internal state. 30 minutes.
2. **Pick a black-box vs white-box split.** Default: black-box
   (`package x_test`). Use white-box only when you need to inspect
   unexported state.
3. **Write the table-driven happy-path test.** One row per documented
   behaviour. Use `t.Parallel()`.
4. **Add deliberate failure rows.** For each input space: empty, nil,
   max-size, malformed, out-of-range, concurrent. Aim for ~50/50
   happy/fail by row count.
5. **For parsers/decoders/wire formats: add `FuzzX` with a small seed
   corpus** (existing inputs from happy-path tests + a few
   pathological cases). The fuzzer should never crash.
6. **For rendered/formatted output: golden files.** Store under
   `testdata/`. Use `internal/testutil.Golden`. Update via
   `go test ./pkg -update`.
7. **For goroutine-spawning code: defer
   `testutil.NoGoroutineLeaks(t)()`.**
8. **Run with `-race -shuffle=on -count=1`** before opening the PR.
   Flakes uncovered now save days later.
9. **Run `go test -cover` and update
   `.ci/coverage-thresholds.txt`** to the new floor.
10. **Look for bugs.** A coverage sweep that produces no bug fixes is
    a coverage sweep with shallow assertions. Real tests find real
    bugs — the modernization phases have caught one bug every ~50
    LOC of test on average.

### Fixtures shared across phases

- `internal/testutil/dialect` — tiny harness that builds an
  Interpreter + dialect against a no-op memory map (Phase 12a
  prerequisite).
- `internal/testutil/disk` — `.dsk` / `.po` / `.woz` builder helpers
  (Phase 15 prerequisite).
- `internal/testutil/network` — in-memory DuckTape server (Phase 11
  prerequisite).

These ship with the first phase that needs them, then get reused.

---

## 6. Anti-goals (won't test, won't measure)

- OpenGL rendering, font rasterisation, texture upload. Requires a
  display.
- Audio output devices. Requires a sound card and is
  acoustically-validated, not unit-validated.
- Joystick / keyboard hardware input. OS-dependent.
- The `main` window, splash screen, hamburger menus. Smoke-tested by
  hand.
- Generated code (`*_gen.go`). By definition no logic of its own.
- Code in `_quarantine` directories. Already excluded.

These are listed in `.ci/coverage-exclude.txt` and their LOC drops out
of the denominator so we're measuring what we're actually testing.

---

## 7. Risk factors

1. **The dialects pull in `core/hardware/apple2helpers` for
   text-mode output.** That couples Phase 13 to a piece of Phase 16.
   Mitigation: build a no-op VDU stub in the dialect test harness so
   dialects can be tested without the full text layer.
2. **The interpreter constructs the full hardware stack on init**
   (verified in Phase 7 — the Producer tests had to work around it).
   Mitigation: extract a minimal-init constructor in Phase 12a; this
   is itself a refactor and may grow the phase.
3. **Klaus-Dormann 6502 tests are AGPL-equivalent.** May not be
   embeddable. Mitigation: write our own opcode-by-opcode tests using
   the documented MOS 6502 reference; slower but unencumbered.
4. **The `update` package live tests require an HTTP fake.** Easy
   with `httptest.Server` but adds a dep dance. Mitigation:
   `testdata/update/version.txt` etc.
5. **Coverage % over-rewards trivial code, under-rewards branching
   code.** A 90%-covered constants file is less valuable than a
   60%-covered state machine. Mitigation: when reviewing a phase PR,
   read the *uncovered* lines, not the cover %. The number is a
   ratchet, not a quality bar.
6. **Some packages will fight 70% because they're mostly delegation.**
   E.g. `core/interfaces` is interface declarations. The plan sets
   them lower (60–65%) explicitly.

---

## 8. Tracking

This file gets a status table appended at the bottom of each merged
phase PR:

| Phase | PR | Coverage before | Coverage after | Bugs found |
|---|---|---|---|---|
| Phase 8 | #? | — (baseline only) | (baseline) | — |
| Phase 9 | #? | x.x% | y.y% | n |
| … | | | | |

Aggregate target: **≥70% of testable-surface statements** by end of
Phase 18.

---

## 9. Open questions for you

Before kicking off Phase 8, please confirm or adjust:

1. **Exclusion list** in §1. Anything you'd add or remove? In
   particular: does `octalyzer/ui/chat/` (1500 LOC) need testing or
   is it UI-coupled enough to exclude?
2. **Ordering.** I've put "recently modernized = highest regression
   risk first." If you'd rather hit the emulator core first
   (CPUs/memory in Phase 11 instead of files/network), the plan is
   reorderable.
3. **Sub-70% packages.** Should `core/hardware/spectrum` (65%) and
   `core/interfaces` (60%) really be below 70%, or should we either
   bring them up or exclude them entirely?
4. **Pace.** This is 10–16 weeks of focused work. If you want to ship
   the full coverage in fewer big PRs, I can fold pairs of phases
   together; if you want smaller PRs per package, I can split.
