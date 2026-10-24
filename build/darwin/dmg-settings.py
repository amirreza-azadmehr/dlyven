# dmgbuild settings for the macOS disk image: the classic window with Dlyven on the left,
# an arrow, and Applications on the right. Used by scripts/package-macos.sh:
#
#   dmgbuild -s build/darwin/dmg-settings.py -D app=build/bin/Dlyven.app Dlyven out.dmg
#
# dmgbuild writes the window layout (.DS_Store) itself, so this needs no Finder or GUI
# session and works on CI runners. Paths are relative to the repository root.
import os.path

app = defines["app"]  # noqa: F821 (provided by dmgbuild -D)
name = os.path.basename(app)

format = "UDZO"
filesystem = "HFS+"
files = [app]
symlinks = {"Applications": "/Applications"}
hide_extensions = [name]  # "Dlyven", not "Dlyven.app"
icon = "build/darwin/iconfile.icns"  # volume icon on the desktop and in Finder's sidebar

# Picks up dmg-background@2x.png next to it for Retina screens.
background = "build/darwin/dmg-background.png"
window_rect = ((200, 160), (640, 400))
default_view = "icon-view"
icon_size = 128
text_size = 13
icon_locations = {name: (160, 190), "Applications": (480, 190)}
