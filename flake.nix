{
  description = "Simple tool to set your hyprpaper, written in Go";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-compat.url = "github:edolstra/flake-compat";
    systems.url = "github:nix-systems/default";
  };

  outputs = {
    self,
    systems,
    nixpkgs,
    ...
  }: let
      eachSystem = nixpkgs.lib.genAttrs (import systems);
  in {
    packages = eachSystem (system: {
      default = nixpkgs.legacyPackages.${system}.callPackage ./nix/package.nix {};
    });

    devShells = eachSystem (system: {
      default = nixpkgs.legacyPackages.${system}.callPackage ./nix/shell.nix {};
    });

  };
}
