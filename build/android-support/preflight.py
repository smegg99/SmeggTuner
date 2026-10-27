#!/usr/bin/env python3
# build/android-support/preflight.py
"""Validate Android tools and the native and packaged artifacts they produce."""

import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import ssl
import struct
import subprocess
import sys
import tempfile
import zipfile

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "scripts"))
from node_env import node_errors

CONFIG = json.loads((Path(__file__).parent / "config.json").read_text())
ABI = {"arm64": "arm64-v8a", "amd64": "x86_64"}


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def run(*args, cwd=ROOT):
    return subprocess.check_output(args, cwd=cwd, text=True, stderr=subprocess.STDOUT).strip()



def signing():
    names = ("ANDROID_KEYSTORE_FILE", "ANDROID_KEYSTORE_PASSWORD", "ANDROID_KEY_ALIAS", "ANDROID_KEY_PASSWORD")
    missing = [name for name in names if not os.environ.get(name, "").strip()]
    require(not missing, "Release signing requires nonempty environment variables: " + ", ".join(missing))
    path = Path(os.environ[names[0]]).expanduser().resolve()
    require(path.is_file() and os.access(path, os.R_OK), "ANDROID_KEYSTORE_FILE must name a readable keystore file")
    return str(path)


def doctor():
    sdk = Path(os.environ["ANDROID_HOME"])
    ndk = sdk / "ndk" / CONFIG["ndk"]
    tools = sdk / "build-tools" / CONFIG["buildTools"]
    for directory, revision in ((ndk, CONFIG["ndk"]), (tools, CONFIG["buildTools"])):
        props = directory / "source.properties"
        require(props.is_file(), f"Missing pinned SDK component: {directory}")
        require(re.search(r"Pkg.Revision\s*=\s*" + re.escape(revision) + r"\s*$", props.read_text(), re.M), f"Unexpected SDK revision: {props}")
    require((sdk / f"platforms/android-{CONFIG['platform']}/android.jar").is_file(), f"Missing pinned platform android-{CONFIG['platform']}")
    for tool in ("sdkmanager", "avdmanager", "adb", "emulator", "java", "go", "node", "npm", "pnpm", "flock", "timeout", "rsvg-convert"):
        require(shutil.which(tool), f"Missing command: {tool}")
    for path in (tools / "aapt", tools / "apksigner", tools / "zipalign", ndk / "toolchains/llvm/prebuilt/linux-x86_64/bin/clang"):
        require(os.access(path, os.X_OK), f"Missing executable: {path}")
    errors = node_errors(shutil.which("node"))
    require(not errors, errors)
    manifest = json.loads((ROOT / "frontend/package.json").read_text())
    expected = manifest["packageManager"].split("@", 1)[1].split("+", 1)[0]
    actual = run("pnpm", "--version", cwd=ROOT / "frontend")
    require(actual == expected, f"pnpm {actual} does not match frontend packageManager pnpm@{expected}")
    mod = (ROOT / "go.mod").read_text()
    minimum = re.search(r"^go ([\d.]+)$", mod, re.M)[1]
    actual_go = run("go", "env", "GOVERSION")
    require(tuple(map(int, re.search(r"go(\d+\.\d+)", actual_go)[1].split('.'))) >= tuple(map(int, minimum.split('.')[:2])), f"Go {actual_go} does not satisfy go {minimum}")
    pinned = re.search(r"github.com/wailsapp/wails/v3\s+(v\S+)", mod)[1]
    selected = json.loads(run("go", "list", "-m", "-json", "github.com/wailsapp/wails/v3"))
    require(selected["Version"] == pinned and "Replace" not in selected, f"Wails module must use pinned {pinned} without replacement")
    require(pinned.removeprefix("v") in run("go", "tool", "wails3", "version"), f"Wails tool does not match pinned {pinned}")
    for relative in re.findall(r"=>\s+(\.\.?/\S+)", mod):
        require((ROOT / relative / "go.mod").is_file(), f"Missing sibling Go module: {relative}")
    for value in manifest.get("dependencies", {}).values():
        if value.startswith("link:"):
            require((ROOT / "frontend" / value[5:] / "package.json").is_file(), f"Missing sibling frontend: {value}")
    print(f"Android preflight passed: {actual_go}, Wails {pinned}, Node {run('node', '--version')}, pnpm {actual}")


