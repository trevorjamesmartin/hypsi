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
    packages.${system}.default = pkgs.callPackage ./derivation.nix {};

    devShells.${system}.default = let
      # GTK app environment settings necessary for normal font rendering.
      #     note: at build time, this is covered by `wrapGApps` (see derivation.nix)
      unwrappedGApp = with pkgs; ''
        export XDG_DATA_DIRS=${gsettings-desktop-schemas}/share/gsettings-schemas/${gsettings-desktop-schemas.name}:${gtk3}/share/gsettings-schemas/${gtk3.name}:$XDG_DATA_DIRS;
        export GIO_MODULE_DIR="${glib-networking}/lib/gio/modules/";
      '';
    in
      pkgs.mkShell {
        packages = with pkgs; [
          imagemagick
          gotools
          gopls
          go-outline
          go_1_22
          gopkgs
          gocode-gomod
          godef
          golint
          pcre2
          webkitgtk
          gtk3

          imagemagick
          glib-networking # tls/ssl
          gsettings-desktop-schemas # viewport, fonts
        ];

        nativeBuildInputs = [pkgs.pkg-config];

        shellHook = ''
          ${unwrappedGApp}
          export SHELL=zsh;
          export PS1="\[\e[01;36m\][dev🐚Go]\[\e[0m\] \[\e[01;37m\]\w\[\e[0m\] $ ";
          ${pkgs.go_1_22}/bin/go version;
        '';
      };
  };
}
