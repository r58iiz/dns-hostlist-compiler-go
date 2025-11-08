# dns-hostlist-compiler-go

A small Go tool to compile and process Adguard-style DNS hostlist rules from a list of links.

Technically an older version of https://github.com/AdguardTeam/HostlistCompiler rewritten in GoLang because I felt not having an executable was "stupid".

`main.go` was split up into `modules/app/cli`, `modules/app/io` and `modules/app/pipeline` with Agent mode in VSCode because it was hardcoded to read `links.txt` as input file and `output.txt` as the output file and all "pipeline" logic was in the `main` function. This seems to be better but will be re-written as a proper tool and perhaps match the current version of [AdguardTeam/HostlistCompiler](https://github.com/AdguardTeam/HostlistCompiler).

## Quick build

Open a terminal in the project root and run:

```powershell
# build executable (Windows)
go build -o dns-hostlist-compiler-go.exe .

# or run directly without building
go run main.go --input=<input-file> --output=<output-file>
```

## Usage

The program expects two positional arguments: an input file containing links (one per line) and an output file where compiled rules will be written.

### Examples

```powershell
# run with go run
go run main.go --input=links.txt --output=rules.txt

# run the built executable
.\dns-hostlist-compiler-go.exe --input=links.txt --output=rules.txt
```

## What it does

- Reads links from the input file
- Deduplicates the links
- Runs a processing pipeline that validates, cleans, and compiles hostlist rules
- Writes the resulting rules to the output file and prints the number of rules written