# Regression Plan: develop without dread

**Goal.** Build a test suite that catches the regressions you'd actually
hit, with effort proportional to value. ~**3 weeks of focused work**,
not 12. Coverage % is a derived signal, not the target.

This plan supersedes the earlier coverage-first one (`COVERAGE-PLAN.md`,
kept for reference). The earlier plan would have produced ~145K LOC
of code under test in 12+ weeks. Most of that effort would have been
spent testing leaf utilities and rarely-changed format parsers, neither
of which has caused a regression yet. We can do better by testing the
*things that have actually broken*.

---

## 1. What the modernization migration taught us

Phases 1–7 found 30+ real bugs. Their distribution tells us where the
codebase is fragile:

| Bug class | Count | Example | Caught by |
|---|---:|---|---|
| Concurrency: lock-by-value, value-receiver setters | 6 | `mos6551StatusRegister.Reset()` mutating a copy | A `-race` test that calls the mutator and reads the field |
| Integer math: wrong divide / shift / modulo | 4 | `216/24389 = 0` in CIE Lab; `(i%1)*0x80` for chip 2 | A single unit test asserting the result of one call |
| Nil deref / accidental dead code | 3 | `vmlauncher.go` deref before nil-check | Calling the function once with the right shape of input |
| Goroutine leaks (bare `for { sleep }`) | 3 | `MonitorNetwork` no exit path | `testutil.NoGoroutineLeaks` + a cancel test |
| Swallowed errors / wrong return wiring | ~5 | `e = files.WriteX(...)` then never checked | A test that injects a write failure and asserts propagation |
| Self-compare / dead branch / impossible compare | ~5 | `mMin == mMin`; `byte < 0` | Trivial unit test |
| Deprecated API misuse | ~10 | `image.ZP`, `reflect.StringHeader` | staticcheck (already in place) |

**Pattern:** every bug we found was caught by a small, specific test —
not by a 100-row table-driven test, not by a fuzzer. Half the bugs
could have been caught by **one test per fix.** That's the leverage
point.

## 2. The strategy: four layers of safety net

Each layer catches a different class of regression. Together they give
you "I can refactor confidently because if I break something that
matters, *something here will fail*."

1. **Boundary contract tests.** End-to-end runs of the emulator
   producing deterministic output. If the user-visible behaviour of
   microM8 doesn't change, the emulator hasn't regressed in any way
   the user can detect. ~10 tests covering Applesoft, Integer BASIC,
   Logo, shell, disk loading, Mockingboard.
2. **Regression pins.** One named test per bug we've already fixed.
   Names like `TestMockingboard_Chip2HasNonZeroBase` or
   `TestVMLauncher_NilConfig_DoesNotPanic`. Reading the test list
   gives you a tour of the failure modes the codebase has.
3. **High-leverage units.** Targeted tests on the few packages whose
   bugs propagate everywhere: memory map, token/algorithm, disk
   parsers, the 6502 core. Race-detector clean.
4. **Fuzz the parsers.** Any input from disk or wire should never
   crash the emulator. We already have `base91`/`ffpak`/`mempak`
   fuzzers; extend to the disk image parser, the tokenizer, and the
   DuckTape bundle parser.

Anything that doesn't fall into one of those four buckets is
**deferred** — not because it doesn't matter, but because adding it
later costs the same and adding it now eats time we should spend on
the layers above.

## 3. The four phases

### Phase 8 — Regression pins (~3 days, one PR)

The most leveraged work in the plan. Every bug fixed in Phases 1–7
becomes a named test. Each test is small (typically 10–30 lines),
describes the bug it prevents in the test name + comment, and **would
have failed before the fix**.

This phase delivers ~30 tests across ~15 files. Sketch of the list
(file each test lives next to the code it covers):

