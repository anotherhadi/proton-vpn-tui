{
  description = "ProtonVPN TUI: A minimal, TUI and keyboard friendly wrapper for proton-vpn-cli.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    git-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    self,
    nixpkgs,
    git-hooks,
  }: let
    supportedSystems = ["x86_64-linux" "aarch64-linux"];

    forAllSystems = f:
      nixpkgs.lib.genAttrs supportedSystems
      (system: f system (import nixpkgs {inherit system;}));
  in {
    devShells = forAllSystems (system: pkgs: {
      default = import ./nix/shell.nix {
        inherit pkgs;
        gitHooksLib = git-hooks.lib.${system};
      };
    });
  };
}
