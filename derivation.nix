{ pkgs, buildGoModule, lib, ... }:

buildGoModule {
  pname = "hypsi";
  version = "1.0";

  src = ./.;
  vendorHash = "sha256-mI9vu+UBOTGdPZkpovXg1ogDnBJSUv8HDsCyd+UVlH8=";
  nativeBuildInputs = with pkgs; [
    git
    pkg-config
    imagemagick
    wrapGAppsHook3
    gobject-introspection
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

  # f around and find out
  preFixup = ''
    for f in $(find $out/bin/ -type f -executable); do
      wrapGApp $f
    done
  '';
}
