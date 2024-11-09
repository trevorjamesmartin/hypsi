{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/2ef132b3585e21aa4dc50f0e0bb9955c1e0d7505.tar.gz") {config = { allowUnfree = true; }; } }:

pkgs.callPackage ./derivation.nix {}
