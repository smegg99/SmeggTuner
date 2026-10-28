#!/usr/bin/env bash
#
# Removes what install.sh wrote; use the same XDG_DATA_HOME as the install.
# Keeps ~/.config/smeggtuner and $XDG_DATA_HOME/smeggtuner (settings, sessions and instruments).
set -euo pipefail

# Wrapped in main so the whole script is read before the installed copy removes itself.
main() {
    [ "$(id -u)" -ne 0 ] || fail "run this as the user who installed SmeggTuner, not as root"
    case "${HOME:-}" in
        /*) ;;
        *) fail "HOME is not an absolute path" ;;
    esac
    local data_home=${XDG_DATA_HOME:-}
    case "$data_home" in
        /*) ;;
        *) data_home="$HOME/.local/share" ;;
    esac
    local config_home=${XDG_CONFIG_HOME:-}
    case "$config_home" in
        /*) ;;
        *) config_home="$HOME/.config" ;;
    esac

    local lib_dir="$HOME/.local/lib/smeggtuner"
    local program="$lib_dir/smeggtuner"
    local link="$HOME/.local/bin/smeggtuner"

    removed=0
    if [ -L "$link" ]; then
        if [ "$(readlink -- "$link")" = "$program" ]; then
            remove "$link"
        else
            echo "Note: kept $link, which links to $(readlink -- "$link"), not to this install"
        fi
    elif [ -e "$link" ]; then
        echo "Note: kept $link, which this package did not make"
    fi
    remove "$data_home/applications/smeggtuner.desktop"
    remove "$data_home/icons/hicolor/scalable/apps/smeggtuner.svg"
    remove "$program"
    remove "$lib_dir/VERSION"
    remove "$lib_dir/uninstall.sh"
    if [ -d "$lib_dir" ] && ! rmdir -- "$lib_dir" 2>/dev/null; then
        echo "Note: kept $lib_dir, which holds files this package did not install"
    fi

    if [ "$removed" -eq 0 ]; then
        echo "OK: SmeggTuner is not installed for $(id -un); nothing to remove"
    else
        echo "OK: SmeggTuner removed"
    fi
    local kept
    for kept in "$config_home/smeggtuner" "$data_home/smeggtuner"; do
        [ -e "$kept" ] && echo "OK: kept $kept"
    done
    return 0
}

fail() {
    echo "ERR: $*" >&2
    exit 1
}

remove() {
    if [ -e "$1" ] || [ -L "$1" ]; then
        rm -f -- "$1"
        echo "OK: removed $1"
        removed=1
    fi
}

main "$@"
