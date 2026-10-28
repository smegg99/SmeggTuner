#!/usr/bin/env python3
# scripts/cross_build.py
"""Run the cross builder with local Go replacements at their container paths."""

import argparse
import json
import os
from pathlib import Path
import posixpath
import subprocess

# The image builds the package in its working directory and writes bin/ there.
MAIN_PACKAGE = "backend"


def replacement_mounts(root):
    module = json.loads(subprocess.check_output(
        ["go", "mod", "edit", "-json"], cwd=root, text=True))
    mounts = {}
    for replacement in module.get("Replace") or []:
        target = replacement["New"]
        if target.get("Version"):
            continue
        local = target["Path"]
        source = (root / local).resolve(strict=True)
        if not source.is_dir():
            raise ValueError(f"local Go replacement is not a directory: {source}")
        destination = posixpath.normpath(posixpath.join("/app", local))
        if destination in ("/", "/app", "/go", "/go/pkg", "/go/pkg/mod"):
            raise ValueError(f"local Go replacement overlaps a build mount: {destination}")
        mounts[destination] = source
    return [(source, destination) for destination, source in mounts.items()]


def volume(source, destination, readonly=False):
    if ":" in str(source) or ":" in destination:
        raise ValueError("Docker volume paths cannot contain ':'")
    return ["--volume", f"{source}:{destination}" + (":ro" if readonly else "")]


def build_command(root, image, target_os, arch, app, tags):
    cache = subprocess.check_output(["go", "env", "GOMODCACHE"], cwd=root, text=True).strip()
    command = ["docker", "run", "--rm", "--workdir", posixpath.join("/app", MAIN_PACKAGE)]
    command += volume(root, "/app") + volume(cache, "/go/pkg/mod")
    for source, destination in replacement_mounts(root):
        command += volume(source, destination, readonly=True)
    command += ["--env", f"APP_NAME={app}", "--env", f"EXTRA_TAGS={tags}"]
    # The image's Go (Wails' Dockerfile.cross pins golang:1.26) may be older than go.mod's; Go then
    # fetches the toolchain go.mod names into the mounted module cache.
    command += ["--env", "GOTOOLCHAIN=auto"]
    # Forward metadata used by the cross image without expanding it through a shell.
    for key in ("VERSION", "CI"):
        if key in os.environ:
            command += ["--env", f"{key}={os.environ[key]}"]
    return command + [image, target_os, arch]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, required=True)
    parser.add_argument("--image", required=True)
    parser.add_argument("--os", choices=("linux", "darwin", "windows"), required=True)
    parser.add_argument("--arch", choices=("amd64", "arm64"), required=True)
    parser.add_argument("--app", required=True)
    parser.add_argument("--tags", default="")
    args = parser.parse_args()
    root = args.root.resolve(strict=True)
    try:
        subprocess.run(build_command(root, args.image, args.os, args.arch, args.app, args.tags),
                       cwd=root, check=True)
    except (OSError, ValueError, subprocess.CalledProcessError) as error:
        parser.exit(1, f"cross build failed: {error}\n")


if __name__ == "__main__":
    main()
