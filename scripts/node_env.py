#!/usr/bin/env python3
"""Run repository tools with a compatible Node, without changing the shell default."""

import os
from pathlib import Path
import shutil
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def run(*args):
    return subprocess.check_output(args, cwd=ROOT, text=True, stderr=subprocess.STDOUT).strip()


def node_errors(node):
    npm = shutil.which("npm")
    require(npm, "npm is required for its bundled semver engine checker")
    script = r"""
const fs = require('fs'), path = require('path');
const {createRequire} = require('module');
const semver = createRequire(fs.realpathSync(process.argv[2]))('semver');
const frontend = path.join(process.argv[1], 'frontend');
const manifest = JSON.parse(fs.readFileSync(path.join(frontend, 'package.json')));
const paths = [path.join(frontend, 'package.json'), ...Object.keys({
  ...manifest.dependencies, ...manifest.devDependencies
}).map(name => path.join(frontend, 'node_modules', name, 'package.json'))];
for (const file of paths) {
  if (!fs.existsSync(file)) continue;
  const pkg = JSON.parse(fs.readFileSync(file)), range = pkg.engines?.node;
  if (range && !semver.satisfies(process.versions.node, range))
    console.log(`Node ${process.versions.node} is incompatible with ${pkg.name} ${pkg.version || ''}: engines.node ${range}`);
}
"""
    return run(str(node), "-e", script, str(ROOT), npm)


def node_bin():
    current = shutil.which("node")
    require(current, "node is missing from PATH")
    errors = node_errors(current)
    if not errors:
        return ""
    candidate = Path.home() / ".local/share/fnm/node-versions/v22.22.3/installation/bin/node"
    require(candidate.is_file() and not node_errors(candidate), errors + f"\nUse a compatible scoped Node installation: {candidate.parent}")
    print(f"Using {run(str(candidate), '--version')} for this command; shell Node remains unchanged.", file=sys.stderr)
    return str(candidate.parent)


if __name__ == "__main__":
    try:
        selected = node_bin()
        if sys.argv[1:] == ["--print-bin"]:
            print(selected)
        else:
            require(len(sys.argv) > 1, "Usage: node_env.py --print-bin | COMMAND [ARGUMENTS]")
            environment = os.environ.copy()
            if selected:
                environment["PATH"] = selected + os.pathsep + environment["PATH"]
            os.execvpe(sys.argv[1], sys.argv[1:], environment)
    except (RuntimeError, OSError, subprocess.CalledProcessError) as error:
        detail = error.output if isinstance(error, subprocess.CalledProcessError) else str(error)
        print(f"Node environment: {detail}", file=sys.stderr)
        sys.exit(1)