def native(path, arch, mode):
    metadata = run("go", "version", "-m", str(path))
    for value in (f"GOARCH={arch}", "GOOS=android", "-buildmode=c-shared", "CGO_ENABLED=1"):
        require(value in metadata, f"{path}: Go metadata missing {value}")
    tags = re.search(r"\tbuild\t-tags=(.*)", metadata)
    require(tags, f"{path}: missing Go build tags")
    tags = set(tags[1].strip('"').split(","))
    expected, forbidden = ("debug", "production") if mode == "debug" else ("production", "debug")
    require(expected in tags and forbidden not in tags, f"{path}: native build tags do not match {mode}")
    with open(path, "rb") as stream:
        header = stream.read(64)
        require(header[:6] == b"\x7fELF\x02\x01", f"{path}: expected little-endian ELF64")
        require(struct.unpack_from("<H", header, 18)[0] == {"arm64": 183, "amd64": 62}[arch], f"{path}: ELF machine does not match {arch}")
        offset = struct.unpack_from("<Q", header, 32)[0]
        size, count = struct.unpack_from("<HH", header, 54)
        loads = 0
        for index in range(count):
            stream.seek(offset + index * size)
            segment = stream.read(size)
            if struct.unpack_from("<I", segment)[0] == 1:
                loads += 1
                fileoff, address = struct.unpack_from("<QQ", segment, 8)
                alignment = struct.unpack_from("<Q", segment, 48)[0]
                require(alignment >= 16384 and (address - fileoff) % 16384 == 0, f"{path}: ELF LOAD is not 16 KB aligned")
        require(loads, f"{path}: ELF has no LOAD segments")


def fields(data):
    # AAPT2 Resources.proto wire fields, including the AAB XmlNode manifest.
    index = 0
    def varint():
        nonlocal index
        value = 0
        for shift in range(0, 70, 7):
            byte = data[index]
            index += 1
            value |= (byte & 127) << shift
            if byte < 128:
                return value
        raise RuntimeError("Invalid protobuf varint")
    result = {}
    while index < len(data):
        key = varint()
        wire = key & 7
        if wire == 0:
            value = varint()
        else:
            length = varint() if wire == 2 else {1: 8, 5: 4}[wire]
            value = data[index:index + length]
            require(len(value) == length, "Truncated protobuf field")
            index += length
        result.setdefault(key >> 3, []).append(value)
    return result


def bundle_manifest(data):
    element = fields(fields(data)[1][0])
    attributes = {}
    for item in element.get(4, []):
        attr = fields(item)
        attributes[attr[2][0].decode()] = attr.get(3, [b""])[0].decode()
    yield element[3][0].decode(), attributes
    for child in element.get(5, []):
        if 1 in fields(child):
            yield from bundle_manifest(child)


def verify_signing(path, bundle, tools, mode):
    if bundle:
        output = run("jarsigner", "-J-Duser.language=en", "-J-Duser.country=US", "-verify", str(path))
        require("jar verified." in output, "AAB signature could not be verified")
    else:
        output = run(str(tools / "apksigner"), "verify", "--verbose", "--print-certs", str(path))
    if mode != "release":
        return
    expected = hashlib.sha256(subprocess.check_output([
        "keytool", "-exportcert", "-keystore", signing(),
        "-alias", os.environ["ANDROID_KEY_ALIAS"],
        "-storepass:env", "ANDROID_KEYSTORE_PASSWORD",
    ], stderr=subprocess.PIPE)).hexdigest()
    if bundle:
        certificates = re.findall(r"-----BEGIN CERTIFICATE-----.*?-----END CERTIFICATE-----", run("keytool", "-printcert", "-jarfile", str(path), "-rfc"), re.S)
        require(certificates, "AAB has no signing certificate")
        actual = hashlib.sha256(ssl.PEM_cert_to_DER_cert(certificates[0])).hexdigest()
    else:
        digests = {value.lower() for value in re.findall(r"^.* certificate SHA-256 digest: ([0-9a-fA-F]+)$", output, re.M)}
        require("Number of signers: 1" in output and len(digests) == 1, "APK must have exactly one signer")
        actual = digests.pop()
    require(actual == expected, "Artifact signing certificate does not match ANDROID_KEY_ALIAS")


