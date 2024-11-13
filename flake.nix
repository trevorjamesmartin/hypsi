{
  description = "Simple tool to set your hyprpaper, written in Go";

  inputs.nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";

  outputs = {
    self,
    nixpkgs,
  }: let
    system = "x86_64-linux";
    pkgs = nixpkgs.legacyPackages.${system};
  in {
    packages.${system}.default = pkgs.callPackage ./nix/package.nix {};
    devShells.${system}.default = pkgs.callPackage ./nix/shell.nix {};
  };
}
