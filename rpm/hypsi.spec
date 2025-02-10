Name:       hypsi
Version:    1.0.4
Release:    2%{?dist}
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

%files
%{_bindir}/%{name}
%{_datadir}/applications/hypsi.desktop

%changelog
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

