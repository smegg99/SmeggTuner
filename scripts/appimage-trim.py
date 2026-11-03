#!/usr/bin/env python3
# scripts/appimage-trim.py
"""Drop the WebKit API the binary does not link from a Wails AppDir and repack the AppImage.

Wails 3.0.0 copies WebKitWebProcess from every WebKit on the build host, so a GTK3 build also
carries webkitgtk-6.0, its JavaScriptCore and GTK4. Usage: appimage-trim.py BUILDDIR OUTDIR
"""

from pathlib import Path
import os
import re
import shutil
import subprocess
import sys

HELPER_DIR = re.compile(r'^webkit2?gtk-[\d.]+$')
NEEDED = re.compile(r'\(NEEDED\)\s+Shared library: \[([^\]]+)\]')


def is_elf(path: Path) -> bool:
    if path.is_symlink() or not path.is_file():
        return False
    with path.open('rb') as f:
        return f.read(4) == b'\x7fELF'


def needed(path: Path) -> list[str]:
    out = subprocess.run(['readelf', '-d', str(path)], capture_output=True, text=True, check=True).stdout
    return NEEDED.findall(out)


def only(paths: list[Path], what: str) -> Path:
    if len(paths) != 1:
        sys.exit(f'ERR: expected one {what}, found {len(paths)}')
    return paths[0]


def main() -> None:
    if len(sys.argv) != 3:
        sys.exit('Usage: appimage-trim.py BUILDDIR OUTDIR')
    build, outdir = Path(sys.argv[1]).resolve(), Path(sys.argv[2]).resolve()
    appdir = only(sorted(build.glob('*.AppDir')), 'AppDir')
    tool = only(sorted(build.glob('linuxdeploy-*.AppImage')), 'linuxdeploy')
    image = outdir / (appdir.stem + '.AppImage')

    # libwebkit2gtk-4.1.so.0 keeps webkit2gtk-4.1/; every other helper directory goes.
    linked = {re.sub(r'^lib|\.so.*$', '', lib) for exe in (appdir / 'usr/bin').iterdir() if is_elf(exe) for lib in needed(exe)}
    stale = [d for d in (appdir / 'usr').rglob('*') if d.is_dir() and HELPER_DIR.match(d.name) and d.name not in linked]
    if not stale:
        print(f'appimage-trim: nothing to trim in {appdir.name}')
        return

    # Only libraries the stale helpers pulled in may go, and only when nothing else links them:
    # GTK's input, print and image modules sit beside them and are loaded without being linked.
    libdirs = [d for d in (appdir / 'usr/lib', appdir / 'usr/lib64', *(appdir / 'usr/lib').glob('*-linux-gnu')) if d.is_dir()]
    by_name = {p.name: p for d in libdirs for p in d.iterdir() if '.so' in p.name}

    def closure(roots: list[Path]) -> set[Path]:
        seen: set[Path] = set()
        queue = list(roots)
        while queue:
            for lib in needed(queue.pop()):
                if lib in by_name and (real := by_name[lib].resolve()) not in seen:
                    seen.add(real)
                    queue.append(real)
        return seen

    pulled = closure([p for d in stale for p in d.rglob('*') if is_elf(p)])
    for d in stale:
        shutil.rmtree(d)
    others = [p for p in appdir.rglob('*') if is_elf(p) and p.resolve() not in pulled]
    unused = pulled - closure(others)
    removed = [p for p in by_name.values() if p.resolve() in unused]
    freed = sum(p.stat().st_size for p in removed if not p.is_symlink())
    for p in removed:
        p.unlink()
    print(f'appimage-trim: removed {", ".join(d.name for d in stale)} and {len(removed)} libraries ({freed / 1e6:.0f} MB)')

    env = os.environ | {'OUTPUT': image.name, 'NO_STRIP': '1'}
    subprocess.run([str(tool), '--appimage-extract-and-run', '--appdir', str(appdir), '--output', 'appimage'],
                   cwd=build, env=env, check=True, stdout=subprocess.DEVNULL)
    shutil.move(build / image.name, image)
    print(f'appimage-trim: {image} ({image.stat().st_size / 1e6:.0f} MB)')


if __name__ == '__main__':
    main()
