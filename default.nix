{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/cb9a96f23c491c081b38eab96d22fa958043c9fa.tar.gz") {config = { allowUnfree = true; }; } }:

pkgs.callPackage ./derivation.nix {}
