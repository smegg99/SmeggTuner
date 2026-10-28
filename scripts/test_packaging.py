#!/usr/bin/env python3
# scripts/test_packaging.py
"""Source-only packaging checks; no Docker daemon, installer or application runs."""

import json
import os
from pathlib import Path
import plistlib
import subprocess
import tempfile
import unittest

import cross_build


ROOT = Path(__file__).resolve().parents[1]


class CrossBuildTests(unittest.TestCase):
    def test_go_parser_handles_block_single_line_versions_spaces_and_absolute_paths(self):
        with tempfile.TemporaryDirectory(prefix="smegg cross ") as directory:
            root = Path(directory) / "app"
            root.mkdir()
            relative = root.parent / "shared module"
            absolute = root.parent / "absolute module"
            inside = root / "internal-module"
            for path in (relative, absolute, inside):
                path.mkdir()
            source = ('module test.invalid/app\n\ngo 1.27\n'
                      'replace (\n test.invalid/relative => "../shared module"\n'
                      f' test.invalid/absolute => {json.dumps(str(absolute))}\n'
                      ' test.invalid/remote => test.invalid/replacement v1.2.3\n)\n'
                      'replace test.invalid/inside v1.0.0 => ./internal-module\n')
            (root / "go.mod").write_text(source)
            command = cross_build.build_command(root, "cross-image", "linux", "arm64", "tuner", "gtk3")
            volumes = [command[i + 1] for i, item in enumerate(command) if item == "--volume"]
            for want in (f"{relative}:/shared module:ro", f"{absolute}:{absolute}:ro",
                         f"{inside}:/app/internal-module:ro"):
                self.assertIn(want, volumes)
            self.assertEqual(command[-3:], ["cross-image", "linux", "arm64"])
            self.assertEqual(command[command.index("--workdir") + 1], "/app/backend")
            self.assertIn("GOTOOLCHAIN=auto", command)
            self.assertEqual((root / "go.mod").read_text(), source)
            self.assertFalse((root / "go.sum").exists())

    def test_missing_local_module_fails_before_docker(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "go.mod").write_text('module test.invalid/app\nreplace test.invalid/lib => ./missing\n')
            with self.assertRaises(FileNotFoundError):
                cross_build.build_command(root, "cross-image", "linux", "amd64", "tuner", "")

    def test_cli_passes_literal_arguments_and_propagates_builder_failure(self):
        with tempfile.TemporaryDirectory(prefix="smegg cross ") as directory:
            root = Path(directory)
            (root / "go.mod").write_text("module test.invalid/app\ngo 1.27\n")
            docker = root / "docker"
            docker.write_text('#!/bin/sh\nprintf "%s\\n" "$@" > docker-args\nexit 23\n')
            docker.chmod(0o755)
            result = subprocess.run(["python3", str(ROOT / "scripts/cross_build.py"),
                                     "--root", str(root), "--image", "cross-image", "--os", "windows",
                                     "--arch", "amd64", "--app", "tuner; touch injected", "--tags", ""],
                                    env=os.environ | {"PATH": str(root) + os.pathsep + os.environ["PATH"]},
                                    capture_output=True, text=True, timeout=30)
            self.assertNotEqual(result.returncode, 0)
            self.assertIn("APP_NAME=tuner; touch injected", (root / "docker-args").read_text().splitlines())
            self.assertFalse((root / "injected").exists())


class PackageSourceTests(unittest.TestCase):
    def test_macos_microphone_reason_is_present_in_both_bundles(self):
        for name in ("Info.plist", "Info.dev.plist"):
            with (ROOT / "build/darwin" / name).open("rb") as source:
                info = plistlib.load(source)
            self.assertIn("microphone", info["NSMicrophoneUsageDescription"])

    def test_nsis_marker_matches_runtime_and_uninstall_preserves_unknown_files(self):
        source = (ROOT / "build/windows/nsis/project.nsi").read_text()
        self.assertIn('FileOpen $0 "$INSTDIR\\smeggtuner.install" w', source)
        self.assertIn('FileWrite $0 "nsis$\\r$\\n"', source)
        uninstall = source[source.index('Section "uninstall"'):]
        self.assertNotRegex(uninstall, r'(?i)RMDir\s+/r\s+"?\$INSTDIR')
        self.assertIn('Delete "$INSTDIR\\${PRODUCT_EXECUTABLE}"', uninstall)
        self.assertIn('RMDir "$INSTDIR"', uninstall)
        macros = (ROOT / "build/windows/nsis/wails_tools.nsh").read_text()
        delete_uninstaller = macros.split("!macro wails.deleteUninstaller", 1)[1].split("!macroend", 1)[0]
        self.assertNotRegex(delete_uninstaller, r"(?i)RMDir")
        self.assertIn('Delete "$INSTDIR\\uninstall.exe"', delete_uninstaller)


if __name__ == "__main__":
    unittest.main()
