{ pkgs, buildGoModule, lib, ... }:

buildGoModule rec {
  pname = "hypsi";
  version = "0.9.9";

  src = ./.;
  vendorHash = "sha256-K1pfhBTKF12oPPHvHJQ5DYxyF2w4YRIVC3dSYw0PbiA=";
  nativeBuildInputs = with pkgs; [
    git
    pkg-config
    imagemagick
  ];


  buildInputs = with pkgs; [
    pcre2
    webkitgtk
    gtk3
  ];
  meta = with lib; {
    description = "Simple tool to set your hyprpaper, written in Go";
    mainProgram = "hypsi";
    license = licenses.bsd3;
  };
}
