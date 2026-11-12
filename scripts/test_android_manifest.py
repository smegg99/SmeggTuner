# scripts/test_android_manifest.py
"""Reject packaged capabilities that are outside the app's privacy declaration."""

import importlib.util
from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location("android_preflight", ROOT / "build/android-support/preflight.py")
PREFLIGHT = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(PREFLIGHT)


class PackagedPermissionTests(unittest.TestCase):
    def permissions(self):
        return [
            "android.permission.INTERNET",
            "android.permission.ACCESS_NETWORK_STATE",
            "android.permission.RECORD_AUDIO",
            "me.smegg.smeggtuner.DYNAMIC_RECEIVER_NOT_EXPORTED_PERMISSION",
        ]

    def test_tuner_and_private_receiver_permissions_are_accepted(self):
        PREFLIGHT.verify_permissions(self.permissions())

    def test_unused_template_and_sdk_permissions_are_rejected(self):
        for permission in ["CAMERA", "ACCESS_FINE_LOCATION", "USE_FINGERPRINT", "POST_NOTIFICATIONS", "FOREGROUND_SERVICE"]:
            with self.subTest(permission=permission), self.assertRaisesRegex(RuntimeError, "unexpected="):
                PREFLIGHT.verify_permissions([*self.permissions(), "android.permission." + permission])

    def test_advertising_identifier_is_rejected(self):
        with self.assertRaisesRegex(RuntimeError, "AD_ID"):
            PREFLIGHT.verify_permissions([*self.permissions(), "com.google.android.gms.permission.AD_ID"])

    def test_missing_microphone_permission_is_rejected(self):
        with self.assertRaisesRegex(RuntimeError, "missing=.*RECORD_AUDIO"):
            PREFLIGHT.verify_permissions(p for p in self.permissions() if not p.endswith("RECORD_AUDIO"))


if __name__ == "__main__":
    unittest.main()
