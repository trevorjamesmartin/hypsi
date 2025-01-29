{
  lib,
  buildGoModule,
  pkg-config,
  pcre2,
  webkitgtk,
  gtk3,
  wrapGAppsHook3,
  gobject-introspection,
  libheif,
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

    # vendorHash = "sha256-+ZmXuL1mo6tcdvJ1dRifga9FkY6B6q7il9ZqqQLbrGg=";
    vendorHash = "sha256-MwLksZIT227HMIRr/MTyGbSrHSF0Ko5bX+anbWycHHE=";
    strictDeps = true;
    nativeBuildInputs = [pkg-config];
    buildInputs = [
      pcre2
      webkitgtk
      gtk3
      wrapGAppsHook3
      gobject-introspection
      libheif
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
