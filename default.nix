{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/dbd8326b7fa64379bdf3983c2fbde9029352c72f.tar.gz") {config = { allowUnfree = true; }; } }:

pkgs.callPackage ./derivation.nix {}
