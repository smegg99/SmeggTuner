#!/usr/bin/env python3
# scripts/sync-privacy.py
"""Copy the bundled policy to the website, or verify that its copy is current."""
import argparse
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("site", type=Path)
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    destination = args.site.resolve() / "web/app/content/privacy"
    if not destination.is_dir():
        parser.error(f"Privacy content directory does not exist: {destination}")
    manifest = ROOT / "backend/common/i18n/loc-manifest.json"
    languages = json.loads(manifest.read_text())["locales"]
    policies = {code: ROOT / ("PRIVACY.md" if code == "en" else f"PRIVACY.{code}.md")
                for code in languages}
    for path in policies.values():
        if not path.is_file():
            parser.error(f"Missing translated privacy policy: {path.name}")
    for locale, path in policies.items():
        name = path.name
        source = path.read_text()
        content = f"<!-- Generated from SmeggTuner/{name} by scripts/sync-privacy.py. -->\n\n" + source
        target = destination / f"{locale}.md"
        if args.check:
            if not target.is_file() or target.read_text() != content:
                parser.error(f"Website policy differs from {name}: run just privacy-site")
        else:
            target.write_text(content)
        print(f"{'Verified' if args.check else 'Synced'} {locale} privacy policy")


if __name__ == "__main__":
    main()
