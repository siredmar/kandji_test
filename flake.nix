{
  description = "gxctl";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {inherit system;};
      buildInputs = with pkgs; [docker go goreleaser gnumake git];
      version = "v0.38.0";
      # git rev is not available on dirty working trees but we still want to be able to build
      revFlag =
        pkgs.lib.strings.optionalString (self ? rev)
        "-X 'github.com/grid-x/gxctl/internal/version.GitCommit=${self.rev}'";
      # this represents the suggested default config and the modified version we need for gxssh
      defaultConfigJSON = builtins.fromJSON (builtins.readFile ./pkg/action/ssh_config.json);
      modifiedConfigJSON =
        pkgs.lib.attrsets.updateManyAttrsByPath [
          {
            path = ["*.gridbox-tunnel" "ProxyCommand"];
            update = old: "${self.packages.${system}.gxctl}/bin/gxctl ssh tunnel --skip-config-check --profile='*' $(echo %h | cut -d'.' -f1)";
          }
          {
            path = ["*.gridbox" "ProxyCommand"];
            update = old: "gx" + old;
          }
        ]
        defaultConfigJSON;
      # predicate to convert from json to SSH config
      toSSHConf = attrOfAttrs: let
        # map function to string for each key val
        mapAttrsToStringsSep = mapFn: attrs:
          pkgs.lib.concatStringsSep "\n"
          (pkgs.lib.mapAttrsToList mapFn attrs);
        # prepend our headers (i.e. attr values) with "Host", then print space separated key value pairs from the lists inside the attr value
        mkSection = sectName: sectValues:
          ''
            Host ${sectName}
          ''
          + pkgs.lib.generators.toKeyValue {mkKeyValue = pkgs.lib.generators.mkKeyValueDefault {} " ";} sectValues;
      in
        mapAttrsToStringsSep mkSection attrOfAttrs;
      gxssh-config = pkgs.writeText "gxssh-config" (toSSHConf modifiedConfigJSON);
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
        gxssh = pkgs.symlinkJoin {
          name = "gxssh";
          buildInputs = [pkgs.makeWrapper];
          paths = [];
          postBuild = ''
            mkdir -p $out/bin
            cp ${pkgs.openssh}/bin/ssh $out/bin/gxssh
            wrapProgram $out/bin/gxssh --add-flags "-F ${gxssh-config}"
          '';
        };
        gxscp = pkgs.symlinkJoin {
          name = "gxscp";
          buildInputs = [pkgs.makeWrapper];
          paths = [];
          postBuild = ''
            mkdir -p $out/bin
            cp ${pkgs.openssh}/bin/scp $out/bin/gxscp
            wrapProgram $out/bin/gxscp --add-flags "-F ${gxssh-config}"
          '';
        };
        default = gxctl;
      };
      devShells = {
        default =
          pkgs.mkShell {buildInputs = buildInputs ++ [pkgs.nixfmt];};
      };
    });
}
