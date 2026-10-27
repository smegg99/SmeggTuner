#!/usr/bin/env python3
# build/android-support/webview-fixture.py
"""Install the pinned AOSP WebView baseline on an API 21/23/28 SDK test emulator."""
import base64
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import urllib.request
import zipfile

SOURCE = Path(__file__).resolve().parent
CONFIG = json.loads((SOURCE / "config.json").read_text())
SERIAL = os.environ.get("ANDROID_SERIAL", "")
COMMIT = "3a8d2089900522ca542c39055d51b08188fb11cb"
URL = f"https://android.googlesource.com/platform/external/chromium-webview/+/{COMMIT}/prebuilt/x86_64/webview.apk?format=TEXT"
SHA256 = "30c0dab7f13ad96405c61454c33021ca479755947d62d05f058f88f4bde2ea61"
CERTIFICATE = "a40da80a59d170caa950cf15c18c454d47a39b26989d8b640ecd745ba71bf5dc"
VERSION = "95.0.4638.74"
CACHE = Path(os.environ.get("XDG_CACHE_HOME", Path.home() / ".cache")) / "smeggtuner/android-webview-donor"
SIGNER = Path(os.environ["ANDROID_HOME"]) / "build-tools" / CONFIG["buildTools"] / "apksigner"


def adb(*args):
    return subprocess.check_output(["adb", "-s", SERIAL, *args], timeout=120).decode().strip()


def certificate(pem):
    der = subprocess.check_output(["openssl", "x509", "-outform", "DER"], input=pem)
    return hashlib.sha256(der).hexdigest()


if not re.fullmatch(r"emulator-\d+", SERIAL):
    raise SystemExit("Set ANDROID_SERIAL to a dedicated API 21/23/28 SDK emulator")
if adb("shell", "getprop", "ro.kernel.qemu") != "1":
    raise SystemExit("WebView fixture installation requires a virtual SDK device")
api = adb("shell", "getprop", "ro.build.version.sdk")
if api not in ("21", "23", "28") or adb("shell", "getprop", "ro.product.cpu.abi") != "x86_64":
    raise SystemExit("This fixture is only for the API 21/23/28 AOSP-provider x86_64 matrix images")
installed = adb("shell", "pm", "path", "com.android.webview")
if not installed.startswith("package:") or "\n" in installed:
    raise SystemExit("The SDK image must provide com.android.webview")
with tempfile.TemporaryDirectory(prefix="smeggtuner-webview-fixture-") as temporary:
    working = Path(temporary)
    stock = working / "installed.apk"
    adb("pull", installed.removeprefix("package:"), str(stock))
    with zipfile.ZipFile(stock) as archive:
        signature = archive.read("META-INF/CERT.RSA")
    pem = subprocess.check_output(["openssl", "pkcs7", "-inform", "DER", "-print_certs"], input=signature)
    if certificate(pem) != CERTIFICATE:
        raise SystemExit("The installed provider is not signed by the AOSP emulator development key")
    CACHE.mkdir(parents=True, exist_ok=True)
    original = CACHE / "aosp-webview95.apk"
    if not original.exists() or hashlib.sha256(original.read_bytes()).hexdigest() != SHA256:
        download = working / "webview.apk"
        print("Fetching pinned AOSP WebView 95", flush=True)
        with urllib.request.urlopen(URL, timeout=60) as response, download.open("wb") as output:
            pending = b""
            while part := response.read(4 << 20):
                pending += part
                boundary = len(pending) // 4 * 4
                output.write(base64.b64decode(pending[:boundary], validate=True))
                pending = pending[boundary:]
            if pending:
                output.write(base64.b64decode(pending, validate=True))
        if hashlib.sha256(download.read_bytes()).hexdigest() != SHA256:
            raise SystemExit("AOSP WebView download hash mismatch")
        original.write_bytes(download.read_bytes())
    # AOSP's Android.bp uses default_dev_cert. Reproduce SDK image build signing with its public test key.
    security = "https://android.googlesource.com/platform/build/+/refs/tags/android-12.1.0_r1/target/product/security/"
    for name in ("testkey.pk8", "testkey.x509.pem"):
        with urllib.request.urlopen(security + name + "?format=TEXT", timeout=30) as response:
            target = working / name
            target.write_bytes(base64.b64decode(response.read(), validate=True))
            target.chmod(0o600)
    if certificate((working / "testkey.x509.pem").read_bytes()) != CERTIFICATE:
        raise SystemExit("AOSP development certificate hash mismatch")
    signed = working / "webview-emulator.apk"
    subprocess.run([str(SIGNER), "sign", "--key", str(working / "testkey.pk8"), "--cert",
                    str(working / "testkey.x509.pem"), "--out", str(signed), str(original)], check=True)
    verification = subprocess.check_output([str(SIGNER), "verify", "--verbose", "--print-certs", str(signed)]).decode()
    digests = {line.split(": ")[-1] for line in verification.splitlines() if "certificate SHA-256 digest:" in line}
    if digests != {CERTIFICATE}:
        raise SystemExit("Fixture signature verification failed")
    adb("shell", "am", "force-stop", CONFIG["applicationId"])
    result = adb("install", "-r", str(signed))
    if "Success" not in result or "Failure" in result:
        raise SystemExit(result)
    metadata = adb("shell", "dumpsys", "package", "com.android.webview")
    if not re.search(r"versionName=" + re.escape(VERSION) + r"\s", metadata):
        raise SystemExit("Installed provider version does not match the pinned fixture")
    report = {"serial": SERIAL, "api": api, "version": VERSION, "source": URL,
              "sourceSha256": SHA256, "emulatorCertificateSha256": CERTIFICATE}
    evidence = SOURCE.parents[1] / "bin/android/checks" / f"webview-fixture-{SERIAL}.json"
    evidence.parent.mkdir(parents=True, exist_ok=True)
    evidence.write_text(json.dumps(report, indent=2) + "\n")
    print(f"Installed emulator fixture {VERSION}; run just android check next. Evidence: {evidence}")
