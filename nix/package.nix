{
  lib,
  makeDesktopItem,
  buildGoModule,
  pkg-config,
  pcre2,
  webkitgtk_4_1,
  gtk3,
  wrapGAppsHook3,
  gobject-introspection,
  libheif,
}: let
  pname = "hypsi"; # program
  mname = "Hypsi"; # menu
  version = "1.0.4-2";
  desktopItem = makeDesktopItem {
      name = "${pname}";
      comment = "a simple hyprpaper management tool";
      exec = "${pname} -webview %f";
      icon = "hypsi";
      desktopName = "${mname}";
      genericName = "Hyprpaper Management";
      categories = [ "Graphics" "Utility" ];
      mimeTypes = ["image/gif" "image/heif" "image/jpeg" "image/png" "image/tiff" "image/x-webp" "image/webp" ];
  };
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
      mkdir -p "$out/share/applications"
    '';

    # TODO: create & install an icon file
    # install -Dm444 -T icon.png $out/share/hicolor/512/hypsi
    # mkdir -p "$out/share/icons/hicolor/128x128/apps";
    # cp $src/xdg/icon.png "$out/share/icons/hicolor/128x128/apps/hypsi"
    postInstall = ''
      mkdir -p "$out/share/applications";
      ln -s "${desktopItem}"/share/applications/* "$out/share/applications/";
    '';

    meta = {
      description = "Simple tool to set your hyprpaper, written in Go";
      mainProgram = "hypsi";
      license = lib.licenses.bsd3;
    };
  }
