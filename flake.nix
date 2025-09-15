{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    utils = { url = "github:numtide/flake-utils"; };
  };
  outputs = { self, nixpkgs, utils }:
    utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs { inherit system; };
          licenser = pkgs.buildGoModule rec {
            pname = "licenser";
            version = "0.8.0";
            src = pkgs.fetchFromGitHub {
              owner = "liamawhite";
              repo = "licenser";
              rev = "v${version}";
              sha256 = "sha256-r/3qCD/3bW548I3y1v403eqwVml1VCIr/va43gLmVc4=";
            };
            vendorHash = "sha256-LBVVhg69VdQVsVARCkwooe6N6DacgViIW/iQWHCya4k=";
          };
        in
        with pkgs;
        {
          devShells.default = pkgs.mkShell {
            buildInputs = [
              buf
              go
              golangci-lint
              goreleaser
              gosec
              protobuf
              git
              nodejs
              nodePackages.npm
              licenser
            ] ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
              # Playwright system dependencies (Linux only)
              glib
              nspr
              nss
              dbus
              atk
              expat
              at-spi2-atk
              xorg.libX11
              xorg.libXcomposite
              xorg.libXdamage
              xorg.libXext
              xorg.libXfixes
              xorg.libXrandr
              mesa # provides libgbm
              xorg.libxcb
              libxkbcommon
              systemd # provides libudev
              alsa-lib
            ];
          };
        }
      );
}

