Name:       hypsi
Version:    1.0.1
Release:    2%{?dist}
Summary:    A hyprpaper management tool

License:    BSD-3-Clause
Source0:    %{name}-%{version}.tar.gz

BuildRequires: golang >= 1.22
BuildRequires: libheif-devel >= 1.16
BuildRequires: webkit2gtk4.0-devel >= 2.45
BuildRequires: git
BuildRequires: gcc-c++

Requires: libheif-devel >= 1.16
Requires: webkit2gtk4.0-devel

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

%files
%{_bindir}/%{name}

%changelog
* Thu Jan 23 2025 Trevor Martin - 1.0.1
- First release%changelog

