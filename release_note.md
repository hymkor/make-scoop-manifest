- Append keywords to judge architectures for Rust
    - for 32bit `486`, `586`, and `686`
    - for 64bit `x86_64`

v0.6.0
======
Jan 25, 2023

- Print usage when no zip-files are given
- Add the option -downloadto DIRECTORY
    - It is same as -D, but does not remove ZIP file and leaves onto DIRECTORY.
- Sort the items of "bin"
- Add the option -license
    - `-license "MIT"` is same as `-inline "{ \"license\":\"MIT\" }"`
- Add the option -description
    - `-description "XXX"` is same as `-inline "{ \"description\":\"XXX\" }"`

v0.5.0
======
Jan 17, 2023

- Add the options:
    - -32 "string" : When these strings are found, set architecture 32bit (default "386,32bit,win32")
    - -64 "string" : When these strings are found, set architecture 64bit (default "amd64,64bit,win64")

v0.4.0
======
Jan 17, 2023

- Add -p option:
    - Set the parent-directry of `*.exe` into `extract_dir`
    - Set the basename of `*.exe` into `bin`

v0.3.0
======
Jan 12, 2023

- Add -anycpu option:
    - Do not use "architecture" field
    - Do not check whether the zip file's name has ether 386 or amd64 keyword.

v0.2.2
======
Jan 10, 2023

- When either "license" or "description" is empty, print warning

v0.2.1
======
Jan 9, 2023

- (#2) Set "github" only into the manifest item "checkver" ( Thx @spiegel-im-spiegel )

v0.2.0
======
Jan 9, 2023

- (#1) 32bit, 64bit and arm64 are available as the architecture keyword. (Thx @spiegel-im-spiegel )

v0.1.1
======
Jan 9, 2023

- Fix: error when repository-name had -

v0.1.0
======
Jan 9, 2023

- The first release.