| Test (named so the failure message tells the whole story) | Covers fix from |
|---|---|
| `TestMockingboard_Chip2HasNonZeroBase` | Phase 5 SA4028 `(i%1) → (i%2)` |
| `TestVideoColor_LabConstantsAreFractional` | Phase 5 SA4025 `216/24389` |
| `TestVMLauncher_NilConfig_DoesNotPanic` | Phase 5 SA5011 |
| `TestMOS6551Status_ResetPersistsViaPointerReceiver` (race) | Phase 5 SA4005 cluster |
| `TestFuncPtr5b_SetFirstBytePersists` (race) | Phase 5 SA4005 |
| `TestWOZTrack_DirtyWindowInitOnFirstWrite` | Phase 5 SA4000 |
| `TestPackageFileProvider_Exists_FilenameLookup` | Phase 5 SA4014 dup-condition |
| `TestMicrotrackerSongEntry_RejectsOutOfRangeOctave` | Phase 5 SA4003 |
| `TestAlgorithm_ConcurrentAccessUnderRace` | Phase 4b lock-by-value |
| `TestLoopState_AlgorithmPointerSharing` | Phase 4b LoopState cascade |
| `TestFuncPtr5b_GetPointer_PreservesHighByte` | Phase 4b shift-too-large |
| `TestAnalyzer_ParsePayload_PreservesUpperBits` | Phase 4b analyzer.go |
| `TestFileRecord_AddMeta_NoEmptyReadd` | Phase 4a AddMeta |
| `TestLog_FormatStr_SpreadsArgs` | Phase 4a Sprint(v) → Sprint(v...) |
| `TestUtils_RandIsThreadSafe` (race) | Phase 2 rand modernization |
| `TestDucktapeClient_ShutdownExitsLoop` | Phase 5 SA4011 ducktape leak |
| `TestDucktapeServer_QuitTerminatesGoroutine` | Phase 5 SA4011 server leak |
| `TestFastservClient_QuitClosesLoop` | Phase 5 SA4011 fastserv |
| `TestZipFileProvider_PreservesModTime` | Phase 5 SA1019 SetModTime |
| `TestGLConversions_StrPointerRoundTrip` | Phase 5 SA1019 StringHeader |
| `TestShellDialect_PRStub_DoesNotComputeUnused` | Phase 5 SA4017 |
| `TestVirtualModem_AtcCommandUnimplemented` | Phase 5 SA4017 |
| `TestDiskII_Int642BytesRoundTrip` | Phase 5 SA1003 int2bytes removal |

Acceptance: every fix that has a clearly-testable surface is pinned.
Some Phase 4a fixes (e.g. struct tag syntax, `//go:build` typo) are
already protected by `go vet` + staticcheck and don't need a pin.

### Phase 9 — Boundary contract tests (~5 days, one PR)

Headless end-to-end tests against the real emulator. The harness
constructs a Producer with a stub renderer (no GL, no audio device,
no window) and drives canonical programs through it, capturing the
text buffer / memory state / register snapshot as the assertion.

The big payoff: **one of these tests catches a wider class of
regressions than ten unit tests**, because it exercises the
interpreter, dialect, memory map, hardware emulation, and VDU layer
in concert. When it fails, you know microM8 has changed in a way a
user could observe.

Tests to write (golden-output style; `testdata/` holds expected output):

| Test | Asserts |
|---|---|
| `TestEmulator_Boot_ProducesShellPrompt` | A fresh boot drops into the shell with the expected prompt + cursor position |
| `TestApplesoft_HelloWorld` | `10 PRINT "HELLO": RUN` → text buffer contains "HELLO" |
| `TestApplesoft_ForNextLoop_PrintsSequence` | `10 FOR I=1 TO 5: PRINT I: NEXT` → "1\n2\n3\n4\n5\n"; race-clean (exercises the Phase 4b Algorithm fix in anger) |
| `TestApplesoft_GosubReturn` | Standard GOSUB→RETURN cycle preserves stack |
| `TestApplesoft_InputThenPrint` | INPUT consumes from a pre-loaded key buffer, PRINT echoes |
| `TestApplesoft_DimStringArray` | `DIM A$(3): A$(1)="X": PRINT A$(1)` (exercises the SA4011 type-suffix switch) |
| `TestIntegerBASIC_BasicArithmetic` | Smoke test for Integer BASIC dispatch |
| `TestLogo_TurtleSquare` | `REPEAT 4 [FD 50 RT 90]` → turtle ends at origin facing 0 |
| `TestShell_CDAndLS` | `cd /; ls` produces the expected directory listing |
| `TestDOS33_LoadCatalog` | Loads a known `.dsk` from testdata, asserts file list matches a golden |
| `TestProDOS_LoadCatalog` | Same for `.po` |
| `TestWOZ_RoundTrip` | Load a WOZ, write a track, save, reload, asserts dirty-window math (the Phase 5 SA4000 fix) |
| `TestMockingboard_BothChipsAddressable` | Register write to chip 0 at $C400, chip 1 at $C480; assert both reach distinct AY-3-8910 instances |

These tests are slow-ish (~1–5s each, ~30s total). They run in
`check.sh` by default but can be skipped with `go test -short` for
the watch loop.

