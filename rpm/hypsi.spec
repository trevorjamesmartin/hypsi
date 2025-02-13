Name:       hypsi
Version:    1.0.4
Release:    6%{?dist}
Summary:    A hyprpaper management tool

License:    BSD-3-Clause
Source0:    %{name}-%{version}.tar.gz
BuildRequires: golang >= 1.22
BuildRequires: libheif-devel >= 1.16
BuildRequires: webkit2gtk4.1-devel
BuildRequires: git
BuildRequires: gcc-c++

Requires: libheif-devel >= 1.16
Requires: webkit2gtk4.1-devel

Provides: %{name} = %{version}

%global debug_package %{nil}

%global disable_source_fetch 0

%description
A simple hyprpaper management tool with highly configurable GUI

%prep
%autosetup

%build
go build -v -o %{name}

%install
install -Dpm 0755 %{name} %{buildroot}%{_bindir}/%{name}
install -Dpm 0644 rpm/hypsi.desktop %{buildroot}%{_datadir}/applications/hypsi.desktop
install -Dpm 0644 rpm/icon.png %{buildroot}%{_datadir}/icons/hicolor/512x512/apps/hypsi.png

%files
%{_bindir}/%{name}
%{_datadir}/applications/hypsi.desktop
%{_datadir}/icons/hicolor/512x512/apps/hypsi.png

%changelog
* Wed Feb 12 2025 Trevor Martin - 1.0.4-6
- test release: center thumbnails
* Wed Feb 12 2025 Trevor Martin - 1.0.4-5
- test release: install 512x512 icon
* Tue Feb 11 2025 Trevor Martin - 1.0.4-4
- test release: modesetting update
* Mon Feb 10 2025 Trevor Martin - 1.0.4-3
- fixed include mode in config
* Mon Feb 10 2025 Trevor Martin - 1.0.4-2
- fixed open-with .desktop item
* Sun Feb 09 2025 Trevor Martin - 1.0.4-1
- launch webview from menu
* Tue Feb 04 2025 Trevor Martin - 1.0.4
- added webp decoder for thumbnails
* Sat Feb 01 2025 Trevor Martin - 1.0.3
- breaking; storage honors XDG folder spec
* Wed Jan 29 2025 Trevor Martin - 1.0.2
- Now building with webkit2gtk4.1
* Thu Jan 23 2025 Trevor Martin - 1.0.1
- First release%changelog