def verify(path, arch, mode, staged=None):
    require(arch in ABI and mode in ("debug", "release"), "Expected arm64|amd64 and debug|release")
    version = (ROOT / "VERSION").read_text().strip()
    major, minor, patch = map(int, version.split("."))
    code = str(major * 1000000 + minor * 1000 + patch)
    tools = Path(os.environ["ANDROID_HOME"]) / "build-tools" / CONFIG["buildTools"]
    bundle = path.suffix == ".aab"
    require(path.suffix in (".apk", ".aab"), "Artifact must be APK or AAB")
    with zipfile.ZipFile(path) as archive:
        libs = [item for item in archive.namelist() if item.endswith(".so")]
        member = f"{'base/' if bundle else ''}lib/{ABI[arch]}/libwails.so"
        require(libs == [member], f"{path}: expected only {member}, found {libs}")
        with tempfile.TemporaryDirectory(prefix="smegg-android-verify-") as temporary:
            library = Path(temporary) / "libwails.so"
            library.write_bytes(archive.read(member))
            native(library, arch, mode)
            if staged:
                require(hashlib.sha256(library.read_bytes()).digest() == hashlib.sha256(staged.read_bytes()).digest(), "Packaged native library differs from staged build")
        if bundle:
            manifest = dict(bundle_manifest(archive.read("base/manifest/AndroidManifest.xml")))
            attrs = manifest["manifest"]
            require((attrs.get("package"), attrs.get("versionName"), attrs.get("versionCode")) == (CONFIG["applicationId"], version, code), "AAB package ID/version mismatch")
            sdk = manifest["uses-sdk"]
            require((sdk.get("minSdkVersion"), sdk.get("targetSdkVersion")) == (str(CONFIG["minSdk"]), str(CONFIG["targetSdk"])), "AAB min/target SDK mismatch")
            require((manifest["application"].get("debuggable") == "true") == (mode == "debug"), "AAB debuggable mode mismatch")
            print("AAB verified; ZIP page alignment applies to APKs generated from the bundle.")
        else:
            badging = run(str(tools / "aapt"), "dump", "badging", str(path))
            for text in (f"name='{CONFIG['applicationId']}'", f"versionCode='{code}'", f"versionName='{version}'", f"sdkVersion:'{CONFIG['minSdk']}'", f"targetSdkVersion:'{CONFIG['targetSdk']}'", f"native-code: '{ABI[arch]}'"):
                require(text in badging, f"{path}: APK metadata missing {text}")
            require(("application-debuggable" in badging) == (mode == "debug"), "APK debuggable mode mismatch")
            run(str(tools / "zipalign"), "-c", "-P", "16", "4", str(path))
    verify_signing(path, bundle, tools, mode)
    print(f"Verified {path}: {CONFIG['applicationId']} {version}, {mode}, {arch}, SDK {CONFIG['minSdk']}..{CONFIG['targetSdk']}")


if __name__ == "__main__":
    try:
        command, *args = sys.argv[1:]
        if command == "doctor":
            doctor()
        elif command == "signing":
            print(signing())
        elif command == "native":
            native(Path(args[0]), args[1], args[2])
        elif command == "verify":
            verify(Path(args[0]), args[1], args[2], Path(args[3]) if len(args) > 3 else None)
        else:
            raise RuntimeError("Usage: preflight.py doctor|signing|native PATH ARCH MODE|verify PATH ARCH MODE [STAGED_LIBRARY]")
    except (RuntimeError, OSError, ValueError, KeyError, IndexError, zipfile.BadZipFile, subprocess.CalledProcessError) as error:
        detail = error.output if isinstance(error, subprocess.CalledProcessError) else str(error)
        print(f"Android preflight: {detail}", file=sys.stderr)
        sys.exit(1)
