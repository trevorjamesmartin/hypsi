{
  lib,
  buildGoModule,
  pkg-config,
  pcre2,
  webkitgtk_4_1,
  gtk3,
  wrapGAppsHook3,
  gobject-introspection,
  libheif,
}: let
  pname = "hypsi";
  version = "1.0.4";
in
  buildGoModule {
    inherit pname version;
    # Reproducible source path for when the Git repository
    # is checked out under different names.
    src = builtins.path {
      path = ../.;
      name = "${pname}-${version}";
    };

    vendorHash = "sha256-yJaBmFTiQ/b78XaMtKL652Xp2bDHJwViW2yboN/HV8U=";
    strictDeps = true;
    nativeBuildInputs = [pkg-config];
    buildInputs = [
      pcre2
      webkitgtk_4_1
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
