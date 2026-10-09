{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { nixpkgs, ... }:
    let
      systems = [
        "aarch64-darwin"
        "aarch64-linux"
        "x86_64-darwin"
        "x86_64-linux"
      ];
      forAllSystems =
        f:
        nixpkgs.lib.genAttrs systems (
          system:
          f {
            inherit system;
            pkgs = import nixpkgs { inherit system; };
          }
        );
    in
    {
      formatter = forAllSystems ({ pkgs }: pkgs.alejandra);
      devShells = forAllSystems (
        { system, pkgs }: {
          default = pkgs.mkShell {
            nativeBuildInputs = with pkgs; [
              git
              direnv
              gnumake
              binutils
              upx

              docker
              docker-compose

              go
              gopls

              hurl
              bruno
            ];
          };
        }
      );
    };
}
