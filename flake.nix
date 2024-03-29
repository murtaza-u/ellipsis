{
  description = "Ellipsis - authentication & session management service";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-23.11";
  outputs = { self, nixpkgs, ... }@inputs:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
      version = "24.03";
    in
    {
      formatter.${system} = pkgs.nixpkgs-fmt;
      packages.${system} = rec {
        ellipsis = pkgs.buildGoModule {
          pname = "ellipsis";
          version = version;
          src = ./.;
          vendorHash = "sha256-HieSQxFYcfzp+WCv4JxgpTRf+A/HbGK6qAy7hDgnAZs=";
          CGO_ENABLED = 1;
          subPackages = [ "cmd/ellipsis" ];
        };
        dockerImage = pkgs.dockerTools.buildImage {
          name = "murtazau/ellipsis";
          tag = version;
          config = {
            Cmd = [ "${ellipsis}/bin/ellipsis" ];
            WorkingDir = "/data";
          };
        };
        default = ellipsis;
      };
      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
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
    };
}
