v0.2 2025/02/05 cleanup, more tests, focus
- 100% test coverage
- APathFrom... methods became NewAPath(pieces...) and NewAPathFromFile(abspath, info, err)
- APath became an interface with concerete 'aPath' as the base
- ToNative became Normalize
- NewAPiece attempts `\` -> `/` conversion regardless of platform
- - Trying to care about whether or not to care about slashes defeats the purpose of the library
- Windows:
- - Normalize always convert `/` -> `\`,
- Posix:
- - Normalize will convert `/` -> `\` if there appears to be a drive letter, e.g. c:/win -> c:\win

v0.1 2025/02/03 initial release
