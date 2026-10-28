#!/usr/bin/env python3
# scripts/test_release.py
"""Exercise scripts/release.sh with temporary files and fake builders only."""

import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[1]
LINUX_BUILD = """#!/usr/bin/env bash
set -eu
if [[ "$*" == *ARCH=arm64* ]]; then exit 17; fi
mkdir -p bin
for suffix in '' .deb .rpm .pkg.tar.zst -amd64.AppImage; do
    echo artifact > "bin/smeggtuner$suffix"
done
"""
WINDOWS_BUILD = """#!/usr/bin/env bash
set -eu
if [[ "$1" == sync ]]; then exit 0; fi
mkdir -p bin/lab-out/win11
echo executable > bin/lab-out/win11/smeggtuner.exe
echo installer > bin/lab-out/win11/smeggtuner-amd64-installer.exe
"""
ANDROID_BUILD = """#!/usr/bin/env bash
set -eu
out=bin/android/release/$2
mkdir -p "$out"
case "$1" in package) ext=apk ;; bundle) ext=aab ;; *) exit 9 ;; esac
echo artifact > "$out/smeggtuner-1.2.3-release-$2.$ext"
"""
PACKAGE = """#!/usr/bin/env bash
set -eu
echo archive > "$2/smeggtuner-1.2.3-linux-$3.tar.zst"
echo sum > "$2/smeggtuner-1.2.3-linux-$3.tar.zst.sha256"
"""
NODE_ENV = "#!/bin/sh\nexec \"$@\"\n"


