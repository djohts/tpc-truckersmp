# Changelog

## v2.7.0

- Update SII Decrypt
- Update dependencies
- Now built using Go 1.24
- A few other minor changes

## v2.6.3

- Fixed app crashing when no config file is present

## v2.6.2

- Fixed config loading

## v2.6.1

forgot to update version constant

## v2.6.0

- Added toggleable refueling for current vehicle
- Made trailer attachment toggleable
- Made teleporting toggleable
- Improved config handling
- Updated dependencies

## v2.5.2

_no changes_

## v2.5.1

_no changes_

## v2.5.0

- progress bar improvements by @djohts
- feat: embedded decrypt by @djohts in https://github.com/djohts/tpc-truckersmp/pull/10
- feat: delete old exe file after update by @djohts in https://github.com/djohts/tpc-truckersmp/pull/11
- chore(updater): verify checksum before applying update by @djohts in https://github.com/djohts/tpc-truckersmp/pull/12
- chore: update dependencies by @djohts in https://github.com/djohts/tpc-truckersmp/pull/15
- feat: toggle trailer attachment by @djohts in https://github.com/djohts/tpc-truckersmp/pull/16

## v2.4.0

- Added automatic download of `SII_Decrypt.exe`
- Added a progressbar for file downloads

## v2.3.0

> [!WARNING]
> Starting with this release, the app will be shipped as two `.exe` files (`SII_Decrypt.exe` and `tpc.exe` respectively).
> `SII_Decrypt.exe` is required for the app to work properly. Make sure it is in the same directory as `tpc.exe` and that it is not renamed.
>
> For now you have to download `SII_Decrypt.exe` manually, but the next version will automatically download it for you.

- Automatic updater (I really hope it works)

## v2.2.0

- Update checker
- More structural changes

## v2.1.0

- Improved logging
- Code cleanup and optimizations
- Updated dependencies

## v2.0.1

- Updated dependencies

## v2.0.0

- Permanent wear will now be removed as well
- Optimized executable size
- Added a config file
- Added auto mode (can be enabled in the config file):
  - Requires `keybinds.quicksave` to be set in the config file
  - Automatically sends the quicksave bind via [kayos/sendkeys](https://git.tcp.direct/kayos/sendkeys) after pressing `Alt+F12`
- Updated dependencies

## unknown (from upstream repo)

- Remove refueling
- Auto attach the trailer when it's detached
