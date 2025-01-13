{
  lib,
  buildGoModule,
  pkg-config,
  pcre2,
  webkitgtk,
  gtk3,
  wrapGAppsHook3,
  gobject-introspection,
}: let
  pname = "hypsi";
  version = "1.2";
in
  buildGoModule {
    inherit pname version;
    # Reproducible source path for when the Git repository
    # is checked out under different names.
    src = builtins.path {
      path = ../.;
      name = "${pname}-${version}";
    };

    vendorHash = "sha256-gtH+RGM2CCKQO1Bz5M9twW+gIWCAmZ4gLDCq0KcQ7/I=";

    strictDeps = true;
    nativeBuildInputs = [pkg-config];
    buildInputs = [
      pcre2
      webkitgtk
      gtk3
      wrapGAppsHook3
      gobject-introspection
    ];

    # f around and find out
    preFixup = ''
      for f in $(find $out/bin/ -type f -executable); do
        wrapGApp $f
      done
    '';

    meta = {
      description = "Simple tool to set your hyprpaper, written in Go";
      mainProgram = "hypsi";
      license = lib.licenses.bsd3;
    };
  }
