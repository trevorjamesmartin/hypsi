{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/c0c77416acb81f698a292bae7bf3c1d524d421d9.tar.gz") {config = { allowUnfree = true; }; } }:

pkgs.callPackage ./derivation.nix {}
