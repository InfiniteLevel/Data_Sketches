{
  description = "Golang Development Environment with Nix Flakes";

  inputs = {
    # The nixpkgs flake from the NixOS GitHub repo
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils"; # Utility for multi-platform support
  };

  outputs = {
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system: {
      # Access mkShell from the nixpkgs package set for the right system
      devShell = with nixpkgs.legacyPackages.${system};
        mkShell {
          buildInputs = [
            go
            gopls
            mockgen
            gotools
            go-junit-report
            gocover-cobertura
            go-task
            goperf
            pssh
            protobuf
            protoc-gen-go-grpc
            protoc-gen-go
            wireshark
            # python stuff for scripts
            (pkgs.python3.withPackages (python-pkgs:
              with python-pkgs; [
                matplotlib
                numpy
                dbus-python
                pandas
                # select Python packages here
              ]))
            basedpyright
          ];

          shellHook = ''
            export GOPATH=$HOME/go
            export GOROOT=$(which go | sed 's!/bin/go!!')/share/go
            export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
            export GOPROXY=https://proxy.golang.org,direct

            echo "Golang development environment setup"
          '';
        };
    });
}
