{
  mkShell,
  pkg-config,
  gotools,
  gopls,
  go-outline,
  go_1_22,
  gopkgs,
  gocode-gomod,
  godef,
  golint,
  pcre2,
  webkitgtk,
  gtk3,
  glib-networking,
  gsettings-desktop-schemas,
  libheif,
  impl,
  ...
}:
mkShell {
  name = "hypsi";
  nativeBuildInputs = [pkg-config];
  packages = [
    gotools
    gopls
    go-outline
    go_1_22
    gopkgs
    gocode-gomod
    godef
    golint
    pcre2
    webkitgtk
    gtk3
    glib-networking # tls/ssl
    gsettings-desktop-schemas # viewport, fonts
    libheif
    impl # generates method stubs for implementing an interface.
  ];
   
  shellHook = let
    # GTK app environment settings necessary for normal font rendering.
    #     note: at build time, this is covered by `wrapGApps` (see derivation.nix)
    unwrappedGApp = ''
      export XDG_DATA_DIRS=${gsettings-desktop-schemas}/share/gsettings-schemas/${gsettings-desktop-schemas.name}:${gtk3}/share/gsettings-schemas/${gtk3.name}:$XDG_DATA_DIRS;
      export GIO_MODULE_DIR="${glib-networking}/lib/gio/modules/";
    '';
  in ''
    ${unwrappedGApp}
    export SHELL=zsh;
    export PS1="\[\e[01;36m\][dev🐚Go]\[\e[0m\] \[\e[01;37m\]\w\[\e[0m\] $ ";
    export DEBUG=OK
    ${go_1_22}/bin/go version;
  '';
}