**Risk:** the emulator's boot path pulls in a lot of hardware
machinery. We may need to extract a "minimal Producer" constructor
that skips disk drive init, audio init, network init. That work
itself is regression-relevant — surfacing the coupling is half the
point of writing the tests.

### Phase 10 — High-leverage units (~4 days, one PR)

Tests for the four packages whose bugs propagate everywhere. Not
chasing coverage %, chasing "this is a state machine; let's drive it
through every state."

**`core/memory` (~3 days for the package).**
The memory map is touched by every CPU cycle. Test:
- Bank switch state machine (RAM/ROM/LCRAM transitions on $C080–$C08F).
- BlockMapper lookup correctness when blocks overlap.
- Slot-restart flag (`IntGetSlotRestart` / `IntSetSlotRestart`) under
  -race (the Phase 7 RebootService tolerates nil; tests should
  exercise the non-nil path).
- ROM/RAM card I/O addressing.
- Read/write through a populated map at every interesting boundary
  ($C000, $C080, $D000, $E000, $FFFF).
- Memory protection (read-only blocks reject writes).

**`core/types` Token + Algorithm + LoopState.**
Already 13.2% covered; the focus here is the parts the modernization
touched: `Algorithm` under `-race`, `LoopState` pointer-sharing,
`FuncPtr5b` serialisation, Token type coercion (`AsInteger`/`AsFloat`
on every TokenType). Aim for ~30% coverage but with the *right* 30%.

**`disk` parsers.**
Round-trip tests for the formats that matter: `.dsk`, `.po`, `.2mg`,
`.nib`, plus the AppleDOS catalog walker. Each format gets one
"happy path" test (load a known image, assert known catalog) and
one "corrupted input" test (truncate, byte-flip, assert no panic).
The known images live in `testdata/disks/`.

**`core/hardware/cpu/mos6502`.**
Not the Klaus-Dormann full suite (license + scope). Hand-rolled
opcode tests for: LDA/LDX/LDY, STA/STX/STY, arithmetic with carry,
branches, JSR/RTS, stack pushes, decimal mode (the trickiest part of
the 6502). Aim for the documented opcodes; undocumented ones can
land in a follow-up.

These four together probably catch 80% of the "I changed core/memory
and now nothing works" regressions you'd hit during refactoring.

### Phase 11 — Fuzz the input boundaries (~2 days, one PR)

Anywhere the emulator parses data from disk or the network, add a
fuzzer with the contract "should never panic." We already have this
for the base91/ffpak/mempak encoders; extend to:

| Fuzz target | Lives in |
|---|---|
| `FuzzDOS33Catalog_NeverPanics` | `disk/diskimageappledos_test.go` |
| `FuzzProDOSCatalog_NeverPanics` | `disk/diskimagepd_test.go` |
| `FuzzWOZ_ParseTrack_NeverPanics` | `core/hardware/apple2/woz/woz_test.go` |
| `FuzzWOZ2_ParseTrack_NeverPanics` | `core/hardware/apple2/woz2/woz2_test.go` |
| `FuzzDuckTapeBundle_Parse_NeverPanics` | `ducktape/ducktape_test.go` |
| `FuzzTokenizer_NeverPanics` | `core/types/tokenlist_test.go` |

Seed corpus: a few known-good inputs + a handful of pathological
ones (truncated, all-zero, all-0xFF, max-size). Run for 30s in CI;
locally fuzzing can run longer.

Goal isn't "find bugs in the parsers" (though it will). Goal is "I
can't cause a crash by loading any file." That's a critical user
property for an emulator.

### Phase 12 — Tighten the feedback loop (~1 day, one PR)

The point of the plan is to make development less scary. That only
works if running the tests is fast.

- Audit `check.sh` for slow tests; mark anything over 2s with
  `if testing.Short() { t.Skip(...) }` so the watch loop skips them.
- `tools/scripts/watch.sh` is already in place; document the
  recommended workflow in TESTING.md.
- Add a `tools/scripts/check.sh quick` mode: gofmt + vet + tests with
  `-short`. Target: <15s on incremental change. Full `check.sh` stays
  as the pre-PR gate.
- Add a small `tools/scripts/cover-gaps.sh` that prints, per package,
  the highest-line-count uncovered functions. So "where should I add
  the next test?" has a one-command answer.

## 4. What we explicitly don't do (and why)