class ReleaseTests(unittest.TestCase):
    def run_release(self, targets, linux=LINUX_BUILD, windows=WINDOWS_BUILD, android=ANDROID_BUILD,
                    fail_tool=None, cross="#!/bin/sh\nexit 16\n"):
        with tempfile.TemporaryDirectory(prefix="smegg-release-test-") as temporary:
            root = Path(temporary)
            tools = root / "tools"
            for directory in (tools, root / "scripts", root / "testlab", root / "build/android-support"):
                directory.mkdir(parents=True)
            (root / "VERSION").write_text("1.2.3\n")
            shutil.copy(ROOT / "scripts/release.sh", root / "scripts/release.sh")
            # node_env.py is run by python3; a shell shim named python3 forwards it.
            executables = {
                tools / "go": linux,
                tools / "python3": "#!/bin/sh\nshift\nexec \"$@\"\n",
                root / "scripts/node_env.py": NODE_ENV,
                root / "scripts/package.sh": PACKAGE,
                root / "scripts/linux-cross.sh": cross,
                root / "testlab/lab.sh": windows,
                root / "build/android-support/android.sh": android,
            }
            if fail_tool:
                executables[tools / fail_tool] = "#!/bin/sh\nexit 23\n"
            for path, content in executables.items():
                path.write_text(content)
                path.chmod(0o755)
            result = subprocess.run(["bash", "scripts/release.sh", *targets.split()], cwd=root, timeout=15,
                                    env=os.environ | {"PATH": str(tools) + os.pathsep + os.environ["PATH"]},
                                    capture_output=True, text=True)
            out = root / "dist/release/1.2.3"
            files = sorted(path.name for path in out.glob("*"))
            sums = (out / "SHA256SUMS").read_text() if (out / "SHA256SUMS").exists() else ""
            return result, files, sums

    def test_complete_linux_release(self):
        result, files, sums = self.run_release("linux/amd64")
        self.assertEqual(result.returncode, 0, result.stderr)
        names = ["smeggtuner-1.2.3-linux-amd64" + suffix for suffix in
                 (".AppImage", ".deb", ".pkg.tar.zst", ".rpm", ".tar.zst")]
        self.assertEqual(files, sorted(names + ["SHA256SUMS"]))
        self.assertEqual(sorted(line.split()[1] for line in sums.splitlines()), sorted(names))

    def test_complete_windows_release(self):
        result, files, _ = self.run_release("windows/amd64")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(files, ["SHA256SUMS", "smeggtuner-1.2.3-windows-amd64-installer.exe",
                                 "smeggtuner-1.2.3-windows-amd64.exe"])

    def test_complete_android_release(self):
        result, files, _ = self.run_release("android/arm64")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(files, ["SHA256SUMS", "smeggtuner-1.2.3-android-arm64.aab",
                                 "smeggtuner-1.2.3-android-arm64.apk"])

    def test_failed_target_cannot_be_hidden_by_success(self):
        for targets in ("linux/amd64 linux/arm64", "linux/arm64 linux/amd64",
                        "linux/amd64 windows/arm64", "linux/amd64 darwin/arm64"):
            with self.subTest(targets=targets):
                result, _, _ = self.run_release(targets)
                self.assertNotEqual(result.returncode, 0, result.stdout)

    def test_a_foreign_linux_architecture_is_cross_built_without_an_appimage(self):
        cross = "#!/usr/bin/env bash\nmkdir -p bin\necho arm64 > \"$2\"\n"
        packages = ("#!/usr/bin/env bash\nset -eu\n[[ \"$*\" == *ARCH=arm64* ]] || exit 19\n"
                    "case \"$*\" in *generate:deb*) s=.deb ;; *generate:rpm*) s=.rpm ;; *generate:aur*) s=.pkg.tar.zst ;; *) exit 18 ;; esac\n"
                    "echo artifact > \"bin/smeggtuner$s\"\n")
        result, files, sums = self.run_release("linux/arm64", linux=packages, cross=cross)
        self.assertEqual(result.returncode, 0, result.stderr)
        names = ["smeggtuner-1.2.3-linux-arm64" + suffix for suffix in (".deb", ".pkg.tar.zst", ".rpm", ".tar.zst")]
        self.assertEqual(files, ["SHA256SUMS", *names])
        self.assertEqual(sorted(line.split()[1] for line in sums.splitlines()), names)

    def test_no_target_is_a_usage_error(self):
        result, _, _ = self.run_release("")
        self.assertEqual(result.returncode, 2, result.stdout)

    def test_successful_builder_must_supply_artifacts(self):
        for target in ("linux/amd64", "windows/amd64", "android/arm64"):
            with self.subTest(target=target):
                result, _, _ = self.run_release(target, linux="#!/bin/sh\nexit 0\n",
                                                windows="#!/bin/sh\nexit 0\n", android="#!/bin/sh\nexit 0\n")
                self.assertNotEqual(result.returncode, 0, result.stdout)

    def test_every_linux_artifact_is_required_and_nonempty(self):
        for suffix in ("", ".deb", ".rpm", ".pkg.tar.zst", "-amd64.AppImage"):
            for action in ("rm -f", ": >"):
                with self.subTest(suffix=suffix, action=action):
                    result, _, _ = self.run_release("linux/amd64", linux=LINUX_BUILD +
                                                    f"{action} bin/smeggtuner{suffix}\n")
                    self.assertNotEqual(result.returncode, 0, result.stdout)

    def test_ambiguous_appimages_are_rejected(self):
        result, _, _ = self.run_release("linux/amd64", linux=LINUX_BUILD +
                                        "echo extra > bin/smeggtuner-other.AppImage\n")
        self.assertNotEqual(result.returncode, 0, result.stdout)

    def test_windows_requires_exactly_one_program_and_installer(self):
        for extra in ("rm bin/lab-out/win11/*installer.exe\n",
                      "rm bin/lab-out/win11/smeggtuner.exe\n",
                      "echo extra > bin/lab-out/win11/extra.exe\n",
                      ": > bin/lab-out/win11/smeggtuner.exe\n"):
            with self.subTest(extra=extra):
                result, _, _ = self.run_release("windows/amd64", windows=WINDOWS_BUILD + extra)
                self.assertNotEqual(result.returncode, 0, result.stdout)

    def test_collection_failures_are_fatal(self):
        for target, command in (("linux/amd64", "mv"), ("android/arm64", "cp"), ("windows/amd64", "mv"),
                                ("linux/amd64", "sha256sum")):
            with self.subTest(target=target, command=command):
                result, _, _ = self.run_release(target, fail_tool=command)
                self.assertNotEqual(result.returncode, 0, result.stdout)

    def test_a_failed_move_marks_its_own_target_failed(self):
        result, _, _ = self.run_release("linux/amd64", fail_tool="mv")
        self.assertIn("failed: linux/amd64", result.stderr)

    def test_release_depends_on_checks(self):
        source = (ROOT / "justfile").read_text()
        self.assertRegex(source, r"(?m)^release .*: check$")


if __name__ == "__main__":
    unittest.main()
