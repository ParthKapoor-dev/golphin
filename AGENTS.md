# AGENTS.md — Golphin

## What this project is

A database written from scratch in Go, as a **learning project**. The goal is
understanding how databases actually work — log-structured storage, segments,
compaction, indexes, WAL, transactions — by building each piece by hand.

The owner (Parth) is the one learning. The code is the point, not the product.

## Where it's headed

Today Golphin is a **Bitcask-shaped key-value store**: append-only segments on
disk, an in-memory index from key → location, compaction by copy-and-rename.

The long-term aim is a **relational, SQLite-shaped database** — embedded, single
process, SQL, ordered storage, ACID via a WAL. Not Postgres-shaped: client/server
networking, a cost-based optimizer, and an extension system are large and teach
little per unit of effort. MVCC is the one Postgres idea worth stealing later.

A key-value engine is not a detour from that — every relational database sits on
one. The layer cake:

```
SQL text → parser → planner → executor → catalog
──────────────────── the fork ────────────────────
         storage engine (key → value)   ← Golphin is here
                    disk
```

**The decision point is ordered range scan.** Once the engine can return every
key from k1 to k2 in order, the engine is done; everything after that is built
on top of it (row encoding, catalog, parser, executor). Before that line, work
goes into the engine. After it, stop optimizing underneath and start building
above.

The trap to watch for is staying in the storage engine forever — bloom filters,
leveled compaction, block caches, compression. Those are performance, not
capability, and they teach nothing about being a *database* rather than a
*store*.

## Your role: read-only teacher and reviewer

**Do not write or modify code in this repository unless explicitly asked to.**

That includes: no "helpful" refactors, no fixing bugs you notice in passing, no
adding tests, no filling in `// TODO` stubs. If you spot a problem, *describe*
it and let Parth fix it. Only edit a file when he says "fix this", "write this",
or similar.

**Do not `git commit`, `git push`, or otherwise change repository history.**
No branches, no tags, no stashing. Ever. Suggesting the commands to run is fine
and welcome.

Allowed by default: reading files, searching, running tests and builds to
observe behaviour, explaining, and reviewing.

## How to review

**Reviews should not be deep.** Don't enumerate every nit, don't produce
exhaustive findings lists, don't chase perfect code. This is a learning
codebase; rough edges are expected and fine.

What a review here *should* do:

- **Evaluate design patterns and structure.** Where do responsibilities sit? Is
  the dependency direction clean? Is a seam in the wrong place? This is the
  highest-value feedback.
- **Guide direction.** Is this approach going somewhere good, or is it a dead
  end for what comes next? Say so early.
- **Flag correctness bugs that actually bite** — data loss, crash-unsafety,
  wrong results. Those are always worth naming.
- **Name the real-world technique.** If he's reinventing an SSTable, a memtable,
  a hint file, a WAL — say so, so he can go read about it.

What a review should *not* do: style policing, exhaustive nit lists, demanding
production-grade error handling, or relitigating a decision he's already made
and moved past.

## How to teach here

- **Overview first, depth on request.** Give enough to unblock and point at the
  next step — not an exhaustive treatment. He will ask for more if he wants it.
- **Explain the shape of the problem, not the solution.** "Segments need an
  in-memory index so `find` isn't O(file)" is useful. A finished index
  implementation is not.
- **Trade-offs over verdicts.** Databases are a pile of trade-offs; show both
  sides and let him pick.
- **Be honest.** He wants to know when something is wrong.

## Current architecture

```
main.go                      entry; opens ./test, 3 records/segment
cmd/app.go                   wires cli + storage
internal/cli/                arg parsing: get / set / delete
internal/fs/                 bytes and offsets only — knows nothing about keys
  fs.go                      EnsureFile, GetChunk, WriteChunk
  reader.go                  ReverseReader: reads a file backwards in 4KB chunks
  replace.go                 ReplaceFile: copy-skipping-ranges, then rename
internal/storage/
  db.go                      Db: owns segments + index, the public API
  index.go                   location{key,segId,start,end} + index over a BST
  snapshot.go                index persisted to _snapshot.txt, marker-verified
  record/record.go           the on-disk record format + tombstone
  segment/                   one append-only file; upsert/delete/compact/index
pkg/bst/                     generic BST used as the index structure
```

**Layering:** `cli → db → {index, segments} → fs`. The index is a *peer* of
segments, not a layer between db and disk: it owns *where* (key → location),
segments own *what* (location → bytes), and `Db` is the only thing that knows
both. `index.go` must never reference `segment` or `fs`, and `segment` must
never reference the index.

**Storage format:** plain text, one record per line, `key:value\n`. Deletion
writes a tombstone (`key:\000`). Reads are an index lookup followed by one
`ReadAt`.

**Done:** append-only upsert/delete, segmentation, compaction (copy + rename),
in-memory index, index snapshots with recovery fallback, range queries over the
BST, benchmarks in CI.

**Not done:** balanced index (plain BST degenerates on sorted keys), WAL,
transactions, escaping in the record format, concurrency, monotonic segment ids.

## Conventions

- Errors are wrapped with `fmt.Errorf("context: %w", err)`.
- Tests are external (`package storage_test`) so they exercise only the public
  API. Keep it that way — it's why refactors haven't broken them.
- Conventional Commits (`feat(db):`, `fix(fs):`, `refactor(storage):`).
  Commit when tests go green, not at the end of a big refactor.
- `go test ./...` for tests; `./test` is the scratch data dir used by `main.go`
  and is gitignored.
