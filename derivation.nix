{ pkgs, buildGoModule, lib, ... }:

buildGoModule {
  pname = "hypsi";
  version = "0.9.9";

  src = ./.;
  vendorHash = "sha256-mI9vu+UBOTGdPZkpovXg1ogDnBJSUv8HDsCyd+UVlH8=";
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