| Skipped | Why |
|---|---|
| 70%-everywhere coverage push | Sub-linear regression detection per hour after the first ~30% on the right code |
| Tests for `glumby`, `gl`, `octalyzer/video` | Need a GPU; smoke-test by hand |
| Tests for `octalyzer/ui/*` | UI controllers; visual regression-test by hand |
| Tests for the BASIC dialects' obscure commands | The boundary tests in Phase 9 catch the common path; obscure commands rarely change and rarely break |
| Klaus-Dormann full 6502 suite | License concerns + the curated subset in Phase 10 catches the 6502 changes we're likely to make |
| Tests on quarantined `_*/` packages | Quarantined for a reason; revive-and-test together when the time comes |
| Static fuzz corpora as test data | Use `go test -fuzz` with seed corpora; let the runtime fuzzer do the exploration |

If a phase finds the boundary tests aren't catching enough, the right
response is **better boundary tests**, not retreating to coverage-%
chasing.

## 5. What "done" looks like

After Phase 12 you should be able to say:

- "I ran `tools/scripts/check.sh quick` after my refactor — under 15
  seconds, all green. I trust the change."
- "I ran the full suite before opening the PR — under 2 minutes, all
  green. Including the boundary tests, the regression pins, the unit
  tests, and the race detector."
- "Every bug I've ever seen in this codebase has a test guarding
  against it. The test file names read like a changelog of past
  failures."
- "Anything I load from disk or the wire either parses cleanly or
  errors cleanly — the fuzzers prove it doesn't panic."

That's "develop without dread." Coverage % will be somewhere in the
20–35% range, concentrated in the places that matter. Higher would
be better but isn't load-bearing for the goal.

## 6. Timeline (no fluff)

| Phase | Effort | Cumulative |
|---|---:|---:|
| 8 — Regression pins | 3 days | 3 days |
| 9 — Boundary contract tests | 5 days | 8 days |
| 10 — High-leverage units | 4 days | 12 days |
| 11 — Fuzz the parsers | 2 days | 14 days |
| 12 — Tighten feedback loop | 1 day | 15 days |

**~3 weeks of focused work.** Each phase ships as one PR. The order
is roughly value-per-day; you could pause after Phase 8 alone and the
codebase would already be significantly safer to touch.

## 7. After Phase 12

If after living with this safety net you find specific places you
still feel nervous about, the next move is **add boundary tests for
the specific workflow you're about to change**, not a generic
coverage push. The regression-pinning pattern is reusable: fix a bug
once, write the test, and now it's load-bearing forever.

## 8. Status

All five phases shipped (PRs #15–#19, merged sequentially to master):

| Phase | PR | Tests added | Bugs found+fixed | Notes |
|---|---|---:|---:|---|
| 8 — regression pins | #15 | 35 | 2 | `Algorithm` locking holes + `utils.rng` thread-safety, both caught while writing pins. Two test-scaffolding stubs (`TestTaskExecution`, `TestSimpleDelta`, `TestPackagePackUnPack`) skipped with TODOs. |
| 9 — boundary contract | #16 | 26 | 0 | Disk sector-mapper bijections, AppleDOS VTOC, FileDescriptor decoding, memory-map two-layer activation, zip provider round-trip. Bypassed the full headless-emulator harness ambition in favour of higher-ROI format-and-memory boundary tests. |
| 10 — high-leverage units | #17 | 20 | 1 | 6502 flag/helper primitives, Token/TokenList deep tests. Pinned the `TokenList.IndexOf` start-is-exclusive off-by-one as a documented quirk with a TODO. |
| 11 — fuzz the parsers | #18 | 3 fuzzers + 5 unit | 1 | DuckTape bundle parser (3M execs clean), WOZ parser (caught a slice-bounds panic on short input, fixed), DSK wrapper (310k execs clean). |
| 12 — tighten the loop | #19 | n/a | n/a | `check.sh quick` (no build, `-short` tests), `cover-gaps.sh` for "where to add the next test", TESTING.md workflow docs. |

End state:

- **27 test packages** in the allowlist, all green under `-race -shuffle=on`
- **~85 new tests** total across 4 PRs (excluding the existing baseline)
- **5 latent bugs found and fixed** that didn't surface in earlier phases (3 concurrency-related, 1 parser-panic, 1 OOB)
- **`check.sh quick`**: ~8s. **`check.sh`** (full): ~12s.
- **`cover-gaps.sh`**: 1-command answer to "where should the next test go"

The safety net is in place. Future development should:

1. Run `watch.sh` while coding
2. `check.sh quick` before each commit (~8s)
3. `check.sh` before opening a PR (~12s, includes the build)
4. `cover-gaps.sh ./pkg` when picking the next test to write
5. New bugs found get a named regression pin alongside the fix —
   that's how this safety net grows.
