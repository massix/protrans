{
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

  outputs = { nixpkgs, self, ... }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
      hardeningDisable = [ "fortify" ];
      version = "1.1";
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        inherit hardeningDisable;
        packages = with pkgs; [ go ];
      };

      packages.${system}.default = pkgs.buildGoModule {
        inherit hardeningDisable version;
        pname = "protrans";

        src =
          let
            noSrcs = [ ".vscode" ".git" ".github" ".gitignore" ".envrc" ];
          in
          builtins.filterSource (path: _: ! builtins.elem (baseNameOf path) noSrcs) ./.;

        ldflags = [ "-X 'main.Version=${version}'" ];
        vendorHash = "sha256-H79018dCud68fYT0l3IGZXQvD22byhnw/GchsiYJc68=";
      };

      overlays.${system} = _: _: { protrans = self.packages.${system}.default; };

      homeManagerModules.default = import ./home-manager.nix self;
    };
}
