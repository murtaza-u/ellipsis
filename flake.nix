{
  description = "Ellipsis - authentication & session management service";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-23.11";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = { self, nixpkgs, flake-utils, ... }@inputs:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        version = "24.04.10";
      in
      {
        formatter = pkgs.nixpkgs-fmt;
        packages = rec {
          ellipsis = pkgs.buildGoModule {
            pname = "ellipsis";
            version = version;
            src = ./.;
            vendorHash = "sha256-c4Ti6Tve06tPNF5k8Yhl6fFVkGayijK5xvOs2/mxSes=";
            CGO_ENABLED = 1;
            subPackages = [ "cmd/ellipsis" ];
          };
          dockerImage = pkgs.dockerTools.buildImage {
            name = "murtazau/ellipsis";
            tag = version;
            copyToRoot = with pkgs.dockerTools; [
              caCertificates
            ];
            config = {
              Cmd = [ "${ellipsis}/bin/ellipsis" ];
              WorkingDir = "/data";
            };
          };
          default = ellipsis;
        };
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            nixd
            nixpkgs-fmt
            go
            go-tools
            gopls
            sqlc
            awscli2
            mycli
            air
            nodejs
            nodePackages.pnpm
            nodePackages.vscode-langservers-extracted
            nodePackages.typescript-language-server
            tailwindcss-language-server
            prettierd
          ];

          ELLIPSIS_CONFIG = "./config.yaml";
        };
      });
}
