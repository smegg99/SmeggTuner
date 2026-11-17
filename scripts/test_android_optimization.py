# scripts/test_android_optimization.py
"""Reject release packages that would restore the Play Console optimization warning."""

import copy
import importlib.util
import io
import json
from pathlib import Path
import unittest
import zipfile


ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location("android_preflight", ROOT / "build/android-support/preflight.py")
PREFLIGHT = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(PREFLIGHT)
METADATA = {
    "options": {name: True for name in ("isObfuscationEnabled", "isOptimizationsEnabled", "isShrinkingEnabled")},
    "stats": {"noObfuscationPercentage": 11.17, "noOptimizationPercentage": 14.63, "noShrinkingPercentage": 11.56},
}
MARKER = b'~~R8{"compilation-mode":"release","r8-mode":"full"}\0'


class OptimizationTests(unittest.TestCase):
    def verify(self, metadata=METADATA, marker=MARKER, mapping=b"com.example.Class -> a:", mode="release", bundle=True):
        with zipfile.ZipFile(io.BytesIO(), "w") as archive:
            archive.writestr("base/dex/classes.dex" if bundle else "classes.dex", marker)
            if metadata is not None:
                archive.writestr("BUNDLE-METADATA/com.android.tools/r8.json", json.dumps(metadata))
                archive.writestr("BUNDLE-METADATA/com.android.tools.build.obfuscation/proguard.map", mapping)
            PREFLIGHT.verify_optimization(archive, bundle, mode)

    def test_optimized_bundle_and_apk_pass(self):
        self.verify()
        self.verify(metadata=None, bundle=False)

    def test_debug_builds_do_not_need_r8(self):
        self.verify(metadata=None, marker=b"~~D8{}\0", mode="debug")

    def test_unoptimized_release_dex_is_rejected(self):
        with self.assertRaisesRegex(RuntimeError, "not built with R8"):
            self.verify(marker=b"~~D8{}\0")

    def test_r8_compatibility_mode_is_rejected(self):
        with self.assertRaisesRegex(RuntimeError, "full mode"):
            self.verify(marker=MARKER.replace(b'"full"', b'"compatibility"'))

    def test_missing_metadata_and_mapping_are_rejected(self):
        with self.assertRaisesRegex(RuntimeError, "missing R8 metadata"):
            self.verify(metadata=None)
        with self.assertRaisesRegex(RuntimeError, "empty deobfuscation mapping"):
            self.verify(mapping=b"")

    def test_disabling_any_optimization_is_rejected(self):
        for option in METADATA["options"]:
            metadata = copy.deepcopy(METADATA)
            metadata["options"][option] = False
            with self.subTest(option=option), self.assertRaisesRegex(RuntimeError, option):
                self.verify(metadata=metadata)

    def test_each_play_score_must_reach_threshold(self):
        for metric in METADATA["stats"]:
            for value in (75.01, None, "11", -1, float("nan")):
                metadata = copy.deepcopy(METADATA)
                metadata["stats"][metric] = value
                with self.subTest(metric=metric, value=value), self.assertRaisesRegex(RuntimeError, metric):
                    self.verify(metadata=metadata)


if __name__ == "__main__":
    unittest.main()
