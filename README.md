# Beat Saber Patcher

patches Beat Saber maps whose BPM is off because of the last update.

## Why?

In the latest update to Beat Saber, a bunch of custom maps broke. This was because there had been a change in how the maps were loaded--instead of loading the BPM of each song from the track data itself, the BPM was instead loaded from the track's `info.json`. This broke a bunch of maps which had different values in those two places, and which were now either too slow or too fast. It would start fine, and then would get progressively worse as time went on, with the blocks sometimes totally out of sync with the rhythm.

## What it does

Beatsaber-patcher copies the correct BPM (from the track itself) into each song's `info.json`. This makes custom tracks behave the same way as they did before the update.

## How to use it

First, make sure all the custom songs you have installed are unzipped--that is, they're folders, not `.zip` files.

On Windows, download [`beatsaber_patcher_windows_amd64.exe`](https://github.com/wgoodall01/beatsaber-patcher/releases/latest) from the project's releases, drop it in your `steamapps\common\Beat Saber\CustomSongs` folder, and run it. You should see a window come up, some text scroll through it, and then see it close. That's good--it means all your maps were patched correctly.

If you want to actually read the text, or if it doesn't work, run it from `cmd.exe`. There are also some command-line options there if that's your thing.

Have fun!
