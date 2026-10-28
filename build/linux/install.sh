#!/usr/bin/env bash
#
# Per-user install without root: ./install.sh from the unpacked package. Rerun to upgrade.
# Writes ~/.local/lib/smeggtuner/, ~/.local/bin/smeggtuner, and the menu entry and icon under $XDG_DATA_HOME.
set -euo pipefail

# Staged files not yet renamed into place, removed if the run stops early.
pending=()
trap 'rm -f -- "${pending[@]}"' EXIT

main() {
    local here
    here=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)

    [ "$(id -u)" -ne 0 ] || fail "run this as the user who will use SmeggTuner, not as root"
    case "${HOME:-}" in
        /*) ;;
        *) fail "HOME is not an absolute path" ;;
    esac
    local data_home=${XDG_DATA_HOME:-}
    case "$data_home" in
        /*) ;;
        *) data_home="$HOME/.local/share" ;;
    esac

    local lib_dir="$HOME/.local/lib/smeggtuner"
    local program="$lib_dir/smeggtuner"
    local link="$HOME/.local/bin/smeggtuner"
    local entry="$data_home/applications/smeggtuner.desktop"
    local icon="$data_home/icons/hicolor/scalable/apps/smeggtuner.svg"

    # Everything is checked before anything is written.
    local file
    for file in VERSION smeggtuner smeggtuner.desktop smeggtuner.svg uninstall.sh; do
        [ -f "$here/$file" ] || fail "the package has no $file; unpack it again"
    done
    local version
    version=$(tr -d '[:space:]' <"$here/VERSION")

    # These characters have special meaning in a .desktop file's quoted path.
    case "$program" in
        *'"'* | *'`'* | *'$'* | *'\'* | *'%'* | *$'\n'*)
            fail "the path $program has a character a menu entry cannot quote" ;;
    esac

    command -v ldd >/dev/null 2>&1 || fail "missing command: ldd"
    local libraries
    if ! libraries=$(ldd "$here/smeggtuner" 2>&1); then
        fail "the program does not run on this machine ($(uname -m)): $libraries"
    fi
    local missing
    missing=$(printf '%s\n' "$libraries" | sed -n 's/^[[:space:]]*\([^[:space:]]*\) => not found$/\1/p')
    if [ -n "$missing" ]; then
        echo "ERR: these libraries are missing:" >&2
        printf '%s\n' "$missing" | sed 's/^/  /' >&2
        echo "ERR: install GTK 3 and WebKitGTK 4.1 (Arch: gtk3 webkit2gtk-4.1; Debian and Ubuntu: libgtk-3-0 libwebkit2gtk-4.1-0; Fedora: gtk3 webkit2gtk4.1), then run install.sh again" >&2
        exit 1
    fi

    if [ -L "$link" ]; then
        [ "$(readlink -- "$link")" = "$program" ] ||
            fail "$link already links to $(readlink -- "$link"); remove it, then run install.sh again"
    elif [ -e "$link" ]; then
        fail "$link already exists and is not this package's link; remove it, then run install.sh again"
    fi

    local previous=''
    [ -f "$lib_dir/VERSION" ] && previous=$(tr -d '[:space:]' <"$lib_dir/VERSION")

    place "$here/smeggtuner" "$program" 0755
    place "$here/uninstall.sh" "$lib_dir/uninstall.sh" 0755
    place "$here/VERSION" "$lib_dir/VERSION" 0644
    place "$here/smeggtuner.svg" "$icon" 0644

    local staged line
    staged=$(stage "$entry")
    pending+=("$staged")
    while IFS= read -r line; do
        case "$line" in
            TryExec=smeggtuner) line="TryExec=$program" ;;
            Exec=smeggtuner) line="Exec=\"$program\"" ;;
        esac
        printf '%s\n' "$line"
    done <"$here/smeggtuner.desktop" >"$staged"
    settle "$staged" "$entry" 0644

    mkdir -p -- "$(dirname -- "$link")"
    [ -L "$link" ] || ln -s -- "$program" "$link"

    if [ -n "$previous" ] && [ "$previous" != "$version" ]; then
        echo "OK: SmeggTuner $previous replaced by $version"
    else
        echo "OK: SmeggTuner $version installed"
    fi
    echo "OK: start it from the application menu, or run $link"
    case ":${PATH:-}:" in
        *":$HOME/.local/bin:"*) ;;
        *) echo "Note: $HOME/.local/bin is not on PATH, so the command smeggtuner is not found in a terminal; the menu entry works." ;;
    esac
    echo "OK: remove it with $lib_dir/uninstall.sh"
}

fail() {
    echo "ERR: $*" >&2
    exit 1
}

# stage makes an empty file beside target so a rename replaces it atomically.
stage() {
    local dir
    dir=$(dirname -- "$1")
    mkdir -p -- "$dir"
    mktemp -- "$dir/.smeggtuner.XXXXXXXX"
}

settle() {
    chmod "$3" -- "$1"
    mv -f -- "$1" "$2"
}

place() {
    local staged
    staged=$(stage "$2")
    pending+=("$staged")
    cp -- "$1" "$staged"
    settle "$staged" "$2" "$3"
}

main "$@"
