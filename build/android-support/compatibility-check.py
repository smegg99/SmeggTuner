#!/usr/bin/env python3
# build/android-support/compatibility-check.py
"""Verify the native recovery flow when an emulator's WebView is unsupported."""
import json
import os
from pathlib import Path
import re
import subprocess
import time
import xml.etree.ElementTree as ET

ROOT = Path(__file__).resolve().parents[2]
SOURCE = Path(__file__).resolve().parent
CONFIG = json.loads((SOURCE / "config.json").read_text())
SERIAL = os.environ.get("ANDROID_SERIAL", "")
if not re.fullmatch(r"emulator-\d+", SERIAL):
    raise SystemExit("Set ANDROID_SERIAL to a dedicated test emulator")
OUT = ROOT / "bin/android/checks" / f"webview-{SERIAL}-{time.time_ns()}"
OUT.mkdir(parents=True)
REPORT = {"serial": SERIAL, "status": "blocked_by_webview", "minimum": CONFIG["webViewMinMajor"], "checks": []}


def adb(*args):
    return subprocess.check_output(["adb", "-s", SERIAL, *args], timeout=30).decode().strip()


def record(name):
    REPORT["checks"].append(name)
    print(f"PASS {name}", flush=True)


def ui():
    temporary = "/data/local/tmp/smeggtuner-webview-check.xml"
    adb("shell", "uiautomator", "dump", temporary)
    content = adb("shell", "cat", temporary)
    adb("shell", "rm", "-f", temporary)
    (OUT / "native-ui.xml").write_text(content + "\n")
    return list(ET.fromstring(content).iter("node"))


def gate():
    nodes = ui()
    messages = [node.get("text", "") for node in nodes if node.get("resource-id") == "android:id/message"]
    expected = str(CONFIG["webViewMinMajor"])
    if not messages or expected not in messages[0]:
        raise RuntimeError("The native WebView requirement screen is missing; use just android check for compatible providers")
    assert CONFIG["applicationId"] + "/com.wails.app.LauncherActivity" in adb("shell", "dumpsys", "activity", "activities")
    REPORT["message"] = messages[0]
    return nodes


def tap(nodes, resource):
    node = next(node for node in nodes if node.get("resource-id") == resource)
    bounds = list(map(int, re.findall(r"\d+", node.get("bounds", ""))))
    assert len(bounds) == 4, "Native control has no bounds"
    adb("shell", "input", "tap", str((bounds[0] + bounds[2]) // 2), str((bounds[1] + bounds[3]) // 2))


try:
    REPORT["api"] = adb("shell", "getprop", "ro.build.version.sdk")
    nodes = gate()
    (OUT / "webview-required.png").write_bytes(subprocess.check_output(["adb", "-s", SERIAL, "exec-out", "screencap", "-p"], timeout=15))
    record("unsupported provider is explained before the frontend or permission prompt")
    tap(nodes, "android:id/button3")
    nodes = gate()
    record("retry keeps an unsupported provider blocked")
    tap(nodes, "android:id/button1")
    end = time.monotonic() + 10
    while time.monotonic() < end:
        activities = adb("shell", "dumpsys", "activity", "activities")
        resumed = [line for line in activities.splitlines() if re.search(r"ResumedActivity|FocusedActivity", line)]
        if resumed and not any(CONFIG["applicationId"] + "/" in line for line in resumed):
            break
        time.sleep(0.25)
    else:
        raise RuntimeError("Update did not open the provider's store or settings")
    REPORT["updater"] = resumed
    ui()
    record("update opens provider recovery")
    adb("shell", "input", "keyevent", "KEYCODE_BACK")
    end = time.monotonic() + 10
    while time.monotonic() < end:
        try:
            nodes = gate()
            break
        except RuntimeError:
            time.sleep(0.25)
    else:
        raise RuntimeError("Returning from provider recovery did not restore the retry screen")
    record("returning from provider recovery retains the retry screen")
    tap(nodes, "android:id/button2")
    end = time.monotonic() + 10
    while time.monotonic() < end:
        activities = adb("shell", "dumpsys", "activity", "activities")
        resumed = [line for line in activities.splitlines() if re.search(r"ResumedActivity|FocusedActivity", line)]
        if resumed and not any(CONFIG["applicationId"] + "/" in line for line in resumed):
            break
        time.sleep(0.25)
    else:
        raise RuntimeError("Close did not leave the requirement screen")
    record("close exits the requirement screen")
except Exception as error:
    REPORT["error"] = str(error)
    raise
finally:
    (OUT / "result.json").write_text(json.dumps(REPORT, indent=2) + "\n")
    (OUT / "logcat.txt").write_text(adb("logcat", "-d", "-t", "1000"))
    print(f"WebView recovery evidence: {OUT}", flush=True)
