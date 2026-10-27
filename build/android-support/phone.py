#!/usr/bin/env python3
# build/android-support/phone.py
"""Inspect one authorized physical Android device before installing the app."""
import json
import os
from pathlib import Path
import re
import subprocess
import sys

CONFIG = json.loads((Path(__file__).with_name("config.json")).read_text())


def adb(*args, serial=None):
    command = ["adb"]
    if serial:
        command.extend(["-s", serial])
    return subprocess.check_output(command + list(args), timeout=20).decode().strip()


def virtual(serial):
    return serial.startswith("emulator-") or any(
        adb("shell", "getprop", key, serial=serial) == "1"
        for key in ("ro.kernel.qemu", "ro.boot.qemu")
    )


def select_phone():
    devices = {}
    for line in adb("devices").splitlines()[1:]:
        fields = line.split()
        if len(fields) >= 2:
            devices[fields[0]] = fields[1]
    requested = os.environ.get("ANDROID_SERIAL")
    if requested:
        state = devices.get(requested, "not connected")
        if state != "device":
            raise ValueError(f"{requested}: {state}. Unlock the phone and accept USB debugging authorization.")
        if virtual(requested):
            raise ValueError("Phone commands require a physical device; use just android check for emulators.")
        return requested
    phones = [serial for serial, state in devices.items() if state == "device" and not virtual(serial)]
    if len(phones) != 1:
        states = ", ".join(f"{serial}: {state}" for serial, state in devices.items()) or "none"
        raise ValueError(f"Connect and authorize one phone, or set ANDROID_SERIAL. ADB devices: {states}")
    return phones[0]


def inspect_phone(serial):
    def prop(name):
        return adb("shell", "getprop", name, serial=serial)

    sdk = int(prop("ro.build.version.sdk"))
    if sdk < CONFIG["minSdk"]:
        raise ValueError(f"Android API {sdk} is below the installation floor {CONFIG['minSdk']}.")
    abis = (prop("ro.product.cpu.abilist") or prop("ro.product.cpu.abi")).split(",")
    arch = next((arch for abi, arch in (("arm64-v8a", "arm64"), ("x86_64", "amd64")) if abi in abis), None)
    if not arch:
        raise ValueError(f"Unsupported phone ABIs: {', '.join(abis)}. APKs require ARM64 or x86_64.")
    webview = adb("shell", "dumpsys", "webviewupdate", serial=serial)
    provider = re.search(r"Current WebView package \(name, version\): \(([^,]+), ([^)]+)\)", webview)
    current = {"package": provider[1], "version": provider[2]} if provider else None
    if not current and sdk >= 24:
        raise ValueError("Enable or update Android System WebView before running the phone check.")
    if current:
        major = int(current["version"].split(".")[0])
        if major < CONFIG["webViewMinMajor"]:
            raise ValueError(f"Update {current['package']} {current['version']} to WebView {CONFIG['webViewMinMajor']} or newer.")
    return {"serial": serial, "model": prop("ro.product.model"), "android": prop("ro.build.version.release"),
            "sdk": sdk, "abis": abis, "arch": arch, "webView": current,
            "webViewMinimum": CONFIG["webViewMinMajor"]}


if __name__ == "__main__":
    try:
        print(json.dumps(inspect_phone(select_phone()), indent=2))
    except (ValueError, subprocess.SubprocessError) as error:
        print(str(error), file=sys.stderr)
        raise SystemExit(1)
