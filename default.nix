{ buildGoModule, lib, fetchFromGitHub, ... }:

buildGoModule rec {
  pname = "hypsi";
  version = "0.9.4";

  src = ./.;
  vendorHash = null;

  meta = with lib; {
    description = "Simple tool to set your hyprpaper, written in Go";
    mainProgram = "hypsi";
    license = licenses.bsd3;
  };
}
