{
  description = "Account - authentication & session management service";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-23.11";
  outputs = { self, nixpkgs, ... }@inputs:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      formatter.${system} = pkgs.nixpkgs-fmt;
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

        ACCOUNT_MYSQL_USER = "root";
        ACCOUNT_MYSQL_PASSWORD = "toor";
        ACCOUNT_MYSQL_DATABASE = "account";
      };
    };
}
