{
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

  outputs = { nixpkgs, self, ... }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [ go ];
        hardeningDisable = [ "fortify" ];
      };

      packages.${system}.default = pkgs.buildGoModule {
        pname = "protrans";
        version = "1.0";
        hardeningDisable = [ "fortify" ];

        src = ./.;

        vendorHash = "sha256-H79018dCud68fYT0l3IGZXQvD22byhnw/GchsiYJc68=";
      };

      overlays.${system} = _: _: { protrans = self.packages.${system}.default; };

      homeManagerModules.default = import ./home-manager.nix self;
    };
}
