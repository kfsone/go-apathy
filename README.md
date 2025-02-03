Apathy - because who has time for all that stat business?
=========================================================

Good, portable path code often wastes a lot of cpu cycles on redundant operations such as
Clean() and Abs(), because these operations return `string`s.

Apathy provides a thin layer of compile-time information about file paths and a way to
remember important things about paths like: did it exist, was it a directory.

`APiece` is a type alias for `string` that promises the _value_ is in posix-separator style
and has undergone a `path.Clean()`.

`APath` guarantees an absolute, posix-separated path, coupled with Lstat-based information,
i.e. whether the object is a file/directory/symlink, it's mtime, and the size for a file.

`PieceMeal` interface for any type that has a `Piece() APiece` method, including `APiece`.


# But, Windows...

Generally copes fine with posix-style paths. In Super Evil Megacorp's (actual name)
asset conversion pipeline, out of ~30 million filesystem interactions on Windows in one run,
under 200 of them actually needed old-school paths.

Eliminating the redundant posix-to-old-win transitions reduced a no-op run from 8 seconds
to under 2, and shaved just over 30 seconds off a full run.


# Target use cases

You're doing something very cross-platform and can reasonably see any kind of mix or
/ and \\ separators in your code's run time.

Or, you're writing a *lot* of functions that take and pass a lot of paths, knowing your
constituent parts of paths can save a lot of cycles.

When you are discovering the majority of your paths via a Walker or ReadDir, or other
means that require an Lstat or Stat of the file it, and will want that information again
later...


# Maximum compatability

If you *are* expecting you might see windows-slashes on a non-posix system, you can
set apathy.ExpectWindowsSlashes and it will do a little extra sanitization to ensure
your paths don't end up with non-windows characters.


# Example

See examples/...

```go
package main

import (
	"fmt"
	"os"
	"runtime"

	apathy "github.com/kfsone/go-apathy"
)

func notExpectingAn(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// A string in some jumbled separator combination referencing the
	// place where the standard 'hosts' file exists.
	var hostsLocation = ""
	if runtime.GOOS == "windows" {
		hostsLocation = "C:\\Windows/System/..\\System32/Drivers\\etc"
	} else {
		hostsLocation = "\\etc" // ooh, that's wrong?
	}

	// This turns it into an APiece, a clean, posix-shaped, pathing component.
	folder := apathy.NewAPiece(hostsLocation)

	fmt.Printf("folder is `%s`\n", folder)
	// windows:
	//      C:/Windows/System32/Drivers/etc"
	// bsd/linux:
	//      /etc

	// apathy.Join allows joining any `PieceMeal` type, eg APiece and string.
	file := apathy.Join(folder, "hosts")
	fmt.Printf("hosts file is `%s`\n", file)

	// APathFromLstat is going to resolve the list of pieces we provide into a single,
	// APiece that we know to be an absolute path - and also clean and posixy.
	// Then it will Lstat that file and capture the information.
	selfStatFile, err := apathy.APathFromLstat(file)
	notExpectingAn(err)

	// Lets see how it looks:
	fmt.Printf("Our APath collected lstat info for `%s`:\n", selfStatFile)
	fmt.Printf("| Exists %t\n", selfStatFile.Exists())
	fmt.Printf("| IsFile %t\n", selfStatFile.IsFile())
	fmt.Printf("| IsDir %t\n", selfStatFile.IsDir())
	fmt.Printf("| IsSymlink %t\n", selfStatFile.IsSymlink())
	fmt.Printf("| ModTime %s\n", selfStatFile.Mtime())
	fmt.Printf("| Size %d\n", selfStatFile.Size())

	// Or you can pass the lstat info yourself.
	fileinfo, err := os.Stat(file.String())
	if err != nil && !os.IsNotExist(err) {
		notExpectingAn(err)
	}
	infodFile, err := apathy.APathFromInfo(fileinfo, err, file)
	notExpectingAn(err)
	fmt.Printf("With our own lstat: exists: %t file:%t dir:%t\n", infodFile.Exists(), infodFile.IsFile(), infodFile.IsDir())

	// Demonstrate the conversion back to native form.
	fmt.Printf("folder.ToNative: %s\n", folder.ToNative())
	fmt.Printf("apathy.ToNative(infodFile): %s\n", apathy.ToNative(infodFile))

	// Ok, but what if we mix things oup.
	var posixStr = "/etc/hosts"
	var winStr = "c:\\windows/system32/drivers/etc/hosts"

	var posixPiece = apathy.NewAPiece(posixStr)
	var winPiece = apathy.NewAPiece(winStr)

	fmt.Printf("`%s` -> APiece -> ToNative = `%s`\n", posixStr, posixPiece.ToNative())
	fmt.Printf("`%s` -> APiece -> ToNative = `%s`\n", winStr, winPiece.ToNative())

	fmt.Printf("what if we're more expectant of windows slashes?\n")
	defer apathy.WithPossibleWindowsSlashes()()

	posixPiece = apathy.NewAPiece(posixStr)
	winPiece = apathy.NewAPiece(winStr)

	fmt.Printf("`%s` -> APiece -> ToNative = `%s`\n", posixStr, posixPiece.ToNative())
	fmt.Printf("`%s` -> APiece -> ToNative = `%s`\n", winStr, winPiece.ToNative())
}
```
