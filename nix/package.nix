{
  lib,
  buildGoModule,
  pkg-config,
  pcre2,
  webkitgtk,
  gtk3,
  imagemagick,
  wrapGAppsHook3,
  gobject-introspection,
}: let
  pname = "hypsi";
  version = "1.0";
in
  buildGoModule {
    inherit pname version;
    # Reproducible source path for when the Git repository
    # is checked out under different names.
    src = builtins.path {
      path = ../.;
      name = "${pname}-${version}";
    };

    vendorHash = "sha256-mI9vu+UBOTGdPZkpovXg1ogDnBJSUv8HDsCyd+UVlH8=";

    strictDeps = true;
    nativeBuildInputs = [pkg-config];
    buildInputs = [
      pcre2
      webkitgtk
      gtk3
      imagemagick
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
