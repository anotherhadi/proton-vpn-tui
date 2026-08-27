{
  pkgs,
  gitHooksLib,
}: let
  hooks = gitHooksLib.run {
    src = ../.;
    hooks = {
      gofmt.enable = true;
      govet.enable = true;

      doctoc = {
        enable = true;
        name = "doctoc";
        entry = "doctoc --notitle README.md";
        language = "system";
        files = "(README\\.md|cmd/)";
        pass_filenames = false;
      };
    };
  };
in
  pkgs.mkShell {
    packages = with pkgs;
      [
        go
        gosec
        govulncheck
        doctoc
      ]
      ++ hooks.enabledPackages;

    shellHook = hooks.shellHook;
  }
