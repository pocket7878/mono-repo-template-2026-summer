{
  description = "Reproducible development environment for the reference Todo architecture";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";

  outputs = { self, nixpkgs }:
    let
      systems = [ "aarch64-darwin" "x86_64-darwin" "aarch64-linux" "x86_64-linux" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in {
      devShells = forAllSystems (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            config.allowUnfreePredicate = pkg: builtins.elem (nixpkgs.lib.getName pkg) [
              "terraform"
            ];
          };
        in {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              nodejs_20
              terraform
              tflint
              gnumake
              git
              curl
              jq
              postgresql_17
              golangci-lint
              sqlc
              go-migrate
            ];

            shellHook = ''
              export PATH="$PWD/backend/bin:$PATH"
              echo "Reference Todo development shell: $(go version), $(node --version)"
              echo "Run 'make setup' once, then 'make gen' and 'make check'."
            '';
          };
        });
    };
}
