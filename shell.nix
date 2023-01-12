{ pkgs ? import <nixpkgs> { }, }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    docker
    go
    goreleaser
    gnumake
    git
  ];
}
