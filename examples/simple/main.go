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
	//      C:/Windows/System32/Drivers/etc
	// bsd/linux:
	//      /etc

	// apathy.Join is like path.Join but also guarantees a Clean().
	// Use apathy.JoinPieces() when all the parts will implement Piecer.
	file := apathy.Join(folder, "hosts")
	fmt.Printf("hosts file is `%s`\n", file)

	// APathFromLstat is going to resolve the list of pieces we provide into a single,
	// APiece that we know to be an absolute path, Clean()d, with posix slashes.
	// Then it will Lstat that file and capture the information.
	selfStatFile, err := apathy.NewAPath(file)
	notExpectingAn(err)

	// Let's see how it looks:
	fmt.Printf("Our APath collected lstat info for `%s`:\n", selfStatFile)
	fmt.Printf("| Exists %t\n", selfStatFile.Exists())
	fmt.Printf("| IsFile %t\n", selfStatFile.IsFile())
	fmt.Printf("| IsDir %t\n", selfStatFile.IsDir())
	fmt.Printf("| IsSymlink %t\n", selfStatFile.IsSymlink())
	fmt.Printf("| ModTime %s\n", selfStatFile.ModTime())
	fmt.Printf("| Size %d\n", selfStatFile.Size())

	// Or you can pass the lstat info yourself.
	fileinfo, err := os.Stat(file.String())
	if err != nil && !os.IsNotExist(err) {
		notExpectingAn(err)
	}
	infodFile, err := apathy.NewAPathWith(file, fileinfo, err)
	notExpectingAn(err)
	fmt.Printf("With our own lstat: exists: %t file:%t dir:%t\n", infodFile.Exists(), infodFile.IsFile(), infodFile.IsDir())

	// Demonstrate the conversion back to native form.
	fmt.Printf("folder.ToNative: %s\n", folder.Normalize())
	fmt.Printf("apathy.ToNative(infodFile): %s\n", apathy.Normalize(infodFile))

	// Ok, but what if we mix things oup.
	var posixStr = "/etc/hosts"
	var winStr = "c:\\windows/system32/drivers/etc/hosts"

	var posixPiece = apathy.NewAPiece(posixStr)
	var winPiece = apathy.NewAPiece(winStr)

	fmt.Printf("`%s` -> APiece -> ToNative = `%s`\n", posixStr, posixPiece.Normalize())
	fmt.Printf("`%s` -> APiece -> ToNative = `%s`\n", winStr, winPiece.Normalize())
}
