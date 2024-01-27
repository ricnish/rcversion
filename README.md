# rcversion
A simple command line utility to read and change the version in [MSVC resource script file (.rc)](https://en.wikibooks.org/wiki/Windows_Programming/Resource_Script_Reference).

## Usage

Reading the version inside a resource file:
```
rcversion.exe read -p [resource file path]

Example:
rcversion.exe read -p c:\solution\project\procject.rc
```

Changing the version inside a resource file:
```
rcversion.exe change -p [resource file path] -v [N.N.N.N]

Example:
rcversion.exe change -p c:\solution\project\procject.rc -v 1.0.0.1
```

You can also change multiple files recursively by specifying the project's root directory:
```
rcversion.exe change -p [directory path] -v [N.N.N.N]

Example:
rcversion.exe change -p c:\solution -v 1.0.0.1
```

## Install

Install the Go programming language and compile the source files.

## Notes

TO DO:
- Read/Write UTF-16LE encoded resource files (my old source files were in ANSI format, sorry)
- Restrict the version to only accept N.N.N.N format

This project is an example (to me) on how to use the [cobra library](https://github.com/spf13/cobra) and to recursively walk into a directory and filter files.