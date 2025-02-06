Apathy - because who has time for all that stat business?
=========================================================

Good, portable path code often wastes a lot of cpu cycles on redundant operations such as
Clean() and Abs(), because these operations return `string`s.

Apathy provides a thin layer of compile-time information about file paths and a way to
remember important things about paths like: did it exist, was it a directory.

## Types

`APiece` is a `string` that promises the _value_ is in posix-separator style and has
undergone a `path.Clean()`, eliminating the need for repetitive untainting Clean()s.

`APath` guarantees an absolute, posix-separated path, coupled with Lstat-based information,
i.e. whether the object is a file/directory/symlink, it's mtime, and the size for a file.

## Interfaces

`PieceMeal` interface for any type that has a `Piece() APiece` method, including `APiece`.


# But, Windows...

Generally copes fine with posix-style paths. In Super Evil Megacorp's (actual name)
asset conversion pipeline, out of ~30 million filesystem interactions on Windows in one run,
under 200 of them actually needed old-school paths.

Eliminating the redundant posix-to-old-win transitions reduced a no-op run from 8 seconds
to under 2, and shaved just over 30 seconds off a full run.


# Target use cases

You're doing something very cross-platform and can reasonably see any kind of mix or
`/` and `\\` separators in your code's run time.

Or, you're writing a *lot* of functions that take and pass a lot of paths, knowing your
constituent parts of paths can save a lot of cycles.

When you are discovering the majority of your paths via a Walker or ReadDir, or other
means that require an Lstat or Stat of the file it, and will want that information again
later...


# What if I know all my paths will be Posix or Windows?

If you aren't doing a lot of Abs() and Clean() transformations through out your code,
we've got nothing going for you.


# Examples

See examples folder.
