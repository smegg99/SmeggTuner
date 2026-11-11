#!/usr/bin/env python3
# scripts/test_android_signing.py
"""Check signing defaults and unattended failures without a real signing key."""

import os
from pathlib import Path
import subprocess
import tempfile
import unittest

SCRIPT = Path(__file__).resolve().parents[1] / "build/android-support/signing.sh"


class AndroidSigningTests(unittest.TestCase):
    def invoke(self, password=None, key_password=None, readable=True):
        with tempfile.TemporaryDirectory(prefix="smegg-signing-test-") as directory:
            key = Path(directory) / "upload key.jks"
            if readable:
                key.touch()
            environment = {k: v for k, v in os.environ.items() if not k.startswith("ANDROID_KEY")}
            environment["ANDROID_KEYSTORE_FILE"] = str(key)
            if password is not None:
                environment["ANDROID_KEYSTORE_PASSWORD"] = password
            if key_password is not None:
                environment["ANDROID_KEY_PASSWORD"] = key_password
            result = subprocess.run([
                "bash", "-euc",
                'source "$1"; android_signing; '
                '[[ "$ANDROID_KEY_ALIAS" == smeggtuner ]]; '
                '[[ "$ANDROID_KEY_PASSWORD" == "${2:-$ANDROID_KEYSTORE_PASSWORD}" ]]',
                "signing-test", str(SCRIPT), key_password or "",
            ], env=environment, stdin=subprocess.DEVNULL, capture_output=True, text=True, timeout=5)
            self.assertNotIn("test-password", result.stdout + result.stderr)
            return result

    def test_store_password_is_used_for_key(self):
        self.assertEqual(self.invoke("test-password").returncode, 0)

    def test_separate_key_password_is_preserved(self):
        self.assertEqual(self.invoke("test-password", "separate-key").returncode, 0)

    def test_unattended_build_requires_password(self):
        result = self.invoke()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Run just android-release in a terminal", result.stderr)

    def test_missing_key_fails_before_password_prompt(self):
        result = self.invoke(readable=False)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("keystore is not readable", result.stderr)
        self.assertNotIn("password:", result.stderr)


if __name__ == "__main__":
    unittest.main()
