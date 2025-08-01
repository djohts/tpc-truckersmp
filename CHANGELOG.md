# Changelog

## v2.10.1

- Attempt to fix the app patching the save before it's been fully written by the game

## v2.10.0

- add credits to readme
- split decrypt utils into a separate package
- replace decrypt with a more efficient and modern one
- cleanup and optimise watcher code

Overall the performance has been improved by ~60%, mostly thanks to the new decryptor.

## v2.9.4

- Fix for save decrypting not working in some cases

## v2.9.3

- Change temp file name to include tag name instead of branch name

## v2.9.2

- Updater will now keep previous app versions to allow rollback in case of issues

## v2.9.1

- Fixed a random crash
- Fixed the app triggering itself from decrypt writes

## v2.9.0

- Improved the updater, should be more reliable now

## v2.8.1

- Fixed a crash due to an incorrect execution call for `SII_Decrypt.exe`

## v2.8.0

- Fixed an app crash when using auto mode without a `cams.txt` file in either of the Document folders
- Changed internal handling of `SII_Decrypt.exe`. It will now be persistent in the same directory as the application and will not spam the temporary folder anymore.
- Updater improvements. Please report any issues by DMing me on Discord or by creating an issue.
- A few message updates.

## v2.7.3

- Added GitHub link to application startup header
- Updated dependencies
- Internal improvements

## v2.7.2

- Temporarily disable checksum verification upon updating

## v2.7.1

- Fix app version variable

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
