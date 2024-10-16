{
  description = "Simple tool to set your hyprpaper, written in Go";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = { self, nixpkgs }: 
  let
    system = "x86_64-linux";
    pkgs = import nixpkgs {
      inherit system;
      config = {}; 
      overlays = [
        (self: super: {
          hypsi = super.callPackage ./derivation.nix {};
        })
      ];
    };
  in
  {
    defaultPackage.${system} = pkgs.hypsi;
    
    devShells.${system}.default = let
      pkgs = import nixpkgs {
        inherit system;
      };
    in pkgs.mkShell {
        packages = with pkgs;[
          imagemagick
          gotools
          gopls
          go-outline
          go_1_22
          gopkgs
          gocode-gomod
          godef
          golint
        ];
        shellHook = ''
          ${pkgs.go_1_22}/bin/go version;
        '';
      };
  };
}

