# Changelog

## 2.6.1 (2024-06-17)

forgot to update version constant

**Full Changelog**: https://github.com/djohts/tpc-truckersmp/compare/v2.6.0...v2.6.1

## 2.6.0 (2024-06-17)

- Added toggleable refueling for current vehicle
- Made trailer attachment toggleable
- Made teleporting toggleable
- Improved config handling
- Updated dependencies

**Full Changelog**: https://github.com/djohts/tpc-truckersmp/compare/v2.5.2...v2.6.0

## 2.5.2 (2024-06-01)

**Full Changelog**: https://github.com/djohts/tpc-truckersmp/compare/v2.5.1...v2.5.2

## 2.5.1 (2024-06-01)

**Full Changelog**: https://github.com/djohts/tpc-truckersmp/compare/v2.5.0...v2.5.1

## 2.5.0 (2024-06-01)

- progress bar improvements by @djohts
- feat: embedded decrypt by @djohts in https://github.com/djohts/tpc-truckersmp/pull/10
- feat: delete old exe file after update by @djohts in https://github.com/djohts/tpc-truckersmp/pull/11
- chore(updater): verify checksum before applying update by @djohts in https://github.com/djohts/tpc-truckersmp/pull/12
- chore: update dependencies by @djohts in https://github.com/djohts/tpc-truckersmp/pull/15
- feat: toggle trailer attachment by @djohts in https://github.com/djohts/tpc-truckersmp/pull/16

**Full Changelog**: https://github.com/djohts/tpc-truckersmp/compare/v2.4.0...v2.5.0

## 2.4.0 (2024-05-07)

- Added automatic download of `SII_Decrypt.exe`
- Added a progressbar for file downloads

**Full Changelog**: https://github.com/djohts/tpc-truckersmp/compare/v2.3.0...v2.4.0

## 2.3.0 (2024-05-07)

> [!WARNING]
> Starting with this release, the app will be shipped as two `.exe` files (`SII_Decrypt.exe` and `tpc.exe` respectively).
> `SII_Decrypt.exe` is required for the app to work properly. Make sure it is in the same directory as `tpc.exe` and that it is not renamed.
>
> For now you have to download `SII_Decrypt.exe` manually, but the next version will automatically download it for you.

- Automatic updater (I really hope it works)

**Full Changelog**: https://github.com/djohts/tpc-truckersmp/compare/v2.2.0...v2.3.0

## 2.2.0 (2024-05-06)

- Update checker
- More structural changes

**Full Changelog**: https://github.com/djohts/tpc-truckersmp/compare/v2.1.0...v2.2.0

## 2.1.0 (2024-05-06)

- Improved logging
- Code cleanup and optimizations
- Updated dependencies

**Full Changelog**: https://github.com/djohts/tpc-truckersmp/compare/v2.0.1...v2.1.0

## 2.0.1 (2024-05-01)

- Updated dependencies

**Full Changelog**: https://github.com/djohts/tpc-truckersmp/compare/v2.0.0...v2.0.1

## 2.0.0 (2024-04-30)

- Permanent wear will now be removed as well
- Optimized executable size
- Added a config file
- Added auto mode (can be enabled in the config file):
  - Requires `keybinds.quicksave` to be set in the config file
  - Automatically sends the quicksave bind via [kayos/sendkeys](https://git.tcp.direct/kayos/sendkeys) after pressing `Alt+F12`
- Updated dependencies

**Full Changelog**: https://github.com/djohts/tpc-truckersmp/compare/v0.2023.0122...v2.0.0

## unknown (from upstream repo)

- Remove refueling
- Auto attach the trailer when it's detached
