#!/usr/bin/env python3
# scripts/check-generated.py
"""Compare committed-output candidates with a fresh build, without requiring Git history."""

from pathlib import Path
import re
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
FILES = (
    'backend/common/config/cue_types_config_gen.go',
    'backend/common/i18n/keys_gen.go',
    'backend/common/i18n/messages_gen.go',
    'backend/common/i18n/loc-manifest.json',
    'build/appicon.png',
    'build/appicon.icon/Assets/smeggtuner-cat.svg',
)
TREES = (
    'backend/common/i18n/locales',
    'backend/internal/trayicons',
    'frontend/app/composables/s99wails',
    'frontend/bindings',
    'frontend/i18n/generated',
)


def generated(root):
    paths = {root / path for path in FILES if (root / path).is_file()}
    for directory in TREES:
        paths.update(path for path in (root / directory).rglob('*') if path.is_file()
                     and path.name not in ('doc.go', 'icons_test.go'))
    return {path.relative_to(root).as_posix(): path.read_bytes() for path in paths}


def differences(before, after):
    return sorted(path for path in before.keys() | after.keys() if before.get(path) != after.get(path))


def copy_sources(source, destination):
    excluded = {'.git', 'node_modules', '.nuxt', '.output', '__pycache__'}
    def ignore(directory, names):
        relative = Path(directory).relative_to(source)
        skip = set(names) & excluded
        if relative == Path('.'):
            skip |= {'bin', 'dist', 'sounds', 'reverse-engineering', 'testlab', 'smeggtuner.db', 'config.json'}
        if relative == Path('build'):
            skip.add('android')
        if relative == Path('backend'):
            skip.add('frontend')
        return skip
    shutil.copytree(source, destination, ignore=ignore, symlinks=True)


def check(source=ROOT):
    cache = Path.home() / '.cache' / 'smeggtuner-checks'
    cache.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix='generated-', dir=cache) as directory:
        parent = Path(directory)
        work = parent / source.name
        copy_sources(source, work)
        # Go replacements and the linked UI layer keep their sibling layout.
        for name in set(re.findall(r'=> \.\./([^/\s]+)', (work / 'go.mod').read_text())) | {'S99AB'}:
            (parent / name).symlink_to(source.parent / name, target_is_directory=True)
        before = generated(work)
        for path in before:
            (work / path).unlink()
        subprocess.run(['bash', 'scripts/generate.sh'], cwd=work, check=True)
        drift = differences(before, generated(work))
        if drift:
            raise SystemExit('ERR: generated output is stale; run just generate:\n' + '\n'.join(drift))
    print('Generated output is current (worktree unchanged).')


if __name__ == '__main__':
    check()
