#!/usr/bin/env python
"""
Pre-commit hook to sort dependencies in a file.

This script sorts the dependencies listed in a file. It expects the file
to contain dependencies in the format 'package_name version', with one
dependency per line. Optional comment lines starting with '#' are allowed.
The script sorts the dependencies alphabetically by package name and
rewrites the file with the sorted dependencies.

Usage: pre-commit hook <file_path>

Arguments:
    file_path: Path to the file containing dependencies to be sorted.
"""

import re
import sys
import argparse
from typing import IO
from typing import Sequence

PASS = 0
FAIL = 1

# captures package version
PACKAGE_REGEX = r"^([\w|\_|\-|\.]+)\s([\w|\_|\-|\.]+)$"


def sort_dependencies(
    f: IO[bytes],
) -> int:
    # Read the content of the file and decode it into a string
    content = f.read().decode()

    package_dict = {}

    # capture an optional comment line followed by the package and its version
    groups = re.findall(f"((^#.*\n)?{PACKAGE_REGEX})", content, re.MULTILINE)

    for group in groups:
        package_name = group[2]
        package_dict[package_name] = group[0]

    sorted_packages = sorted(package_dict.items())

    sorted_content = "\n\n".join([package[1] for package in sorted_packages])
    sorted_content += "\n"
    # Write sorted content back to the file
    f.seek(0)
    f.write(sorted_content.encode())
    f.truncate()  # Truncate any extra content beyond what's written

    # Compare sorted content with original content
    if sorted_content.encode() == content.encode():
        return PASS
    else:
        return FAIL


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("filenames", nargs="+", help="Files to sort")
    args = parser.parse_args(argv)

    retv = PASS

    for arg in args.filenames:
        with open(arg, "rb+") as file_obj:
            ret_for_file = sort_dependencies(
                file_obj,
            )

            if ret_for_file:
                print(f"Sorting {arg}")

            retv |= ret_for_file

    return retv


if __name__ == "__main__":
    raise SystemExit(main())
