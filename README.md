[![Discord](https://img.shields.io/discord/999601338407190569.svg?label=&logo=discord&logoColor=ffffff&color=7389D8&labelColor=6A7EC2)](https://discord.gg/mYBn9XqRBr)

Feng is a yaml-based template format for specifying binary file structures,
and a commandline app working with the format.

**WARNING : The format is not yet stabilized and things may break.**




# Template example

Binary formats are specified in a template file, see [TEMPLATE.md](TEMPLATE.md) for the format documentation.

Here's the OpenType Font table parsing:

```yaml
endian: big

magic:
  - offset: 0000
    match: c'OTTO'

structs:
  header:
    ascii[4] Magic: c'OTTO'
    u16 TableCount: ??
    u16 SearchRange: ??
    u16 EntrySelector: ??
    u16 RangeShift: ??

  table:
    ascii[4] Tag: ??
    label: self.Tag

    u32 Checksum: ??
    u32 Offset: ??
    u32 Length: ??

    offset: self.Offset
    u8[self.Length] Data: ??
    u8[alignment(self.Length, 4)] Padding: ??
    offset: restore

layout:
  - header Header
  - table[Header.TableCount] Table
```

Mapping a file with the above structure would result in something like [smoketest/reference/fonts/otf](smoketest/reference/fonts/otf).

*Template is from [templates/fonts/otf.yml](templates/fonts/otf.yml).*



# Installation

Windows/macOS and Linux binaries is available on the [Releases](https://github.com/martinlindhe/feng/releases) page.

Assuming you have golang, then installing the cli app `feng` from git is as easy as:

    go install github.com/martinlindhe/feng/cmd/feng@latest


# Command line tools

```
cmd/feng: Lists data fields of input file.
cmd/renamer: Renames all files in input folder to use the correct file extensions.
```

