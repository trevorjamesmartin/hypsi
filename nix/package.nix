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
  mname = "Hypsi"; # menu entry / desktop file
  version = "1.0.5-1;
  desktopItem = makeDesktopItem {
      name = "${pname}";
      comment = "a simple hyprpaper management tool";
      exec = "${pname} -webview %f";
      icon = "hypsi";
      type = "Application";
      terminal = false;
      startupNotify = true;
      noDisplay = false;
      desktopName = "${mname}";
      genericName = "Hyprpaper Management";
      categories = [ "Graphics" "Utility" ];
      mimeTypes = [ "image/heif" "image/jpeg" "image/png" "image/tiff" "image/x-webp" "image/webp" ];
      actions = {
        OpenExistingFile = { name = "Open Existing File"; exec = "/usr/bin/hypsi %f"; };
        OpenNewWindow = { name = "Open New Window"; exec = "/usr/bin/hypsi -webview"; };
      };
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

    postInstall = ''
      mkdir -p "$out/share/applications";
      mkdir -p "$out/share/icons/hicolor/512x512/apps"
      ln -s "${desktopItem}"/share/applications/* "$out/share/applications/";
      install -Dm444 -T rpm/icon.png "$out/share/icons/hicolor/512x512/apps/hypsi.png"
    '';

    meta = {
      description = "Simple tool to set your hyprpaper, written in Go";
      mainProgram = "hypsi";
      license = lib.licenses.bsd3;
    };
  }
