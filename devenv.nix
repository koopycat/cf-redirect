{ pkgs, ... }:

{
  dotenv.disableHint = true;

  # Build only with the nix Go; a go.mod that needs a newer one fails loudly
  # instead of silently downloading a toolchain.
  env.GOTOOLCHAIN = "local";

  packages = with pkgs; [
    git
    go_1_26
    just
  ];

  scripts.check.exec = "just check";
  scripts.build.exec = "just build";
  scripts.test.exec = "just test";
  scripts.race.exec = "just race";
  scripts.fmt.exec = "just fmt";
}
