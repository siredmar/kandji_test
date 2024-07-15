{
  description = "gxctl";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        buildInputs = with pkgs; [ docker go goreleaser gnumake git ];
        version = "v0.38.0";
        # git rev is not available on dirty working trees but we still want to be able to build
        revFlag = pkgs.lib.strings.optionalString (self ? rev)
          "-X 'github.com/grid-x/gxctl/internal/version.GitCommit=${self.rev}'";
      in {
        packages = rec {
          gxctl = pkgs.buildGoModule {
            pname = "gxctl";
            inherit buildInputs version;
            src = ./.;
            CGO_ENABLED = 0;
            vendorHash = null;
            ldflags = [
              "-w"
              "-s"
              # We want deterministic builds that are the same independent of when we build this
              "-X 'github.com/grid-x/gxctl/internal/version.BuildTime=1970-01-01T00:00:00Z'"
              "-X 'github.com/grid-x/gxctl/internal/version.Version=${version}'"
              revFlag
            ];
          };
          default = gxctl;
        };
        devShells = {
          default =
            pkgs.mkShell { buildInputs = buildInputs ++ [ pkgs.nixfmt ]; };
        };
      });
}
