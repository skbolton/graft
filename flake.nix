{
  description = "graft: multi-repo git worktree workspaces";

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];
      perSystem =
        {
          config,
          self',
          inputs',
          pkgs,
          system,
          ...
        }:
        let
          version = "0.0.0-dev";
        in
        {
          packages.graft = pkgs.buildGoModule {
            pname = "graft";
            inherit version;
            src = ./.;
            vendorHash = "sha256-/hdb5+3IMAUunZOLPd9MsSwu7N9q5rfgpC7oc9g2hkE=";

            nativeBuildInputs = [ pkgs.git ];

            ldflags = [
              "-s"
              "-w"
              "-X main.version=${version}"
            ];

            meta = {
              description = "Multi-repo git worktree workspaces";
              mainProgram = "graft";
            };
          };

          packages.default = self'.packages.graft;

          apps.check = {
            type = "app";
            program = "${pkgs.writeShellScriptBin "graft-check" ''
              set -euo pipefail
              export PATH="${pkgs.nix}/bin:${pkgs.go}/bin:$PATH"
              go vet ./...
              go test ./...
              nix build .#graft
            ''}/bin/graft-check";
          };

          devShells.default = pkgs.mkShell {
            packages = [
              pkgs.go
              pkgs.openspec
            ];
          };
        };
      flake = {
      };
    };
}
