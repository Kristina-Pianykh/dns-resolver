{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }:
    # calling a function from `flake-utils` that takes a lambda
    # that takes the system we're targetting
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
      {
        formatter = pkgs.nixfmt-rfc-style;
        devShells.default = pkgs.mkShell {
          name = "dns";
          packages = with pkgs; [
            # unix coreutils
            gnumake
            ldns
            dig
            go
          ];
        };
      }
    );
}
