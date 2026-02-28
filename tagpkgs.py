#!/usr/bin/env python3

"""Tag packages with a version and push to origin.

Usage: tagpkgs.py <version>
  version: must match vN.N.N (e.g. v1.2.3)
"""

import re
import subprocess
import sys


def main() -> int:
    usage = (
        "Usage: tagpkgs.py <version>\n"
        "  version: must match vN.N.N (e.g. v1.2.3)"
    )

    if len(sys.argv) != 2:
        print(usage, file=sys.stderr)
        return 1

    version = sys.argv[1]
    if not re.match(r"^v\d+\.\d+\.\d+$", version):
        print(f"tagpkgs.py: invalid version '{version}'", file=sys.stderr)
        print(usage, file=sys.stderr)
        return 1

    subprocess.run(["git", "tag", version], check=True)
    subprocess.run(["git", "tag", f"go/{version}"], check=True)
    subprocess.run(["git", "push", "origin", version], check=True)
    subprocess.run(["git", "push", "origin", f"go/{version}"], check=True)
    return 0


if __name__ == "__main__":
    sys.exit(main())
