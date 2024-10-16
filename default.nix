{ pkgs, buildGoModule, lib, fetchFromGitHub, ... }:

buildGoModule rec {
  pname = "hypsi";
  version = "0.9.7";

  src = ./.;
  #vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
  vendorHash = "sha256-uBpnnb7C2KOraXl/axTpKD1v7voEMSNCUGDdAHTzE8g=";
  nativeBuildInputs = with pkgs; [
    git
    pkg-config
    imagemagick
  ];


  buildInputs = with pkgs; [
    pcre2.dev
    webkitgtk.dev
    gtk3.dev
  ];
  meta = with lib; {
    description = "Simple tool to set your hyprpaper, written in Go";
    mainProgram = "hypsi";
    license = licenses.bsd3;
  };
}
