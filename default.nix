{ buildGoModule, lib, fetchFromGitHub, ... }:

buildGoModule rec {
  pname = "hyprpaperplanes";
  version = "0.3";

  src = ./.;
  vendorHash = null;

  meta = with lib; {
    description = "Simple tool to set your hyprpaper, written in Go";
    mainProgram = "hyprPaperPlanes";
    license = licenses.mit;
  };
}
